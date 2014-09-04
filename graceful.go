package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

type GracefulShutdown struct {
	timeout   time.Duration
	wg        sync.WaitGroup
	functions []func()
}

func NewGracefulShutdown(t time.Duration) *GracefulShutdown {
	return &GracefulShutdown{timeout: t}
}

func (g *GracefulShutdown) Handler(c martini.Context) {
	g.wg.Add(1)
	c.Next()
	g.wg.Done()
}

func (g *GracefulShutdown) RunOnShutDown(f func()) error {
	g.functions = append(g.functions, f)
	return nil
}
func (g *GracefulShutdown) WaitForSignal(signals ...os.Signal) error {
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, signals...)
	<-sigchan

	log.Println("Waiting for all requests to finish")
	for _, f := range g.functions {
		f()
	}

	waitChan := make(chan struct{})
	go func() {
		g.wg.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-time.After(g.timeout):
		return fmt.Errorf("timed out waiting %v for shutdown", g.timeout)
	case <-waitChan:
		return nil
	}
}

type ConnectionLimit struct {
	numConnections int32
	limit          int32
}

func (c *ConnectionLimit) Handler(ctx martini.Context, rw http.ResponseWriter) {
	if atomic.AddInt32(&c.numConnections, 1) > c.limit {
		http.Error(rw, "maximum connections exceeded", http.StatusServiceUnavailable)
		atomic.AddInt32(&c.numConnections, -1)
		return
	}

	ctx.Next()
	atomic.AddInt32(&c.numConnections, -1)
}
