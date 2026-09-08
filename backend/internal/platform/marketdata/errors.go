package marketdata

import (
	"errors"
	"fmt"
)

// Classification sentinels. A provider failure is not just noise to log: it
// decides whether the user's stored key gets marked invalid (so we stop
// retrying it every morning), merely rate-limited (retry tomorrow), or whether
// the symbol was simply not covered.
var (
	// ErrUnauthorized means the provider rejected the key itself.
	ErrUnauthorized = errors.New("marketdata: the provider rejected the API key")
	// ErrRateLimited means the key is fine but its quota is exhausted.
	ErrRateLimited = errors.New("marketdata: provider rate limit reached")
	// ErrThrottled means the provider refused this one call for pacing reasons
	// and said nothing lasting about the key: it is asking for the next request
	// later, not reporting a budget that is gone.
	//
	// It is separate from ErrRateLimited because the two have different
	// lifetimes, and only one of them is worth writing down. Alpha Vantage
	// answers a burst with "please consider spreading out your free API requests
	// more sparingly (1 request per second)"; the same key succeeds seconds
	// later. Recording that as an exhausted quota leaves a standing "no quota"
	// badge on a key that works, which is exactly what it used to do.
	//
	// The throttle is per **IP** as much as per key — a call with a different
	// key from the same host trips it too — so our own concurrency can cause it,
	// and no verdict about the user's credential follows from it.
	ErrThrottled = errors.New("marketdata: provider is throttling requests right now")
	// ErrUnsupported means this provider has no data for the symbol or pair,
	// which is the signal for the fallback chain to try the next one.
	ErrUnsupported = errors.New("marketdata: provider does not cover this symbol")
	// ErrNoCredentials means the caller has no usable key, so no provider chain
	// could be built.
	ErrNoCredentials = errors.New("marketdata: no credential configured")
)

// providerError is the only error type the clients return. Its message is
// scrubbed of the API key at construction, and the sentinel it wraps is a
// static string, so neither Error() nor any errors.Is/As traversal can surface
// the key.
//
// It also records which provider failed. That matters because a fallback chain
// joins the failures of several of the user's keys: without attribution, one
// provider rejecting its key would look like grounds to invalidate all of them.
type providerError struct {
	provider ProviderName
	msg      string
	sentinel error
}

func (e *providerError) Error() string { return e.msg }

// Unwrap exposes the sentinel for errors.Is without exposing the original
// transport error, whose message would still hold the key-bearing URL.
func (e *providerError) Unwrap() error { return e.sentinel }

// Errorf builds a provider error whose message cannot contain apiKey. sentinel
// may be nil when the failure needs no classification.
//
// It deliberately does not wrap the underlying error with %w: Go's transport
// errors quote the full request URL, and providers take the key as a query
// parameter, so keeping the original in the chain would keep the key one
// Error() call away.
func Errorf(provider ProviderName, apiKey string, sentinel error, format string, args ...any) error {
	return new(providerError{
		provider: provider,
		msg:      scrub(fmt.Sprintf(format, args...), apiKey),
		sentinel: sentinel,
	})
}

// Verdict is what one provider said about one call.
type Verdict struct {
	Provider ProviderName
	// Err is the sentinel (ErrUnauthorized, ErrRateLimited, ErrUnsupported) or
	// nil when the failure was not classified.
	Err error
	// Message is the scrubbed text, safe to persist and to log.
	Message string
}

// Verdicts attributes the failures inside err to the providers that produced
// them, walking the tree errors.Join builds inside a fallback chain.
//
// Callers use it to act on the right key: if Finnhub is throttled but Alpha
// Vantage rejected its key, only the second one should be marked invalid.
func Verdicts(err error) []Verdict {
	var out []Verdict

	var walk func(error)
	walk = func(e error) {
		if e == nil {
			return
		}

		// A direct assertion, not errors.As: As searches the whole subtree and
		// returns the first match, which would stop the walk at the first
		// provider and leave every sibling in a joined error unattributed.
		if pe, ok := e.(*providerError); ok && pe.provider != "" {
			out = append(out, Verdict{Provider: pe.provider, Err: pe.sentinel, Message: pe.msg})
			// A providerError is a leaf: its Unwrap yields only the sentinel.
			return
		}

		switch unwrapped := e.(type) {
		case interface{ Unwrap() []error }:
			for _, inner := range unwrapped.Unwrap() {
				walk(inner)
			}
		case interface{ Unwrap() error }:
			walk(unwrapped.Unwrap())
		}
	}

	walk(err)

	return out
}
