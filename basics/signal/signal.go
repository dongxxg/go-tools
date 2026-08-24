package signal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WaitExit will block until os signal happened
func WaitExit(c chan os.Signal, exit func()) {
	for i := range c {
		switch i {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fmt.Println("receive exit signal ", i.String(), ",exit...")
			exit()
			os.Exit(0)
		}
	}
}

// NewShutdownSignal new normal Signal channel
func NewShutdownSignal() chan os.Signal {
	c := make(chan os.Signal)
	// SIGHUP: terminal closed
	// SIGINT: Ctrl+C
	// SIGTERM: program exit
	// SIGQUIT: Ctrl+/
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return c
}


func HttpExit() {
	s := &http.Server{}
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<- quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		//close() //
	}
	log.Println("Server exiting")
}
