package memo

import (
	"sync"
)

// Memo is a concurrent-safe, singleflight cache with optional Loader.
//
// Two-level locking strategy:
//   - Memo.Mutex (global lock): protects the cache map (insert/delete).
//     Held only briefly during map lookups, never during Loader execution.
//   - item.Mutex (per-key lock): protects value/err/expireAt/loaded/loading.
//     The first goroutine to load a key holds this lock during Loader execution;
//     subsequent goroutines for the same key block here until the load completes.
//
// Lock ordering: global lock (m.Mutex) → item lock (item.Mutex).
// Never acquire global lock while holding an item lock.
type Memo struct {
	sync.Mutex
	Options
	cache map[Key]*item
}

// item represents a cached entry for a single key.
//
// State transitions (all under item.Mutex):
//
//	(created) → loading=true → loaded=true, loading=false  (success or CacheError=true)
//	              or
//	           → loaded=false, loading=false, item deleted  (error with CacheError=false)
type item struct {
	sync.Mutex
	value    Value
	err      error
	expireAt int64
	loaded   bool // true after setCache completes
	loading  bool // true while Loader is in progress (singleflight barrier)
}

func NewMemo(opts ...Option) *Memo {
	m := &Memo{
		Options: newOptions(opts...),
		cache:   make(map[Key]*item),
	}
	return m
}

// setCache writes the Loader result into the item and marks it as loaded.
func (i *item) setCache(o Options, value Value, err error) {
	i.value, i.err = value, err
	if o.Expiration == NoExpire {
		i.expireAt = 0
	} else {
		i.expireAt = o.Clock.Now().Add(o.Expiration).UnixNano()
	}
	i.loaded = true
	i.loading = false
}

// isValid returns true if the item has been loaded and has not expired.
func (i *item) isValid(now int64) bool {
	return i.loaded && (i.expireAt == 0 || i.expireAt > now)
}

func (m *Memo) Get(key Key, opts ...Option) (value Value, err error) {
	o := m.newGetOptions(opts...)

	m.Lock()
	i := m.cache[key]

	if o.Loader == nil {
		// --- No Loader: pure cache read ---
		if i == nil {
			m.Unlock()
		} else {
			m.Unlock()
			i.Lock()
			defer i.Unlock()
			if i.isValid(m.Clock.Now().UnixNano()) {
				return i.value, i.err
			}
		}
		return nil, ErrNotFound
	}

	// --- Loader is provided ---
	// Ensure item exists in the cache map. Create if missing.
	if i == nil {
		i = &item{}
		m.cache[key] = i
	}
	// At this point i is in the cache. Lock the item before releasing global lock
	// so that we can safely read i.loading under item lock.
	m.Unlock()
	i.Lock()
	defer i.Unlock()

	if i.loading {
		// [singleflight follower] Another goroutine is currently loading this key.
		// We just acquired i.Mutex, which means the leader has finished.
		// The leader's setCache has already updated i.loaded/i.value/i.err.
		if i.isValid(m.Clock.Now().UnixNano()) {
			return i.value, i.err
		}
		// Leader finished but cache is invalid (expired or error-not-cached).
		// Fall through to become the new loader.
	} else if i.isValid(m.Clock.Now().UnixNano()) {
		// [cache hit] Item is loaded and not expired — return directly.
		return i.value, i.err
	}
	// Item is not valid (expired, error-not-cached, or just created with loaded=false).
	// Become the singleflight leader: set loading=true so any new goroutines
	// that arrive will see loading=true and wait for us.
	i.loading = true

	value, err = o.Loader(key)
	if err != nil && !o.CacheError {
		// CacheError=false: error is not cached.
		// Remove item from cache so next Get retries Loader.
		// Lock ordering: we hold i.Mutex, then acquire m.Mutex. This is safe
		// because the reverse order never occurs (m.Mutex is never held while
		// waiting on i.Mutex since we released m.Mutex above).
		i.loading = false
		i.loaded = false
		m.Lock()
		delete(m.cache, key)
		m.Unlock()
		return value, err
	}

	i.setCache(o, value, err)
	return value, err
}

func (m *Memo) Set(key Key, value Value, opts ...Option) {
	o := m.newSetOptions(opts...)
	m.Lock()
	i := m.cache[key]
	if i == nil {
		i = &item{}
		m.cache[key] = i
	}
	i.value, i.err = value, nil
	if o.Expiration == NoExpire {
		i.expireAt = 0
	} else {
		i.expireAt = m.Clock.Now().Add(o.Expiration).UnixNano()
	}
	i.loaded = true
	m.Unlock()
}

func (m *Memo) Invalid(key Key) {
	m.Lock()
	delete(m.cache, key)
	m.Unlock()
}
