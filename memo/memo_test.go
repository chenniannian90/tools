package memo

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chenniannian90/tools/memo/clock"
)

var errTest = errors.New("test error")

func TestGet_NoLoader_ReturnsErrNotFound(t *testing.T) {
	m := NewMemo()
	_, err := m.Get("key")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSet_ThenGet(t *testing.T) {
	m := NewMemo()
	m.Set("key", "value")

	val, err := m.Get("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected 'value', got %v", val)
	}
}

func TestGet_WithLoader(t *testing.T) {
	var calls int32
	m := NewMemo(WithLoader(func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		return "loaded:" + key.(string), nil
	}))

	val, err := m.Get("k1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "loaded:k1" {
		t.Fatalf("expected 'loaded:k1', got %v", val)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	// Second Get should use cache, not call loader again.
	val, err = m.Get("k1")
	if val != "loaded:k1" {
		t.Fatalf("expected cached value, got %v", val)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call (cached), got %d", calls)
	}
}

func TestGet_ExpiredEntry(t *testing.T) {
	fc := clock.NewFakeClock()
	m := NewMemo(
		WithClock(fc),
		WithLoader(func(key Key) (Value, error) {
			return "value", nil
		}),
	)

	m.Get("key")
	fc.Advance(time.Hour)

	// Expired entry should trigger reload.
	val, err := m.Get("key", WithExpiration(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected 'value', got %v", val)
	}
}

func TestSet_WithExpiration(t *testing.T) {
	fc := clock.NewFakeClock()
	m := NewMemo(WithClock(fc))

	m.Set("key", "value", WithExpiration(time.Minute))
	val, _ := m.Get("key")
	if val != "value" {
		t.Fatalf("expected 'value', got %v", val)
	}

	fc.Advance(time.Minute + 1)
	_, err := m.Get("key")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after expiration, got %v", err)
	}
}

func TestInvalid(t *testing.T) {
	m := NewMemo(WithLoader(func(key Key) (Value, error) {
		return "loaded", nil
	}))

	m.Get("key")
	m.Invalid("key")

	_, err := m.Get("key", WithLoader(nil))
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after Invalid, got %v", err)
	}
}

func TestLoaderError_Cached(t *testing.T) {
	fc := clock.NewFakeClock()
	var calls int32
	loaderErr := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		return nil, errTest
	}

	m := NewMemo(
		WithClock(fc),
		WithLoader(loaderErr),
		WithExpiration(time.Hour),
	)

	_, err := m.Get("key")
	if err != errTest {
		t.Fatalf("expected test error, got %v", err)
	}

	// Default CacheError=true — error IS cached.
	_, err = m.Get("key")
	if err != errTest {
		t.Fatalf("expected cached test error, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 loader call (cached error), got %d", calls)
	}

	// After expiration, Loader should be called again.
	fc.Advance(time.Hour + 1)
	_, err = m.Get("key")
	if err != errTest {
		t.Fatalf("expected test error after reload, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 loader calls (after expiration), got %d", calls)
	}
}

func TestLoaderError_NotCached(t *testing.T) {
	var calls int32
	loaderErr := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		return nil, errTest
	}

	m := NewMemo(
		WithLoader(loaderErr),
		WithCacheError(false),
	)

	_, err := m.Get("key")
	if err != errTest {
		t.Fatalf("expected test error, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 loader call, got %d", calls)
	}

	// CacheError=false — error NOT cached, next call retries.
	_, err = m.Get("key")
	if err != errTest {
		t.Fatalf("expected test error on retry, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 loader calls (no cache), got %d", calls)
	}
}

func TestConcurrentGet_Singleflight(t *testing.T) {
	var calls int32
	slowLoader := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return "value", nil
	}

	m := NewMemo(WithLoader(slowLoader))

	var wg sync.WaitGroup
	const n = 20
	results := make([]Value, n)
	errors := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errors[idx] = m.Get("key")
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error %v", i, err)
		}
		if results[i] != "value" {
			t.Fatalf("goroutine %d: expected 'value', got %v", i, results[i])
		}
	}
	// singleflight should deduplicate — only 1 load call.
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 load call (singleflight), got %d", calls)
	}
}

// TestConcurrentGet_Singleflight_100Concurrent 验证同一 key 100 个并发请求，
// Loader 只执行 1 次，所有 goroutine 都拿到相同结果。
func TestConcurrentGet_Singleflight_100Concurrent(t *testing.T) {
	var calls int32
	slowLoader := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(100 * time.Millisecond) // 慢 Loader，放大并发窗口
		return "result:" + key.(string), nil
	}

	m := NewMemo(WithLoader(slowLoader))

	const n = 100
	var wg sync.WaitGroup
	results := make([]Value, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = m.Get("same-key")
		}(i)
	}
	wg.Wait()

	// Loader 只应被调用 1 次
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("singleflight failed: expected 1 loader call, got %d", got)
	}

	// 所有 goroutine 都应拿到正确结果
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, err)
		}
		if results[i] != "result:same-key" {
			t.Fatalf("goroutine %d: expected 'result:same-key', got %v", i, results[i])
		}
	}
}

// TestConcurrentGet_Singleflight_MultipleKeys 验证不同 key 的并发请求互不影响，
// 每个 key 的 Loader 各自只执行 1 次。
func TestConcurrentGet_Singleflight_MultipleKeys(t *testing.T) {
	var calls int32
	slowLoader := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return "value:" + key.(string), nil
	}

	m := NewMemo(WithLoader(slowLoader))

	keys := []string{"a", "b", "c"}
	const perKey = 30
	var wg sync.WaitGroup
	results := make(map[string][]Value)
	errs := make(map[string][]error)
	for _, k := range keys {
		results[k] = make([]Value, perKey)
		errs[k] = make([]error, perKey)
	}

	for _, k := range keys {
		for i := 0; i < perKey; i++ {
			wg.Add(1)
			go func(key string, idx int) {
				defer wg.Done()
				results[key][idx], errs[key][idx] = m.Get(key)
			}(k, i)
		}
	}
	wg.Wait()

	// 3 个 key，每个 Loader 只调 1 次，总共 3 次
	if got := atomic.LoadInt32(&calls); got != int32(len(keys)) {
		t.Fatalf("expected %d loader calls (1 per key), got %d", len(keys), got)
	}

	for _, k := range keys {
		expected := "value:" + k
		for i, err := range errs[k] {
			if err != nil {
				t.Fatalf("key=%s goroutine %d: unexpected error: %v", k, i, err)
			}
			if results[k][i] != expected {
				t.Fatalf("key=%s goroutine %d: expected '%s', got %v", k, i, expected, results[k][i])
			}
		}
	}
}

// TestConcurrentGet_Singleflight_ExpiredReload 验证缓存过期后并发请求只重新加载 1 次。
func TestConcurrentGet_Singleflight_ExpiredReload(t *testing.T) {
	fc := clock.NewFakeClock()
	var calls int32
	slowLoader := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return "loaded", nil
	}

	m := NewMemo(
		WithClock(fc),
		WithLoader(slowLoader),
		WithExpiration(time.Minute),
	)

	// 第一次加载
	m.Get("key")
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("initial load: expected 1 call, got %d", got)
	}

	// 推进时间使缓存过期
	fc.Advance(time.Minute + 1)

	// 50 个并发请求同时触发，应只重新加载 1 次
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := m.Get("key")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if val != "loaded" {
				t.Errorf("expected 'loaded', got %v", val)
			}
		}()
	}
	wg.Wait()

	// 1 (初始) + 1 (过期重载) = 2
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expired reload: expected 2 total calls, got %d", got)
	}
}

// TestConcurrentGet_Singleflight_LoaderError 验证 Loader 返回错误时，
// 所有并发 goroutine 都拿到同一个错误，且 Loader 只调 1 次。
func TestConcurrentGet_Singleflight_LoaderError(t *testing.T) {
	var calls int32
	slowFailLoader := func(key Key) (Value, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return nil, errTest
	}

	m := NewMemo(WithLoader(slowFailLoader))

	const n = 50
	var wg sync.WaitGroup
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = m.Get("key")
		}(i)
	}
	wg.Wait()

	// Loader 只调 1 次
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 loader call, got %d", got)
	}

	// 所有 goroutine 都应拿到同一个错误
	for i, err := range errs {
		if err != errTest {
			t.Fatalf("goroutine %d: expected errTest, got %v", i, err)
		}
	}
}

// TODO: uncomment after implementing WithCleanupInterval and Close
// func TestClose_Idempotent(t *testing.T) {
// 	m := NewMemo(WithCleanupInterval(time.Millisecond))
// 	m.Close()
// 	m.Close()
// 	m.Close()
// }
//
// func TestClose_ConcurrentSafe(t *testing.T) {
// 	m := NewMemo(WithCleanupInterval(time.Millisecond))
// 	var wg sync.WaitGroup
// 	for i := 0; i < 100; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			m.Close()
// 		}()
// 	}
// 	wg.Wait()
// }
//
// func TestCleanup(t *testing.T) {
// 	fc := clock.NewFakeClock()
// 	m := NewMemo(
// 		WithClock(fc),
// 		WithCleanupInterval(100*time.Millisecond),
// 	)
//
// 	m.Set("a", "1", WithExpiration(time.Minute))
// 	m.Set("b", "2", WithExpiration(time.Minute))
//
// 	fc.Advance(time.Minute + 1)
// 	time.Sleep(200 * time.Millisecond) // wait for cleanup tick
//
// 	_, err := m.Get("a")
// 	if err != ErrNotFound {
// 		t.Fatalf("expected ErrNotFound for expired key 'a', got %v", err)
// 	}
// 	_, err = m.Get("b")
// 	if err != ErrNotFound {
// 		t.Fatalf("expected ErrNotFound for expired key 'b', got %v", err)
// 	}
//
// 	m.Close()
// }
