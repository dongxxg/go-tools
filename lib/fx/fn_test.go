package fx

import (
	"testing"
)

func TestGetFuncCaller(t *testing.T) {
	caller := GetFuncCaller()
	t.Log(caller)
}
