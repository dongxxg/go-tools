package abstract_factory

import "fmt"

// 客户端代码：
func main() {
	adidasFactory, _ := getSportsFactory("adidas")
	nikeFactory, _ := getSportsFactory("nike")

	nikeShoe := nikeFactory.makeShoe()
	nikeShirt := nikeFactory.makeShirt()
	fmt.Println(nikeShoe, nikeShirt)

	adidasShoe := adidasFactory.makeShoe()
	adidasShirt := adidasFactory.makeShirt()
	fmt.Println(adidasShoe, adidasShirt)
}
