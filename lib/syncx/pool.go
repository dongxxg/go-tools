package syncx

import (
	"sync"
	"time"

	"github.com/dongxxg/go-tools/lib/timex"
)

type (
	// PoolOption自定义pool的函数
	PoolOption func(*Pool)

	node struct {
		item     interface{}
		next     *node
		lastUsed time.Duration
	}

	// Pool is used to pool resources.
	// 跟sync.Pool的区别在于:
	//  1. 可以设置资源的limit
	//  2. 可以设置资源的maxAge
	//  3. 可以自定义资源的create和destroy函数
	Pool struct {
		limit   int
		created int
		maxAge  time.Duration
		lock    sync.Locker
		cond    *sync.Cond
		head    *node
		create  func() interface{}
		destroy func(interface{})
	}
)

// 1. pool 中的对象会根据使用时间做懒销毁；
// 2. 使用 cond 做对象消费和生产的通知以及阻塞；
// 3. 可以自定义自己的生产函数，销毁函数；
func NewPool(n int, create func() interface{}, destroy func(interface{}), opts ...PoolOption) *Pool {
	if n <= 0 {
		panic("pool size can't be negative or zero")
	}

	lock := new(sync.Mutex)
	pool := &Pool{
		limit:   n,
		lock:    lock,
		cond:    sync.NewCond(lock),
		create:  create,
		destroy: destroy,
	}

	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

// Get取出资源
func (p *Pool) Get() interface{} {
	// 调用 cond.Wait 时必须要持有c.L的锁
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		// 1. pool中对象池是一个用链表连接的nodelist
		if p.head != nil {
			head := p.head
			p.head = head.next
			// 1.1 如果当前节点：当前时间 >= 上次使用时间+对象最大存活时间
			if p.maxAge > 0 && head.lastUsed+p.maxAge < timex.Now() {
				p.created--
				// 说明当前节点已经过期了 -> 销毁节点对应的对象，然后继续寻找下一个节点
				// 【⚠️：不是销毁节点，而是销毁节点对应的对象】
				p.destroy(head.item)
				continue
			} else {
				return head.item
			}
		}

		// 2. 对象池是懒加载的，get的时候才去创建对象链表
		if p.created < p.limit {
			p.created++
			// 由开发者自己传入：生产函数
			return p.create()
		}

		p.cond.Wait()
	}
}

// Put放回资源
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}
	// 互斥访问 pool 中nodelist
	p.lock.Lock()
	defer p.lock.Unlock()

	p.head = &node{
		item:     x,
		next:     p.head,
		lastUsed: timex.Now(),
	}
	// 放入head，通知其他正在get的协程
	p.cond.Signal()
}

// WithMaxAge返回一个自定义maxAge的函数
func WithMaxAge(duration time.Duration) PoolOption {
	return func(pool *Pool) {
		pool.maxAge = duration
	}
}