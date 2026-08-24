package fx

import (
	"fmt"
	"runtime"
)

// If is a ternary operator
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// GetFuncCaller reports file and line number information about function invocations on the calling goroutine's stack.
func GetFuncCaller() string {
	_, file, no, ok := runtime.Caller(0)
	if ok {
		return fmt.Sprintf("file: %s, line: %d", file, no)
	}
	return ""
}
