package main

import (
	"fmt"
	"github.com/spf13/cast"
)

func raise(s []int) {
	//s = append(s, 0)
	for i := range s {
		s[i]++
	}
}

func main() {
	s1 := []int{1, 2}
	s2 := s1

	s2 = append(s2, 3)
	s2 = append(s2, 4)
	raise(s1)
	raise(s2)
	fmt.Println(s1, s2)

	s3 := []int{1, 2, 3}
	s4 := s3[1:]
	s4[1] = 5
	fmt.Println(s3, "\n", s4)

	str := "172.27.56.44:8123,172.27.56.45:8123,172.27.56.46:8123,172.27.56.22:8123"
	str2 := cast.ToStringSlice(str)
	fmt.Printf("str2=%v\n", str2)
}