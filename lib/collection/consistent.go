package collection

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// ErrEmptyCircle 当环为空时尝试获取物理节点的错误
var ErrEmptyCircle = errors.New("empty circle")
var ErrBadParameter = errors.New("bad parameter")

// 一致性哈希是指将「存储节点」和「数据」都映射到一个首尾相连的哈希环上
// 不同的虚拟节点有不同的hash值,但都对应同一个物理node。
// 虚拟node均衡分布在环上,数据存储在顺时针方向的虚拟node上
type Consistent struct {
	circle           map[uint32]string // 环上的虚拟节点, map的value为对应的物理节点名
	members          map[string]bool   // 物理节点，key为节点名
	sortedHashes     uints             // 虚拟节点circle的key
	NumberOfReplicas int               // 每个物理节点对应的虚拟节点个数
	count            int64             // 环上物理节点个数

	sync.RWMutex
}

// NewConsistent 新建一个一致性哈希环，默认每个物理节点有20个虚拟节点。
// 在Add物理节点后不能再修改虚拟节点数。
func NewConsistent() *Consistent {
	c := new(Consistent)
	c.NumberOfReplicas = 20
	c.circle = make(map[uint32]string)
	c.members = make(map[string]bool)
	return c
}

// 向一致性哈希环添加物理节点
func (c *Consistent) Add(elt string) {
	c.Lock()
	defer c.Unlock()
	c.add(elt)
}

// 从环中删除物理节点，节点不存在时无操作
func (c *Consistent) Remove(elt string) {
	c.Lock()
	defer c.Unlock()
	c.remove(elt)
}

// 设置环的全部物理节点。如果环中已有节点不在elts中，则会被删除
func (c *Consistent) Set(elts []string) {
	c.Lock()
	defer c.Unlock()
	for k := range c.members {
		found := false
		for _, v := range elts {
			if k == v {
				found = true
				break
			}
		}
		if !found {
			c.remove(k)
		}
	}
	for _, v := range elts {
		_, exists := c.members[v]
		if exists {
			continue
		}
		c.add(v)
	}
}

// Members 返回环中现有物理节点
func (c *Consistent) Members() []string {
	c.RLock()
	defer c.RUnlock()
	var m []string
	for k := range c.members {
		m = append(m, k)
	}
	return m
}

// Get根据name的hash值返回环中最接近的物理节点
func (c *Consistent) Get(name string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.circle) == 0 {
		return "", ErrEmptyCircle
	}
	key := c.hashKey(name)
	i := c.search(key)
	return c.circle[c.sortedHashes[i]], nil
}

// GetN 根据name的hash值返回环中最接近的N个不同物理节点
func (c *Consistent) GetN(name string, n int) ([]string, error) {
	if n < 1 {
		return nil, ErrBadParameter
	}

	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return nil, ErrEmptyCircle
	}

	if c.count < int64(n) {
		n = int(c.count)
	}

	var (
		key   = c.hashKey(name)
		i     = c.search(key)
		start = i
		res   = make([]string, 0, n)
		elem  = c.circle[c.sortedHashes[i]]
	)

	res = append(res, elem)

	if len(res) == n {
		return res, nil
	}

	for i = start + 1; i != start; i++ {
		if i >= len(c.sortedHashes) {
			i = 0
		}
		elem = c.circle[c.sortedHashes[i]]
		if !sliceContainsMember(res, elem) {
			res = append(res, elem)
		}
		if len(res) == n {
			break
		}
	}

	return res, nil
}

// 一致哈希算法是对 2^32 进行取模运算
func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	// reallocate if we're holding on too much (1/4th)
	if cap(c.sortedHashes)/(c.NumberOfReplicas*4) > len(c.circle) {
		hashes = nil
	}
	for k := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(hashes)
	c.sortedHashes = hashes
}

func sliceContainsMember(set []string, member string) bool {
	for _, m := range set {
		if m == member {
			return true
		}
	}
	return false
}

// need c.Lock() before calling
func (c *Consistent) add(elt string) {
	for i := 0; i < c.NumberOfReplicas; i++ {
		c.circle[c.hashKey(c.eltKey(elt, i))] = elt
	}
	c.members[elt] = true
	c.updateSortedHashes()
	c.count++
}

// need c.Lock() before calling
func (c *Consistent) remove(elt string) {
	if c.count <= 0 {
		return
	}

	// 从环上删除所有elt节点对应的虚节点
	for i := 0; i < c.NumberOfReplicas; i++ {
		delete(c.circle, c.hashKey(c.eltKey(elt, i)))
	}
	delete(c.members, elt)
	c.updateSortedHashes()
	c.count--
}

func (c *Consistent) search(key uint32) (i int) {
	f := func(x int) bool {
		return c.sortedHashes[x] > key
	}
	i = sort.Search(len(c.sortedHashes), f)
	if i >= len(c.sortedHashes) {
		i = 0
	}
	return
}

// eltKey generates a string key for an element with an index.
func (c *Consistent) eltKey(elt string, idx int) string {
	return strconv.Itoa(idx) + elt
}

type uints []uint32

// Len returns the length of the uints array.
func (x uints) Len() int { return len(x) }

// Less returns true if element i is less than element j.
func (x uints) Less(i, j int) bool { return x[i] < x[j] }

// Swap exchanges elements i and j.
func (x uints) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
