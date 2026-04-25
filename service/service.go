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
	Run(ctx context.Context, timeout time.Duration) error
	Name() string
}

var _ ServiceRunner = &defaultRunnableService{}

func WrapDefault(svc Service) *defaultRunnableService {
	return &defaultRunnableService{
		svc: svc,
		wg:  sync.WaitGroup{},
	}
}

type defaultRunnableService struct {
	svc Service
	wg  sync.WaitGroup
}

func (s *defaultRunnableService) Name() string {
	if s.svc.Name() == "" {
		return reflect.TypeOf(s.svc).String()
	}
	return s.svc.Name()
}

func (s *defaultRunnableService) Run(ctx context.Context, timeout time.Duration) error {
	s.wg.Add(1)
	stopErr := make(chan error, 1)

	if err := s.svc.Start(); err != nil {
		s.wg.Done()
		return fmt.Errorf("service start error: %w", err)
	}

	go func() {
		defer s.wg.Done()
		<-ctx.Done()

		timeoutCtx, cancelFn := context.WithTimeout(context.Background(), timeout)
		defer cancelFn()

		if err := s.svc.Stop(timeoutCtx); err != nil {
			stopErr <- fmt.Errorf("service stop error: %w", err)
			return
		}
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			stopErr <- fmt.Errorf("service stop: context deadline exceeded")
			return
		}
		stopErr <- nil
	}()

	s.wg.Wait()
	return <-stopErr
}
