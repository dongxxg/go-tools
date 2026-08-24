package collection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dongxxg/go-tools/lib/lang"
)

type Set interface {
	// 向集合中添加元素，返回是否成功
	Add(i interface{}) bool

	// 集合中元素数量
	Cardinality() int

	// 清空集合
	Clear()

	// 复制集合
	Clone() Set

	// 检查参数中元素是否在集合中
	Contains(i ...interface{}) bool

	// 集合的差集。
	// 注意元素类型必须一致，否则会panic
	Difference(other Set) Set

	// 集合的并集。
	// 注意元素类型必须一致，否则会panic
	Union(other Set) Set

	// 集合的交集。
	// 注意元素类型必须一致，否则会panic
	Intersect(other Set) Set

	// 判断集合是否相同（元素个数和值相等，与顺序无关）。
	// 注意元素类型必须一致，否则会panic
	Equal(other Set) bool

	// 判断是否是真子集。
	// 注意元素类型必须一致，否则会panic
	IsProperSubset(other Set) bool

	// 判断是否是真超集。
	// 注意元素类型必须一致，否则会panic
	IsProperSuperset(other Set) bool

	// 判断是否是子集。
	// 注意元素类型必须一致，否则会panic
	IsSubset(other Set) bool

	// 判断是否是超集。
	// 注意元素类型必须一致，否则会panic
	IsSuperset(other Set) bool

	// 遍历集合中元素并执行传入方法，如果方法返回为true，则停止执行
	Each(func(interface{}) bool)

	// 返回一个channel用于遍历
	Iter() <-chan interface{}

	// 从集合中删除单个元素
	Remove(i interface{})

	// 返回集合的字符串表示
	String() string

	// Pop removes and returns an arbitrary item from the set.
	Pop() interface{}

	// 返回全部子集
	PowerSet() Set

	// 返回两个集合的笛卡尔积
	CartesianProduct(other Set) Set

	// 返回集合元素的切片
	ToSlice() []interface{}
}

// 新建一个空集合，非并发安全
func NewSet() Set {
	set := newThreadUnsafeSet()
	return &set
}

// 新建一个空集合，并根据输入参数初始化。非并发安全
func NewSetFromSlice(s []interface{}) Set {
	a := NewSet()
	for _, item := range s {
		a.Add(item)
	}
	return a
}

type threadUnsafeSet map[interface{}]lang.PlaceholderType

func newThreadUnsafeSet() threadUnsafeSet {
	return make(threadUnsafeSet)
}

func (set *threadUnsafeSet) Add(i interface{}) bool {
	_, found := (*set)[i]
	if found {
		return false // False if it existed already
	}

	(*set)[i] = struct{}{}
	return true
}

func (set *threadUnsafeSet) Contains(i ...interface{}) bool {
	for _, val := range i {
		if _, ok := (*set)[val]; !ok {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) IsSubset(other Set) bool {
	_ = other.(*threadUnsafeSet)
	if set.Cardinality() > other.Cardinality() {
		return false
	}
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) IsProperSubset(other Set) bool {
	return set.IsSubset(other) && !set.Equal(other)
}

func (set *threadUnsafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *threadUnsafeSet) IsProperSuperset(other Set) bool {
	return set.IsSuperset(other) && !set.Equal(other)
}

func (set *threadUnsafeSet) Union(other Set) Set {
	o := other.(*threadUnsafeSet)

	unionedSet := newThreadUnsafeSet()

	for elem := range *set {
		unionedSet.Add(elem)
	}
	for elem := range *o {
		unionedSet.Add(elem)
	}
	return &unionedSet
}

func (set *threadUnsafeSet) Intersect(other Set) Set {
	o := other.(*threadUnsafeSet)

	intersection := newThreadUnsafeSet()
	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range *set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range *o {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return &intersection
}

func (set *threadUnsafeSet) Difference(other Set) Set {
	_ = other.(*threadUnsafeSet)

	difference := newThreadUnsafeSet()
	for elem := range *set {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (set *threadUnsafeSet) Clear() {
	*set = newThreadUnsafeSet()
}

func (set *threadUnsafeSet) Remove(i interface{}) {
	delete(*set, i)
}

func (set *threadUnsafeSet) Cardinality() int {
	return len(*set)
}

func (set *threadUnsafeSet) Each(cb func(interface{}) bool) {
	for elem := range *set {
		if cb(elem) {
			break
		}
	}
}

func (set *threadUnsafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for elem := range *set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (set *threadUnsafeSet) Equal(other Set) bool {
	_ = other.(*threadUnsafeSet)

	if set.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) Clone() Set {
	clonedSet := newThreadUnsafeSet()
	for elem := range *set {
		clonedSet.Add(elem)
	}
	return &clonedSet
}

func (set *threadUnsafeSet) String() string {
	items := make([]string, 0, len(*set))

	for elem := range *set {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

func (set *threadUnsafeSet) Pop() interface{} {
	for item := range *set {
		delete(*set, item)
		return item
	}
	return nil
}

func (set *threadUnsafeSet) PowerSet() Set {
	powSet := NewSet()
	nullset := newThreadUnsafeSet()
	powSet.Add(&nullset)

	for es := range *set {
		u := newThreadUnsafeSet()
		j := powSet.Iter()
		for er := range j {
			p := newThreadUnsafeSet()
			if reflect.TypeOf(er).Name() == "" {
				k := er.(*threadUnsafeSet)
				for ek := range *(k) {
					p.Add(ek)
				}
			} else {
				p.Add(er)
			}
			p.Add(es)
			u.Add(&p)
		}

		powSet = powSet.Union(&u)
	}

	return powSet
}

func (set *threadUnsafeSet) CartesianProduct(other Set) Set {
	o := other.(*threadUnsafeSet)
	cartProduct := NewSet()

	for i := range *set {
		for j := range *o {
			elem := OrderedPair{First: i, Second: j}
			cartProduct.Add(elem)
		}
	}

	return cartProduct
}

func (set *threadUnsafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, set.Cardinality())
	for elem := range *set {
		keys = append(keys, elem)
	}

	return keys
}

// MarshalJSON creates a JSON array from the set, it marshals all elements
func (set *threadUnsafeSet) MarshalJSON() ([]byte, error) {
	items := make([]string, 0, set.Cardinality())

	for elem := range *set {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}

		items = append(items, string(b))
	}

	return []byte(fmt.Sprintf("[%s]", strings.Join(items, ","))), nil
}

// UnmarshalJSON recreates a set from a JSON array, it only decodes
// primitive types. Numbers are decoded as json.Number.
func (set *threadUnsafeSet) UnmarshalJSON(b []byte) error {
	var i []interface{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	for _, v := range i {
		switch t := v.(type) {
		case []interface{}, map[string]interface{}:
			continue
		default:
			set.Add(t)
		}
	}

	return nil
}