package main

// Strategy 定义一个策略接口
type Strategy interface {
    Calculate(a, b int) int
}

// AddStrategy 实现加法策略
type AddStrategy struct{}

func (s *AddStrategy) Calculate(a, b int) int {
    return a + b
}

// SubtractStrategy 实现减法策略
type SubtractStrategy struct{}

func (s *SubtractStrategy) Calculate(a, b int) int {
    return a - b
}

// Context 定义上下文
type Context struct {
    strategy Strategy
}

// SetStrategy 设置策略
func (c *Context) SetStrategy(strategy Strategy) {
    c.strategy = strategy
}

// Execute 执行策略
func (c *Context) Execute(a, b int) int {
    return c.strategy.Calculate(a, b)
}

func main() {
    // 创建上下文
    context := &Context{}

    // 设置加法策略
	addStrategy := &AddStrategy{}
	context.SetStrategy(addStrategy)
	result := context.Execute(5, 3)
	println("5 + 3 =", result) // 输出: 5 + 3 = 8

	// 设置减法策略
	subStrategy := &SubtractStrategy{}
	context.SetStrategy(subStrategy)
	result = context.Execute(5, 3)
	println("5 - 3 =", result) // 输出: 5 - 3 = 2
}
