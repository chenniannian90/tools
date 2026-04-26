package memo

import (
	"errors"
	"github.com/chenniannian90/tools/memo/clock"
	"time"
)

const (
	DoExpire Expiration = -1
	NoExpire Expiration = 0
)

var ErrNotFound = errors.New("memo: not found")

type (
	Key        = interface{}
	Value      = interface{}
	Clock      = clock.Clock
	Loader     func(Key) (Value, error)
	Expiration = time.Duration
)

// Options configures Memo behavior. Set via WithXxx() Option functions.
type Options struct {
	Clock      Clock      // time source; defaults to real system clock
	Loader     Loader     // function to load a missing/expired value
	Expiration Expiration // TTL for cached entries; NoExpire (0) means never expire
	// CacheError controls whether Loader errors are cached.
	//   - true (default): errors are cached with the same Expiration as successes.
	//     Concurrent goroutines sharing a singleflight will all receive the same error.
	//   - false: on error, the cache entry is removed so the next Get retries Loader.
	//     Concurrent goroutines already waiting in singleflight still receive the error,
	//     but a subsequent serial Get will trigger a fresh Loader call.
	CacheError bool
}

type Option func(*Options)

// newOptions creates Options with defaults: real clock, CacheError=true.
func newOptions(opts ...Option) Options {
	o := Options{
		Clock:      clock.NewRealClock(),
		CacheError: true,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// newGetOptions inherits all fields from base Options, then applies per-call overrides.
func (base *Options) newGetOptions(opts ...Option) Options {
	o := Options{
		Clock:      base.Clock,
		Loader:     base.Loader,
		Expiration: base.Expiration,
		CacheError: base.CacheError,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// newSetOptions inherits only Expiration from base Options (Loader/Clock not relevant for Set).
func (base *Options) newSetOptions(opts ...Option) Options {
	o := Options{
		Expiration: base.Expiration,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func WithClock(clock Clock) Option {
	return func(o *Options) {
		o.Clock = clock
	}
}

func WithLoader(loader Loader) Option {
	return func(o *Options) {
		o.Loader = loader
	}
}

func WithExpiration(expiration Expiration) Option {
	return func(o *Options) {
		o.Expiration = expiration
	}
}

// WithCacheError sets whether Loader errors should be cached.
// Default is true. Pass false to make Get retry Loader on every call after an error.
func WithCacheError(cache bool) Option {
	return func(o *Options) {
		o.CacheError = cache
	}
}
