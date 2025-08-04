package main

import (
	"fmt"
	"sync"
)

// @Author dx
// @Date 2025-08-04 21:52:00
// @Desc

// Observer 泛型观察者接口
type Observer[T any] interface {
	Update(data T)
}

// Subject 泛型主题
type Subject[T any] struct {
	mu        sync.RWMutex
	observers []Observer[T]
}

func NewSubject[T any]() *Subject[T] {
	return &Subject[T]{
		observers: make([]Observer[T], 0),
	}
}

func (s *Subject[T]) Attach(observer Observer[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *Subject[T]) Notify(data T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, observer := range s.observers {
		go observer.Update(data)
	}
}

// 具体的数据类型
type StockPrice struct {
	Symbol string
	Price  float64
}

// 具体观察者
type StockDisplay struct {
	name string
}

func (d *StockDisplay) Update(price StockPrice) {
	fmt.Printf("%s: %s 当前价格 %.2f\n", d.name, price.Symbol, price.Price)
}

func main() {
	// 创建股票价格主题
	stockSubject := NewSubject[StockPrice]()

	// 创建观察者
	display1 := &StockDisplay{name: "显示器1"}
	display2 := &StockDisplay{name: "显示器2"}

	// 注册观察者
	stockSubject.Attach(display1)
	stockSubject.Attach(display2)

	// 发送股票价格更新
	stockSubject.Notify(StockPrice{Symbol: "AAPL", Price: 150.25})
	stockSubject.Notify(StockPrice{Symbol: "GOOGL", Price: 2750.80})

	fmt.Scanln()
}
