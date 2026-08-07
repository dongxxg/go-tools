
## 并发和原子操作

原子操作即是进行过程中不能被中断的操作。也就是说，针对某个值的原子操作在被进行的过程当中，CPU绝不会再去进行其它的针对该值的操作。无论这些其它的操作是否为原子操作都会是这样。为了实现这样的严谨性，原子操作仅会由一个独立的CPU指令代表和完成。只有这样才能够在并发环境下保证原子操作的绝对安全。

Go语言提供的原子操作都是非侵入式的。它们由标准库代码包sync/atomic中的众多函数代表。我们可以通过调用这些函数对几种简单的类型的值进行原子操作。这些类型包括int32、int64、uint32、uint64、uintptr和unsafe.Pointer类型，共6个。这些函数提供的原子操作共有5种，即：增或减、比较并交换、载入、存储和交换。

用原子操作来替换mutex锁其主要原因是，原子操作由底层硬件支持，而锁则由操作系统提供的API实现。若实现相同的功能，前者通常会更有效率。

syncx包中提供了AtomicDuration、AtomicBool、AtomicFloat64等，对并发操作进行封装，包括:

| name              | 作用                     |
| :---------------- | :----------------------- |
| AtomicBool        | bool类型 原子类          |
| AtomicDuration    | Duration有关 原子类      |
| AtomicFloat64     | float64类型 原子类       |
| Barrier           | 栏栅(包装加锁解锁)   |
| Cond              | 条件变量                 |
| DoneChan          | 优雅通知关闭             |
| ImmutableResource | 创建后不会修改的资源     |
| Limit             | 控制请求数               |
| LockedCalls       | 确保方法的串行调用       |
| ManagedResource   | 资源管理                 |
| Once              | 提供 `once func`         |
| OnceGuard         | 一次性使用的资源管理     |
| Pool              | pool，简单的池           |
| RefResource       | 引用计数的资源           |
| ResourceManager   | 资源管理器               |
| SharedCalls       | 类似 `singflight` 的功能 |
| SpinLock          | 自旋锁：自旋+CAS         |
| TimeoutLimit      | Limit + timeout 控制     |

## atomic

因为没有 *泛型* 支持，所以才会出现多种类型的原子类支持。以下采用 `float64` 作为例子：

```go
func (f *AtomicFloat64) Add(val float64) float64 {
 for {
  old := f.Load()
  nv := old + val
  if f.CompareAndSwap(old, nv) {
   return nv
  }
 }
}

func (f *AtomicFloat64) CompareAndSwap(old, val float64) bool {
 return atomic.CompareAndSwapUint64((*uint64)(f), math.Float64bits(old), math.Float64bits(val))
}

func (f *AtomicFloat64) Load() float64 {
 return math.Float64frombits(atomic.LoadUint64((*uint64)(f)))
}

func (f *AtomicFloat64) Set(val float64) {
 atomic.StoreUint64((*uint64)(f), math.Float64bits(val))
}
```

- `Add(val)`：如果 `CAS` 失败，不断for循环重试，获取 old val，并set old+val；
- `CompareAndSwap(old, new)`：调用底层 `atomic` 的 `CAS`；
- `Load()`：调用 `atomic.LoadUint64` ，然后转换
- `Set(val)`：调用 `atomic.StoreUint64`

至于其他类型，可以依照上述，调用原始 `atomic` 操作，然后转换为需要的类型，比如：遇到 `bool` 可以借助 `0, 1` 来分辨对应的 `false, true`。

## Barrier

这里 `Barrier` 只是将业务函数操作封装，作为闭包传入，内部将 `lock` 操作的加锁解锁自行解决了。

```go
func (b *Barrier) Guard(fn func()) {
 b.lock.Lock()
 defer b.lock.Unlock()
  // 自己的业务逻辑
 fn()
}
```

## Cond/Limit/TimeoutLimit

这个数据结构和 `Limit` 一起组成了 `TimeoutLimit`：

```go
func NewTimeoutLimit(n int) TimeoutLimit {
 return TimeoutLimit{
  limit: NewLimit(n),
  cond:  NewCond(),
 }
}

func NewLimit(n int) Limit {
 return Limit{
  pool: make(chan lang.PlaceholderType, n),
 }
}
```

- `limit` 这里是有缓冲的 `channel`；
- `cond` 是无缓冲的；

所以这里结合名字来理解：因为 `Limit` 是限制某一种资源的使用，所以需要预先在资源池中放入预置数量的资源；`Cond` 类似阀门，需要两边都准备好，才能进行数据交换，所以使用无缓冲，同步控制。

举例理解资源控制：

```go
func (cs *concurrentSession) takeSession(opts ...Option) (*mgo.Session, error) {
  // 选项参数注入
 ...
  // 看 limit 中是否还能取出资源
 if err := cs.limit.Borrow(o.timeout); err != nil {
  return nil, err
 } else {
  return cs.Copy(), nil
 }
}

func (l TimeoutLimit) Borrow(timeout time.Duration) error {
  // 1. 如果还有 limit 中还有资源，取出一个，返回
 if l.TryBorrow() {
  return nil
 }
 // 2. 如果 limit 中资源已经用完了
 var ok bool
 for {
    // 只有 cond 可以取出一个【无缓存，也只有 cond <- 此条才能通过】
  timeout, ok = l.cond.WaitWithTimeout(timeout)
    // 尝试取出一个【上面 cond 通过时，就有一个资源返回了】
    // 看 `Return()`
  if ok && l.TryBorrow() {
   return nil
  }
  // 超时控制
  if timeout <= 0 {
   return ErrTimeout
  }
 }
}

func (l TimeoutLimit) Return() error {
  // 返回去一个资源
 if err := l.limit.Return(); err != nil {
  return err
 }
 // 同步通知另一个需要资源的协程【实现了阀门，两方交换】
 l.cond.Signal()
 return nil
}
```

## 资源管理

ManagedResource &`ResourceManager`从名字上类似，这里将两个组件放在一起讲解。

先从结构上：

```go
type ManagedResource struct {
  // 资源
 resource interface{}
 lock     sync.RWMutex
  // 生成资源的逻辑，由开发者自己控制
 generate func() interface{}
  // 对比资源
 equals   func(a, b interface{}) bool
}

type ResourceManager struct {
  // 资源：这里看得出来是 I/O，
 resources   map[string]io.Closer
 sharedCalls SharedCalls
  // 对资源map互斥访问
 lock        sync.RWMutex
}
```

然后来看获取资源的方法签名：

```go
func (manager *ResourceManager) GetResource(key, create func() (io.Closer, error)) (io.Closer, error)

// 获取一个资源（有就直接获取，没有生成一个）
func (mr *ManagedResource) Take() interface{}
// 判断这个资源是否不符合传入的判断要求，不符合则重置
func (mr *ManagedResource) MarkBroken(resource interface{})
```

1. `ResourceManager` 使用 `SharedCalls` 做防重复请求，并将资源缓存在内部的 `sourMap`；另外传入的 `create func` 和 `IO` 操作有关，常见用在网络资源的缓存；
2. `ManagedResource` 缓存资源没有 `map` 而是单一的 `interface` ，说明只有一份，但是它提供了 `Take()` 和传入 `generate()`说明可以让开发者自行更新 `resource`；

所以在用途上：

- `ResourceManager`：用在网络资源的管理。如：数据库连接管理；
- `ManagedResource`：用在一些变化资源，可以做资源前后对比，达到更新资源。如：`token` 管理和验证

## RefResource

这个就和 `GC` 中引用计数类似：

- `Use() -> ref++`
- `Clean() -> ref--; if ref == 0 -> ref clean`

```go
func (r *RefResource) Use() error {
  // 互斥访问
 r.lock.Lock()
 defer r.lock.Unlock()
 // 清除标记
 if r.cleaned {
  return ErrUseOfCleaned
 }
 // 引用 +1
 r.ref++
 return nil
}
```

## SharedCalls

一句话形容：**使用SharedCalls可以使得同时多个请求只需要发起一次拿结果的调用，其他请求"坐享其成"，这种设计有效减少了资源服务的并发压力，可以有效防止缓存击穿**。

这个组件被反复应用在其他组件中，上面说的 `ResourceManager`。

类似当需要高频并发访问一个资源时，就可以使用 `SharedCalls` 缓存。

```go
// 当多个请求同时使用Do方法请求资源时
func (g *sharedGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
  // 先申请加锁
  g.lock.Lock()

  // 根据key，获取对应的call结果,并用变量c保存
  if c, ok := g.calls[key]; ok {
    // 拿到call以后，释放锁，此处call可能还没有实际数据，只是一个空的内存占位
    g.lock.Unlock()
    // 调用wg.Wait，判断是否有其他goroutine正在申请资源，如果阻塞，说明有其他goroutine正在获取资源
    c.wg.Wait()
    // 当wg.Wait不再阻塞，表示资源获取已经结束，可以直接返回结果
    return c.val, c.err
  }

  // 没有拿到结果，则调用makeCall方法去获取资源，注意此处仍然是锁住的，可以保证只有一个goroutine可以调用makecall
  c := g.makeCall(key, fn)
  // 返回调用结果
  return c.val, c.err
}
```
