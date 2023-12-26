package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/server"
	"github.com/c1emon/gcommon/service"
)

type TestSVC struct {
	delay int
}

func (s *TestSVC) Name() string {
	return "test svc"
}

func (s *TestSVC) Start() error {
	return nil
}

func (s *TestSVC) Stop(timeOutCtx context.Context) error {
	i := 0
	for i > s.delay {
		select {
		case <-timeOutCtx.Done():
			return timeOutCtx.Err()
		default:
			time.Sleep(time.Millisecond * time.Duration(100))
			i += 1
		}
	}
	return nil
}

func TestSVCA(t *testing.T) {
	repo := service.NewServiceRepo()
	repo.Registe(&TestSVC{
		delay: 10,
	})

	logFactory := logx.NewLogrusLoggerFactory(logx.LevelDebug)

	server, _ := server.New(repo, logFactory.Get("svc test logger"), server.PreRunFunc(func(ctx context.Context) error {
		fmt.Printf("pre run opt")
		return nil
	}))

	go func() {
		time.Sleep(time.Second * time.Duration(5))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx, "exit"); err != nil {
			fmt.Fprintf(os.Stderr, "Timed out waiting for server to shut down\n")
		}
	}()
	server.Run()

}
