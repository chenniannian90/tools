package memo

import (
	"sync"
)

// Memo 是并发安全的 singleflight 缓存，支持可选的 Loader。
//
// 两级锁策略:
//   - Memo.Mutex（全局锁）: 保护 cache map（增删）。仅在 map 操作时短暂持有，Loader 执行期间不持有。
//   - item.Mutex（per-key 锁）: 保护单个 key 的 value/err/expireAt/loaded/loading。
//     第一个加载该 key 的 goroutine 在 Loader 执行期间持有此锁；
//     后续相同 key 的 goroutine 在此阻塞直到加载完成。
//
// 锁顺序: 全局锁 (m.Mutex) → item 锁 (item.Mutex)。
// 绝不会在持有 item 锁时获取全局锁（除 CacheError=false 清理时，此时 item 锁已持有且不会等待全局锁）。
//
// # 原始代码存在的问题
//
// 1. expireAt==0 歧义导致缓存穿透:
//
//	新建的 item 所有字段为零值（expireAt=0），和"永不过期"的语义相同。
//	当 goroutine A 释放全局锁后、还未获取 item 锁时，goroutine B 可能先拿到 item 锁，
//	此时 expireAt==0 被误判为"永不过期的缓存命中"，返回零值 nil。
//	测试表现为偶发的 "expected 'value', got <nil>" 错误。
//
// 2. 缺少 loading 标记导致 singleflight 失效:
//
//	原始代码没有 loading 字段。goroutine 拿到 item 锁后无法区分:
//	  - "第一个 goroutine 正在加载中"（应等待）
//	  - "缓存已加载且有效"（应直接返回）
//	  - "缓存过期需要重新加载"（应成为新的 loader）
//	导致多个 goroutine 同时发现缓存过期后各自调 Loader，违背 singleflight 语义。
//
// 3. else 分支的过期重载无去重:
//
//	缓存过期时，N 个 goroutine 同时进入 else 分支，各自发现过期，
//	全部调用 Loader，造成 N 次重复加载。
//
// 修复方案:
//   - 新增 loaded 字段: 区分"还没加载"（loaded=false）和"加载完成且永不过期"（loaded=true, expireAt=0）
//   - 新增 loading 字段: singleflight 屏障，leader 设 loading=true，follower 看到后等待
//   - isValid 要求 loaded=true: 防止读取未加载完成的零值
type Memo struct {
	sync.Mutex
	Options
	cache map[Key]*item
}

// item 表示单个 key 的缓存条目。
//
// 状态流转（均在 item.Mutex 保护下）:
//
//	(创建) → loading=true → loaded=true, loading=false    （成功 或 CacheError=true）
//	            或
//	         → loaded=false, loading=false, item 被删除    （出错 且 CacheError=false）
type item struct {
	sync.Mutex
	value    Value
	err      error
	expireAt int64
	loaded   bool // setCache 完成后为 true；修复原始代码 expireAt==0 歧义问题
	loading  bool // Loader 执行期间为 true；singleflight 屏障，防止并发重复加载
}

func NewMemo(opts ...Option) *Memo {
	m := &Memo{
		Options: newOptions(opts...),
		cache:   make(map[Key]*item),
	}
	return m
}

// setCache 将 Loader 结果写入 item 并标记为已加载。
// 原始代码直接在 Get 中赋值，expireAt 的设置无法区分"未加载"和"永不过期"。
// 现在通过 loaded=true 明确标记加载已完成。
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

// isValid 判断 item 是否已加载且未过期。
// 原始代码只检查 expireAt: expireAt==0 || expireAt>now
// 这会将"新建未加载的 item"（expireAt 默认为 0）误判为有效缓存。
// 修复: 增加 loaded 条件，只有 setCache 完成后才视为有效。
func (i *item) isValid(now int64) bool {
	return i.loaded && (i.expireAt == 0 || i.expireAt > now)
}

func (m *Memo) Get(key Key, opts ...Option) (value Value, err error) {
	o := m.newGetOptions(opts...)

	m.Lock()
	i := m.cache[key]

	if o.Loader == nil {
		// --- 无 Loader: 纯缓存读取 ---
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

	// --- 有 Loader ---
	// 确保 item 存在于 cache map 中，不存在则创建。
	// 此操作在全局锁内完成，保证后续 goroutine 能通过 m.cache[key] 拿到同一个 item。
	if i == nil {
		i = &item{}
		m.cache[key] = i
	}
	// 先锁住 item 再释放全局锁，保证 i.loading 的读取在 item 锁保护下进行。
	// 原始代码的问题: 释放全局锁后再读 item 状态（expireAt 等），
	// 可能读到第一个 goroutine 还未写入的零值。
	m.Unlock()
	i.Lock()
	defer i.Unlock()

	if i.loading {
		// [singleflight 跟随者] 另一个 goroutine 正在加载该 key。
		// 能拿到 item 锁说明 leader 已经完成，setCache 已更新了值。
		// 原始代码没有 loading 标记，follower 无法判断 leader 是否在执行中，
		// 可能误读零值或重复调 Loader。
		if i.isValid(m.Clock.Now().UnixNano()) {
			return i.value, i.err
		}
		// leader 完成但缓存无效（过期 或 错误未缓存），继续往下成为新的 loader
	} else if i.isValid(m.Clock.Now().UnixNano()) {
		// [缓存命中] 已加载且未过期，直接返回
		return i.value, i.err
	}
	// 缓存无效（过期 / 错误未缓存 / 刚创建 loaded=false）。
	// 成为 singleflight leader: 设置 loading=true，后续新到达的 goroutine 会看到并等待。
	// 原始代码没有此屏障，过期时多个 goroutine 会同时调 Loader。
	i.loading = true

	value, err = o.Loader(key)
	if err != nil && !o.CacheError {
		// CacheError=false: 错误不缓存。
		// 重置状态并从 cache 删除，下次 Get 会重新调 Loader。
		// 锁顺序安全: 已持有 i.Mutex，再获取 m.Mutex。
		// 不会死锁因为不存在反向顺序（m.Mutex 不会在等待 i.Mutex 时持有）。
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
