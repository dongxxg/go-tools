package collection

import "fmt"

// 有序二元组
type OrderedPair struct {
	First  interface{}
	Second interface{}
}

// Equal 两个元素相同且顺序一致
func (pair *OrderedPair) Equal(other OrderedPair) bool {
	if pair.First == other.First &&
		pair.Second == other.Second {
		return true
	}

	return false
}

// String 以"(A, B)"形式输出字符串
func (pair OrderedPair) String() string {
	return fmt.Sprintf("(%v, %v)", pair.First, pair.Second)
}
