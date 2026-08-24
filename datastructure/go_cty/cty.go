package main

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"
)

// go get -u github.com/zclconf/go-cty/cty

func main() {
	intValue := cty.NumberIntVal(12)

	stringValue := cty.StringVal("shane")

	boolValue := cty.BoolVal(true)

	listValue := cty.ListVal([]cty.Value{
		cty.NumberIntVal(1),
		cty.NumberIntVal(2),
		cty.NumberIntVal(3),
	})

	objectValue := cty.ObjectVal(map[string]cty.Value{
		"name": cty.StringVal("dx"),
		"age":  cty.NumberIntVal(20),
	})

	fmt.Println("intValue:", intValue.Type(), intValue)
	fmt.Println("stringValue:", stringValue.Type(), stringValue)
	fmt.Println("boolValue:", boolValue.Type(), boolValue)
	fmt.Println("listValue:", listValue.Type(), listValue)
	fmt.Println("objectValue:", objectValue.Type(), objectValue)

	if listValue.Type().IsListType() {
		v := listValue.Index(cty.NumberIntVal(1))
		fmt.Println("list element:", v)
	}

	if objectValue.Type().IsObjectType() {
		name := objectValue.GetAttr("name")
		age := objectValue.GetAttr("age")
		fmt.Println("name:", name)
		fmt.Println("age:", age)
	}
}
