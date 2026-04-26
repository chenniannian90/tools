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

// Options 配置 Memo 的行为，通过 WithXxx() Option 函数设置。
type Options struct {
	Clock      Clock      // 时钟源，默认为系统真实时钟
	Loader     Loader     // 加载函数，用于加载缺失或过期的值
	Expiration Expiration // 缓存 TTL；NoExpire (0) 表示永不过期
	// CacheError 控制 Loader 错误是否被缓存:
	//   - true（默认）: 错误和成功一样缓存，使用相同的 Expiration。
	//     并发 singleflight 中的 goroutine 都会收到同一个错误。
	//   - false: 出错时删除缓存条目，下次 Get 会重新调 Loader。
	//     已在 singleflight 中等待的 goroutine 仍会收到错误，
	//     但之后的串行 Get 会触发新的 Loader 调用。
	CacheError bool
}

type Option func(*Options)

// newOptions 创建默认 Options: 真实时钟，CacheError=true。
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

// newGetOptions 继承 base Options 的所有字段，再应用单次调用的覆盖项。
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

// newSetOptions 仅继承 Expiration（Set 不需要 Loader/Clock）。
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

// WithCacheError 设置 Loader 错误是否缓存。默认 true。传 false 使 Get 在出错后每次重试 Loader。
func WithCacheError(cache bool) Option {
	return func(o *Options) {
		o.CacheError = cache
	}
}
