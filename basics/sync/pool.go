package main

import (
	"bytes"
	"log"
	"sync"
)

//It will reuse buffers in between calls to avoid allocations.
var bpool sync.Pool

// Get returns a buffer from the pool or creates a new one if
// the pool is empty.
func Get() *bytes.Buffer {
	b := bpool.Get()
	if b == nil {
		return &bytes.Buffer{}
	}
	return b.(*bytes.Buffer)
}

// Put returns a buffer into the pool.
func Put(b *bytes.Buffer) {
	b.Reset()
	bpool.Put(b)
}

func testPool() {
	b := []byte(`dx test`)
	temp := Get()
	defer Put(temp)

	temp.Write(b)
	log.Printf("%+v", &temp)
	var i int
	for i < 20 {
		a := Get()
		log.Printf("%+v", &a)
		i++
	}
}

func main() {
	testPool()
	bpool.New()
	log.Printf("%+v", Get())
}
