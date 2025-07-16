package support

import (
	"runtime/debug"
	"support/logger"
	"testing"
)

func TestPanicLog(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			logger.Debug("%s ", logger.PanicLog(e, debug.Stack()))
		}
	}()
	// TT
	var f func() = nil
	f()
}
