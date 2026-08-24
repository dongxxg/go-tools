package collection

import "sync"

// The DefaultRingCapacity of an uninitialized Ring buffer.
const DefaultRingCapacity int = 10

// NewRing 新建一个容量为capacity的环
func NewRing(capacity int) *Ring {
	if capacity < 0 {
		return nil
	}

	r := &Ring{}
	r.SetCapacity(capacity)
	return r
}

// Ring实现了环形缓存结构，类似于FIFO队列，默认capacity是10。
// 从head入，从tail出。
// goroutine并发安全
type Ring struct {
	sync.Mutex
	head int // 最后加入的元素，初始值为-1
	tail int // 最早加入的元素，初始值为0
	buff []interface{}
}

// SetCapacity 设置环的最大容量
func (r *Ring) SetCapacity(size int) {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	r.extend(size)
}

// Capacity 返回环的容量.
func (r *Ring) Capacity() int {
	r.Lock()
	defer r.Unlock()

	return r.capacity()
}

// ContentSize 返回环中实际的对象数量
func (r *Ring) ContentSize() int {
	r.Lock()
	defer r.Unlock()

	if r.head == -1 {
		return 0
	}

	difference := (r.head - r.tail)
	if difference < 0 {
		difference += r.capacity()
	}
	return difference + 1
}

// Enqueue 向环中添加一个数据，放在head.
// 如果原来位置已有数据，会覆盖并继续写入
func (r *Ring) Enqueue(i interface{}) {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	r.set(r.head+1, i)
	old := r.head
	r.head = r.mod(r.head + 1)
	// 当环已经满了，继续写入时覆盖之前的元素，tail前进1
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
}

// Dequeue 从缓存的tail拿出一个对象.
// 如果环中没有对象则返回nil.
func (r *Ring) Dequeue() interface{} {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	if r.head == -1 {
		return nil
	}
	v := r.get(r.tail)
	// 清空原来位置的数据
	r.set(r.tail, nil)

	// 更新指针
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

// 获取环的tail的值，也就是环中最早放进去的一个对象的值。
// 如果环中没有对象则返回nil.
func (r *Ring) Peek() interface{} {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.get(r.tail)
}

// Values 返回环中全部对象，并放到切片中。
// 注意，切片中每个元素引用原始对象。
func (r *Ring) Values() []interface{} {
	r.Lock()
	defer r.Unlock()

	if r.head == -1 {
		return []interface{}{}
	}
	arr := make([]interface{}, 0, r.capacity())
	for i := 0; i < r.capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

// IsFull检查环是否已满。已满的环仍旧可以继续写入，会覆盖已有值。
func (r *Ring) IsFull() bool {
	r.Lock()
	defer r.Unlock()

	return r.head != -1 && r.head == r.tail-1
}

func (r *Ring) capacity() int {
	return len(r.buff)
}

// sets a value at the given unmodified index and returns the modified index of the value
func (r *Ring) set(p int, v interface{}) {
	r.buff[r.mod(p)] = v
}

// gets a value based at a given unmodified index
func (r *Ring) get(p int) interface{} {
	return r.buff[r.mod(p)]
}

// returns the modified index of an unmodified index
func (r *Ring) mod(p int) int {
	v := p % len(r.buff)
	for v < 0 { // this bit fixes negative indices
		v += len(r.buff)
	}
	return v
}

// 初始化环
func (r *Ring) checkInit() {
	if r.buff != nil {
		return
	}

	r.buff = make([]interface{}, DefaultRingCapacity)
	for i := range r.buff {
		r.buff[i] = nil
	}
	r.head, r.tail = -1, 0
}

// 调整环的容量为size
func (r *Ring) extend(size int) {
	if size == len(r.buff) {
		return
	}

	if size < len(r.buff) {
		// shrink the buffer
		if r.head == -1 {
			// nothing in the buffer, so just shrink it directly
			r.buff = r.buff[0:size]
		} else {
			newb := make([]interface{}, 0, size)
			// buffer has stuff in it, so save the most recent stuff...
			// start at HEAD-SIZE-1 and walk forwards
			for i := size - 1; i >= 0; i-- {
				idx := r.mod(r.head - i)
				newb = append(newb, r.buff[idx])
			}
			// reset head and tail to proper values
			r.head = len(newb) - 1
			r.tail = 0
			r.buff = newb
		}
		return
	}

	// grow the buffer
	newb := make([]interface{}, size-len(r.buff))
	for i := range newb {
		newb[i] = nil
	}
	if r.head == -1 {
		// nothing in the buffer
		r.buff = append(r.buff, newb...)
	} else if r.head >= r.tail {
		// growing at the end is safe
		r.buff = append(r.buff, newb...)
	} else {
		// buffer has stuff that wraps around the end
		// have to rearrange the buffer so the contents are still in order
		part1 := make([]interface{}, len(r.buff[:r.head+1]))
		copy(part1, r.buff[:r.head+1])
		part2 := make([]interface{}, len(r.buff[r.tail:]))
		copy(part2, r.buff[r.tail:])
		r.buff = append(r.buff, newb...)
		newTail := r.mod(r.tail + len(newb))
		r.tail = newTail
		copy(r.buff[:r.head+1], part1)
		copy(r.buff[r.head+1:r.tail], newb)
		copy(r.buff[r.tail:], part2)
	}
}
