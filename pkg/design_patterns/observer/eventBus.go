package main

import (
	"fmt"
	"sync"
)

// @Author dx
// @Date 2025-08-04 21:53:00
// @Desc

// EventType 事件类型
type EventType string

const (
	EventUserLogin  EventType = "user_login"
	EventUserLogout EventType = "user_logout"
	EventOrderPaid  EventType = "order_paid"
)

// Event 事件
type Event struct {
	Type EventType
	Data interface{}
}

// EventHandler 事件处理函数
type EventHandler func(event Event)

// EventBus 事件总线
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish 发布事件
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		// 异步执行，避免阻塞
		go handler(event)
	}
}

// 使用示例
func main() {
	eventBus := NewEventBus()

	// 订阅用户登录事件
	eventBus.Subscribe(EventUserLogin, func(event Event) {
		user := event.Data.(map[string]string)
		fmt.Printf("发送欢迎邮件给: %s\n", user["email"])
	})

	eventBus.Subscribe(EventUserLogin, func(event Event) {
		user := event.Data.(map[string]string)
		fmt.Printf("记录登录日志: 用户 %s 登录\n", user["name"])
	})

	// 订阅订单支付事件
	eventBus.Subscribe(EventOrderPaid, func(event Event) {
		order := event.Data.(map[string]interface{})
		fmt.Printf("发送订单确认邮件: 订单号 %s\n", order["orderId"])
	})

	// 发布事件
	eventBus.Publish(Event{
		Type: EventUserLogin,
		Data: map[string]string{
			"name":  "张三",
			"email": "zhangsan@example.com",
		},
	})

	eventBus.Publish(Event{
		Type: EventOrderPaid,
		Data: map[string]interface{}{
			"orderId": "ORD123456",
			"amount":  99.99,
		},
	})

	// 等待异步处理完成
	fmt.Scanln()
}
