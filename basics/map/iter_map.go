package main

import "fmt"

type Student struct {
	name string
	age  int
}

func main() {
	ss := make(map[string]*Student)
	ss["1"] = &Student{name: "dx", age: 10}
	ss["2"] = &Student{name: "dx2", age: 20}
	ss["3"] = &Student{name: "dx3", age: 30}

	for index, s := range ss {
		fmt.Println(index, s)
	}
}
