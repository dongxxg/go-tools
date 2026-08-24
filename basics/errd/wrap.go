package main

import (
	"errors"
	"fmt"
)

func Wrap(err *error, f string, v ...interface{}) {
	if *err != nil {
		*err = fmt.Errorf(f+": %w", append(v, *err)...)
	}
}

func testWrap() (err error) {
	defer Wrap(&err, "failed to process")

	err = errors.New("test error")
	return fmt.Errorf("failed err: %w", err)
}

func testPanic() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic ", err)
		}
	}()

	//i := 10
	//j := i / 0

	i := []int{1, 2, 3}
	fmt.Println("j", i[4])
}

func main() {
	testPanic()
    //log.Println(testWrap())
}

