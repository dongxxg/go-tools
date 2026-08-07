package fx

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebounce(t *testing.T) {
	var counter uint64

	spammyChan := make(chan interface{}, 1)
	go Debounce(300*time.Millisecond, spammyChan, func() {
		atomic.AddUint64(&counter, 1)
	})

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, uint64(0), counter)

	for i := 0; i < 5; i++ {
		spammyChan <- struct{}{}
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(400 * time.Millisecond)
	assert.Equal(t, uint64(1), counter)

	for i := 0; i < 5; i++ {
		spammyChan <- struct{}{}
		time.Sleep(400 * time.Millisecond)
	}
	assert.Equal(t, uint64(6), counter)
}
