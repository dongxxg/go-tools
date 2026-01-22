package errors

// ErrMaxRecursion 递归深度超限
type ErrMaxRecursion struct{}

func (e ErrMaxRecursion) Error() string {
	return "the maximum recursion depth has been reached"
}
