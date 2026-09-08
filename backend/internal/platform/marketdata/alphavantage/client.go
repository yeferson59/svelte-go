// Package alphavantage talks to the Alpha Vantage API with a key the user
// brought themselves.
//
// Alpha Vantage only accepts the key as a URL query parameter, so the key is
// present in every request URL. That makes error handling security-relevant:
// Go's transport errors quote the URL they failed on, so no error from this
// package may be returned verbatim. Everything goes through
// marketdata.Errorf, which scrubs the key out of the message.
package alphavantage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	json "github.com/bytedance/sonic"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const baseURL = "https://www.alphavantage.co/query"

var _ marketdata.Provider = (*Client)(nil)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New builds a client for one user's key. Callers should pass the shared
// marketdata.DefaultHTTPClient; a nil client falls back to it rather than
// minting a new one per call.
func New(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = marketdata.DefaultHTTPClient
	}

	return new(Client{apiKey: apiKey, httpClient: httpClient})
}

// get issues a query and decodes it into out. The response envelope is
// inspected first: Alpha Vantage reports a bad key and an exhausted quota with
// HTTP 200 and a JSON field, not a status code.
func (c *Client) get(ctx context.Context, params url.Values, what string, out any) error {
	params.Set("apikey", c.apiKey)

	endpoint := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: build request %s: %v", what, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: http get %s: %v", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnauthorized, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrRateLimited, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: %s: status %d", what, resp.StatusCode)
	}

	// Decode into a buffer first so the envelope can be classified before the
	// caller's own decoding.
	var raw json.NoCopyRawMessage
	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: decode %s: %v", what, err)
	}

	if err := c.classify(raw, what); err != nil {
		return err
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, nil, "alphavantage: decode %s: %v", what, err)
	}

	return nil
}

type clasifyEnvelope struct {
	ErrorMessage string `json:"Error Message"`
	Note         string `json:"Note"`
	Information  string `json:"Information"`
}

// classify turns Alpha Vantage's 200-with-a-message replies into the sentinels
// the credential store uses to decide whether a key is dead, spent or merely
// being asked to slow down.
//
// The field a message arrives in does not say what happened, so the field is
// not what decides. `Information` alone carries at least four unrelated
// answers — a spent daily quota, a per-second burst, a premium-only endpoint
// and "the demo key is for demo purposes only" — and the old code called all
// of them an exhausted quota. That is how a key that works a second later
// ended up wearing a standing "no quota" badge: the verdict was persisted
// against the credential.
func (c *Client) classify(raw json.NoCopyRawMessage, what string) error {
	var envelope clasifyEnvelope

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil
	}

	for _, msg := range []string{envelope.ErrorMessage, envelope.Note, envelope.Information} {
		if msg == "" {
			continue
		}

		return marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, sentinelFor(msg), "alphavantage: %s: %s", what, msg)
	}

	return nil
}

// sentinelFor reads what Alpha Vantage actually said.
//
// Order matters and is not alphabetical. The quota message names the key ("We
// have detected your API key as … and our standard API rate limit is 25
// requests per day") and the burst message names the premium plans, so a naive
// "mentions the key ⇒ bad key" or "mentions premium ⇒ premium endpoint" test
// would misread both. Each case is therefore matched on the phrase that only
// that answer uses, most specific first.
//
// An unrecognised message yields nil: no sentinel means "this says nothing we
// can act on", which is what keeps a message we have never seen from being
// written down as a verdict on the user's key.
func sentinelFor(msg string) error {
	lower := strings.ToLower(msg)

	switch {
	// Transient and shared: the same key succeeds when the calls are spaced
	// out, and a *different* key from the same host trips it too. Nothing about
	// the credential follows from it.
	case strings.Contains(lower, "spreading out") ||
		strings.Contains(lower, "request per second") ||
		strings.Contains(lower, "calls per minute"):
		return marketdata.ErrThrottled

	// The day's budget really is gone. This is the only one worth recording.
	case strings.Contains(lower, "requests per day") || strings.Contains(lower, "daily rate limit"):
		return marketdata.ErrRateLimited

	// The key is valid and its quota untouched; this plan just does not serve
	// this call. Same shape as a symbol the provider does not carry, so the
	// fallback chain should move on to the next key.
	case strings.Contains(lower, "premium endpoint"):
		return marketdata.ErrUnsupported

	// "The **demo** API key is for demo purposes only": a real statement about
	// the key, and the one case in this field that is.
	case strings.Contains(lower, "demo api key"):
		return marketdata.ErrUnauthorized

	// "the parameter apikey is invalid or missing". Reached only after the
	// quota case above, which also mentions the key.
	case strings.Contains(lower, "apikey") || strings.Contains(lower, "api key"):
		return marketdata.ErrUnauthorized

	// "Invalid API call. Please retry or visit the documentation…" is what a
	// symbol Alpha Vantage does not know produces. It is about the request, not
	// the credential — blaming the key here used to mark it invalid, and an
	// invalid key is dropped from every later sync.
	case strings.Contains(lower, "invalid api call"):
		return marketdata.ErrUnsupported

	default:
		return nil
	}
}

type envelopExchange struct {
	Data map[string]string `json:"Realtime Currency Exchange Rate"`
}

func (c *Client) FetchExchangeRate(ctx context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error) {
	what := fmt.Sprintf("%s/%s", from, to)

	var envelope envelopExchange

	params := url.Values{}

	params.Set("function", "CURRENCY_EXCHANGE_RATE")
	params.Set("from_currency", from.String())
	params.Set("to_currency", to.String())

	if err := c.get(ctx, params, what, &envelope); err != nil {
		return marketdata.ExchangeRateResult{}, err
	}

	rate, ok := envelope.Data["5. Exchange Rate"]
	if !ok || rate == "" {
		return marketdata.ExchangeRateResult{}, marketdata.Errorf(marketdata.AlphaVantage, c.apiKey, marketdata.ErrUnsupported, "alphavantage: no rate for %s", what)
	}

	return marketdata.ExchangeRateResult{Rate: rate, Source: marketdata.AlphaVantage, FetchedAt: time.Now().UTC()}, nil
}
