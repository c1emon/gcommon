package test

import (
	"context"
	"fmt"
	"time"

	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/server"
	"github.com/c1emon/gcommon/service"
)

func NewTestSVC(logger logx.Logger) *TestSVC {
	rootCtx, shutdownFn := context.WithCancel(context.Background())

	return &TestSVC{
		logger:           logger,
		ctx:              rootCtx,
		cancelFn:         shutdownFn,
		shutdownFinished: make(chan any),
	}
}

type TestSVC struct {
	logger           logx.Logger
	ctx              context.Context
	cancelFn         context.CancelFunc
	shutdownFinished chan any
}

func (s *TestSVC) Name() string {
	return "test svc"
}

func (s *TestSVC) Start() error {
	s.logger.Info("start test svc")
	go func() {
		defer close(s.shutdownFinished)
		ticker := time.NewTicker(time.Millisecond * time.Duration(500))
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				// timeout test
				// time.Sleep(time.Minute)
				s.logger.Info("stop svc")
				return
			case <-ticker.C:
				s.logger.Info("svc sleep...")
			}
		}

	}()
	return nil
}

func (s *TestSVC) Stop(timeOutCtx context.Context) error {
	s.cancelFn()

	select {
	case <-s.shutdownFinished:
		s.logger.Info("test svc shutdown")
		return nil
	case <-timeOutCtx.Done():
		s.logger.Warn("timed out while waiting for server to shutdown")
		return fmt.Errorf("timeout waiting for shutdown")
	}
}

func main() {
	logFactory := logx.NewLogrusLoggerFactory(logx.LevelDebug)

	repo := service.NewServiceRepo()
	repo.Registe(NewTestSVC(logFactory.Get("test svc")))

	server, _ := server.New(repo, logFactory.Get("svc test logger"), server.PreRunFunc(func(ctx context.Context) error {
		fmt.Printf("pre run opt")
		return nil
	}))

	go server.ListenToSystemSignals(context.Background())
	server.Run()

	// time.Sleep(time.Second)
}
