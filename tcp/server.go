package tcp

import (
	"context"
	"github.com/qianyiwangmeng/go-redis/interface/tcp"
	"github.com/qianyiwangmeng/go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan // 用协程开，不影响本节运行
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe closeChan 感知系统信号
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {

	go func() {
		<-closeChan
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background() // 上下文设置过期时间，超时时间
	var waitDone sync.WaitGroup
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait() // 协程工作完再退出
}
