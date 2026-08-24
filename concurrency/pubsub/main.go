package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	subscriber chan interface{}
	topicFunc  func(v interface{}) bool
)

type Publisher struct {
	rw           sync.RWMutex
	buffer       int
	timeout      time.Duration
	subscribers  map[subscriber]topicFunc
}

func NewPublisher(t time.Duration, bu int) *Publisher {
	return &Publisher{
		buffer: bu,
		timeout: t,
		subscribers: make(map[subscriber]topicFunc),
	}
}

func (pub *Publisher) Subscribe() chan interface{}{
    return pub.SubscribeTopic(nil)
}

func (pub *Publisher) SubscribeTopic(topic topicFunc) chan interface{}{
	ch := make(chan interface{}, pub.buffer)
	pub.rw.Lock()
	defer pub.rw.Unlock()
	pub.subscribers[ch] = topic
	return ch
}

func (pub *Publisher) Evict(sub chan interface{}) {
    pub.rw.Lock()
    defer pub.rw.Unlock()
    delete(pub.subscribers, sub)
    close(sub)
}

func (pub *Publisher) Close()  {
    pub.rw.Lock()
    defer pub.rw.Unlock()

    for sub := range pub.subscribers {
    	delete(pub.subscribers, sub)
    	close(sub)
	}
}

func (pub *Publisher) publish(v interface{})  {
    pub.rw.Lock()
    defer pub.rw.Unlock()

    var wg sync.WaitGroup
    for sub, topic := range pub.subscribers {
    	wg.Add(1)
    	go pub.sendTopic(sub, topic, v, &wg)
	}
	wg.Wait()
}

func (pub *Publisher) sendTopic(sub chan interface{}, topic topicFunc, v interface{}, wg *sync.WaitGroup)  {
    defer wg.Done()
    if topic != nil && !topic(v) {
    	return
	}
	select {
	case sub <- v:
		case <- time.After(pub.timeout):
	}
}

func main() {
    publisher := NewPublisher(100 * time.Millisecond, 10)
    defer publisher.Close()

    all := publisher.Subscribe()
    golang := publisher.SubscribeTopic(func(v interface{}) bool {
		if s, ok := v.(string); ok {
			return strings.Contains(s, "golang")
		}
		return false
    })
    publisher.publish("hello world!")
    publisher.publish("hello golang!")

    go func() {
    	for msg := range all {
    		fmt.Println("all:", msg)
		}
	}()

    go func() {
    	for msg := range golang {
    		fmt.Println("golang:", msg)

		}
    }()
    time.Sleep(3*time.Second)
}




