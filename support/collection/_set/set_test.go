package _set

import (
	"fmt"
	"testing"
)

func TestSet(t *testing.T) {
	s := New("A")
	fmt.Println(s)
	s.Add("B")
	fmt.Println(s)
	s.Add("A")
	fmt.Println(s)
	fmt.Println("has C:", s.Has("C"))
	fmt.Println("has A:", s.Has("A"))
	s.Remove("C")
	fmt.Println(s)
	s.Remove("B")
	fmt.Println(s)
}

func TestGetOne(t *testing.T) {
	set := New[int]()
	a, ok := set.GetOne()
	fmt.Println("a:", a)
	fmt.Println("ok:", ok)
}
