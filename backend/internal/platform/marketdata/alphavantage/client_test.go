package alphavantage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// roundTripFunc lets a test stub the HTTP transport, capturing the outgoing
// request and returning a canned response without touching the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(body string) *http.Response {
	return new(http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	})
}

// newTestClient builds a Client whose HTTP transport is replaced by fn.
func newTestClient(fn roundTripFunc) *Client {
	// A dedicated client per test: New's default is process-wide, and mutating
	// its transport here would leak the stub into every other caller.
	return New("test-key", new(http.Client{Transport: fn}))
}

func TestFetchQuote(t *testing.T) {
	t.Run("returns price from a valid response", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"Global Quote":{"01. symbol":"AAPL","05. price":"192.53"}}`), nil
		})

		res, err := c.FetchQuote(context.Background(), "AAPL")
		if err != nil {
			t.Fatalf("FetchQuote: %v", err)
		}
		if res.Price != "192.53" {
			t.Errorf("Price = %q, want 192.53", res.Price)
		}
		if res.FetchedAt.IsZero() {
			t.Error("FetchedAt should be set")
		}
		if !strings.Contains(gotURL, "function=GLOBAL_QUOTE") ||
			!strings.Contains(gotURL, "symbol=AAPL") ||
			!strings.Contains(gotURL, "apikey=test-key") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors when the price is missing", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"Global Quote":{}}`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected error for missing price")
		}
	})

	t.Run("errors on malformed JSON", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`not json`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestFetchExchangeRate(t *testing.T) {
	t.Run("returns rate from a valid response", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"Realtime Currency Exchange Rate":{"5. Exchange Rate":"1.0850"}}`), nil
		})

		res, err := c.FetchExchangeRate(context.Background(), money.EUR, money.USD)

		if err != nil {
			t.Fatalf("FetchExchangeRate: %v", err)
		}
		if res.Rate != "1.0850" {
			t.Errorf("Rate = %q, want 1.0850", res.Rate)
		}
		if res.FetchedAt.IsZero() {
			t.Error("FetchedAt should be set")
		}
		if !strings.Contains(gotURL, "function=CURRENCY_EXCHANGE_RATE") ||
			!strings.Contains(gotURL, "from_currency=EUR") ||
			!strings.Contains(gotURL, "to_currency=USD") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors when the rate is missing", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"Realtime Currency Exchange Rate":{"5. Exchange Rate":""}}`), nil
		})

		if _, err := c.FetchExchangeRate(context.Background(), money.EUR, money.USD); err == nil {
			t.Fatal("expected error for missing rate")
		}
	})
}

// The messages below are what alphavantage.co actually answered when this test
// was written, copied verbatim. They matter because Alpha Vantage reports every
// one of these with HTTP 200 and a JSON field, and reuses `Information` for
// answers that have nothing to do with each other: the field cannot classify
// them, only the prose can.
func TestClassifyReadsWhatTheProviderSaid(t *testing.T) {
	cases := []struct {
		name string
		body string
		want error
	}{
		{
			// Transient, and shared across keys because it follows our IP: the
			// same key succeeds when the calls are spaced a few seconds apart.
			// Calling this an exhausted quota is what left a working key with a
			// standing "no quota" badge, because the verdict was persisted.
			name: "a burst is a throttle, not a spent quota",
			body: `{"Information":"Thank you for using Alpha Vantage! Please consider spreading out your free API requests more sparingly (1 request per second). You may subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to lift the free key rate limit (25 requests per day), raise the per-second burst limit, and instantly unlock all premium endpoints"}`,
			want: marketdata.ErrThrottled,
		},
		{
			// The day really is over. The one message in this family worth
			// writing down against the key.
			name: "the daily budget being gone is a spent quota",
			body: `{"Information":"We have detected your API key as ABC123 and our standard API rate limit is 25 requests per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits."}`,
			want: marketdata.ErrRateLimited,
		},
		{
			// Note the trap: the quota message above also names the API key, so
			// a "mentions the key ⇒ bad key" test would misread it.
			name: "the demo key notice is about the key",
			body: `{"Information":"The **demo** API key is for demo purposes only. Please claim your free API key at (https://www.alphavantage.co/support/#api-key) to explore our full API offerings. It takes fewer than 20 seconds."}`,
			want: marketdata.ErrUnauthorized,
		},
		{
			// The key is valid and its quota untouched; this plan just does not
			// serve this call, so the chain should try the next key.
			name: "a premium-only endpoint is not the key's fault",
			body: `{"Information":"Thank you for using Alpha Vantage! This is a premium endpoint. You may subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly unlock all premium endpoints."}`,
			want: marketdata.ErrUnsupported,
		},
		{
			name: "a missing key is about the key",
			body: `{"Error Message":"the parameter apikey is invalid or missing. Please claim your free API key on (https://www.alphavantage.co/support/#api-key)."}`,
			want: marketdata.ErrUnauthorized,
		},
		{
			// This one used to mark the credential invalid, and an invalid
			// credential is dropped from every later sync — so one bad ticker
			// silently retired a working key.
			name: "an invalid call is about the request, not the credential",
			body: `{"Error Message":"Invalid API call. Please retry or visit the documentation (https://www.alphavantage.co/documentation/) for GLOBAL_QUOTE."}`,
			want: marketdata.ErrUnsupported,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(func(*http.Request) (*http.Response, error) {
				return jsonResponse(tc.body), nil
			})

			_, err := c.FetchQuote(context.Background(), "AAPL")
			if err == nil {
				t.Fatal("expected an error")
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want it to wrap %v", err, tc.want)
			}
		})
	}

	// A message nobody has catalogued must not become a verdict on the key. No
	// sentinel is what makes the callers say "we learned nothing" instead of
	// guessing, which is the whole reason the old catch-all was wrong.
	t.Run("an unrecognised message carries no verdict", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"Information":"Something Alpha Vantage has not said before."}`), nil
		})

		_, err := c.FetchQuote(context.Background(), "AAPL")
		if err == nil {
			t.Fatal("expected an error")
		}

		for _, sentinel := range []error{
			marketdata.ErrThrottled, marketdata.ErrRateLimited,
			marketdata.ErrUnauthorized, marketdata.ErrUnsupported,
		} {
			if errors.Is(err, sentinel) {
				t.Errorf("err wraps %v; an unknown message should classify as nothing", sentinel)
			}
		}
	})
}
