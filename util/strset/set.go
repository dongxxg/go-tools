package strset

import (
	"fmt"
	"strings"
)

var (
	keyExists   = struct{}{}
	nonExistent string
)

const maxInt = int(^uint(0) >> 1)

// 字符串set
type Set struct {
	m map[string]struct{}
}

// 创建set
func New(ts ...string) *Set {
	s := NewWithSize(len(ts))
	s.Add(ts...)
	return s
}

// 使用slice创建set
func NewBySlice(ts []string) *Set {
	s := NewWithSize(len(ts))
	for _, s2 := range ts {
		s.Add(s2)
	}
	return s
}

func NewWithSize(size int) *Set {
	return &Set{make(map[string]struct{}, size)}
}

func (s *Set) Add(items ...string) {
	for _, item := range items {
		s.m[item] = keyExists
	}
}

func (s *Set) Remove(items ...string) {
	for _, item := range items {
		delete(s.m, item)
	}
}

func (s *Set) Has(items ...string) bool {
	has := false
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

func (s *Set) HasAny(items ...string) bool {
	has := false
	for _, item := range items {
		if _, has = s.m[item]; has {
			break
		}
	}
	return has
}

func (s *Set) Size() int {
	return len(s.m)
}

func (s *Set) Clear() {
	s.m = make(map[string]struct{})
}

func (s *Set) IsEmpty() bool {
	return s.Size() == 0
}

func (s *Set) IsEqual(t *Set) bool {
	// return false if they are no the same size
	if s.Size() != t.Size() {
		return false
	}

	equal := true
	t.Each(func(item string) bool {
		_, equal = s.m[item]
		return equal // if false, Each() will end
	})

	return equal
}

func (s *Set) IsSubset(t *Set) bool {
	if s.Size() < t.Size() {
		return false
	}

	subset := true

	t.Each(func(item string) bool {
		_, subset = s.m[item]
		return subset
	})

	return subset
}

func (s *Set) IsSuperset(t *Set) bool {
	return t.IsSubset(s)
}

func (s *Set) Each(f func(item string) bool) {
	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

func (s *Set) Copy() *Set {
	u := NewWithSize(s.Size())
	for item := range s.m {
		u.m[item] = keyExists
	}
	return u
}

// String returns a string representation of s
func (s *Set) String() string {
	v := make([]string, 0, s.Size())
	for item := range s.m {
		v = append(v, fmt.Sprintf("%v", item))
	}
	return fmt.Sprintf("[%s]", strings.Join(v, ", "))
}

func (s *Set) Slice() []string {
	v := make([]string, 0, s.Size())
	for item := range s.m {
		v = append(v, item)
	}
	return v
}

func (s *Set) Merge(t *Set) {
	for item := range t.m {
		s.m[item] = keyExists
	}
}

func (s *Set) Pop() (bool, string) {
	res := ""
	for item := range s.m {
		res = item
		break
	}
	if len(res) == 0 {
		return false, ""
	} else {
		s.Remove(res)
		return true, res
	}
}

// Separate removes the Set items containing in t from Set s. Please aware that
// it's not the opposite of Merge.
func (s *Set) Separate(t *Set) {
	for item := range t.m {
		delete(s.m, item)
	}
}

func Union(sets ...*Set) *Set {
	maxPos := -1
	maxSize := 0
	for i, set := range sets {
		if l := set.Size(); l > maxSize {
			maxSize = l
			maxPos = i
		}
	}
	if maxSize == 0 {
		return New()
	}

	u := sets[maxPos].Copy()
	for i, set := range sets {
		if i == maxPos {
			continue
		}
		for item := range set.m {
			u.m[item] = keyExists
		}
	}
	return u
}

func Difference(set1 *Set, sets ...*Set) *Set {
	s := set1.Copy()
	for _, set := range sets {
		s.Separate(set)
	}
	return s
}

func Intersection(sets ...*Set) *Set {
	minPos := -1
	minSize := maxInt
	for i, set := range sets {
		if l := set.Size(); l < minSize {
			minSize = l
			minPos = i
		}
	}
	if minSize == maxInt || minSize == 0 {
		return New()
	}

	t := sets[minPos].Copy()
	for i, set := range sets {
		if i == minPos {
			continue
		}
		for item := range t.m {
			if _, has := set.m[item]; !has {
				delete(t.m, item)
			}
		}
	}
	return t
}

func SymmetricDifference(s *Set, t *Set) *Set {
	u := Difference(s, t)
	v := Difference(t, s)
	return Union(u, v)
}
