//go:build windows

package proc

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongxxg/go-tools/log"
)

var done = make(chan struct{})

// Windows 无 SIGUSR1/2 与 syscall.Kill：init 只挂 Ctrl+C(os.Interrupt) /
// SIGTERM 走优雅停机监听器；profiling 转储（SIGUSR 系列）仅 unix 提供。
func init() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		v := <-signals
		log.Infof("Got signal %v, shutting down...", v)
		select {
		case <-done:
			// already closed
		default:
			close(done)
		}

		go wrapUpListeners.notifyListeners()
		time.Sleep(wrapUpTime)
		go shutdownListeners.notifyListeners()
		time.Sleep(delayTimeBeforeForceQuit - wrapUpTime)
		os.Exit(1)
	}()
}

// Done returns the channel that notifies the process quitting.
func Done() <-chan struct{} {
	return done
}
