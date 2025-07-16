package _slice

import (
	"fmt"
	"testing"
)

func TestSetAdd(t *testing.T) {
	arr := []string{"A"}
	arr = SetAdd(arr, "B")
	fmt.Println(arr)
}
