package support

// 三元运算符 if
func If[T any](judge bool, valueA, valueB T) T {
	if judge {
		return valueA
	} else {
		return valueB
	}
}
