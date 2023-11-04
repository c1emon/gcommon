package service

import (
	"context"
	"sync"
	"time"
)

type Service interface {
	Priority() int
	Name() string
	Start() error
	Stop(timeOutCtx context.Context) error
}

func Wrap(svc Service) *RunableService {
	return &RunableService{
		svc: svc,
	}
}

type RunableService struct {
	svc Service
}

func (s *RunableService) Run(ctx context.Context, timeout int) error {
	var wg sync.WaitGroup
	wg.Add(1)

	// handle http shutdown on server context done
	go func() {
		defer wg.Done()
		<-ctx.Done()

		// shutdown server here
		srv_ctx, cancelFn := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancelFn()
		if err := s.svc.Stop(srv_ctx); err != nil {
			// s.log.Fatal("Server Shutdown:", err)
			panic("svc shutdown ")
		}
		// catching ctx.Done(). timeout of 5 seconds.
		<-srv_ctx.Done()
		// s.log.Println("timeout of 5 seconds.")

	}()

	// start server here
	// block
	if err := s.svc.Start(); err != nil {
		return err
	}

	wg.Wait()
	return nil
}
