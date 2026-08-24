package main

import (
	"fmt"
	"sort"
)

var ages = map[string]int{"dx": 80, "shane": 20, "xiang": 18}

func main() {
	var names []string
	for name := range ages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("%s\t%d\n", name, ages[name])
	}
}
