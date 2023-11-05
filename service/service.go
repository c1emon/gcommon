package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Service interface {
	Name() string
	Start() error
	Stop(timeOutCtx context.Context) error
}

type ServiceRunner interface {
	Run(ctx context.Context, timeout int) error
	Name() string
}

var _ ServiceRunner = &defaultRunableService{}

func WrapDefault(svc Service) *defaultRunableService {
	return &defaultRunableService{
		svc: svc,
		wg:  sync.WaitGroup{},
	}
}

type defaultRunableService struct {
	svc Service
	wg  sync.WaitGroup
}

func (s *defaultRunableService) Name() string {
	if s.svc.Name() == "" {
		return reflect.TypeOf(s.svc).String()
	}
	return s.svc.Name()
}

func (s *defaultRunableService) Run(ctx context.Context, timeout int) error {
	s.wg.Add(1)
	var err error

	// start service here
	// non-block
	if err = s.svc.Start(); err != nil {
		s.wg.Done()
		return fmt.Errorf("service start error: %s", err)
	}

	// handle http shutdown on server context done
	go func() {
		defer s.wg.Done()
		<-ctx.Done()

		// shutdown server here
		timeoutCtx, cancelFn := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancelFn()

		if err = s.svc.Stop(timeoutCtx); err != nil {
			err = fmt.Errorf("service stop error: %s", err)
			return
		}

		// <-timeoutCtx.Done()
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			err = fmt.Errorf("service stop timeout: %s", err)
		}
	}()

	// block here
	s.wg.Wait()
	return err
}
