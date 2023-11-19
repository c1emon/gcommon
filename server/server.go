package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/service"
	"golang.org/x/sync/errgroup"
)

// from https://github.com/grafana/grafana/blob/4cc72a22ad03132295ab3428ed9877ba2cb42eb2/pkg/server/server.go
func New(repo *service.ServiceRepo, logger logx.Logger) (*Server, error) {
	s, err := newServer(repo, logger)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

func newServer(repo *service.ServiceRepo, logger logx.Logger) (*Server, error) {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	s := &Server{
		context:          childCtx,
		childRoutines:    childRoutines,
		shutdownFn:       shutdownFn,
		shutdownFinished: make(chan any),
		logger:           logger,
		// cfg:                cfg,
		svcRepo: repo,
	}

	return s, nil
}

type Server struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
	logger        logx.Logger
	// cfg              *setting.Config
	shutdownOnce     sync.Once
	shutdownFinished chan any
	isInitialized    bool
	mtx              sync.Mutex

	// pidFile     string
	// version     string
	// commit      string
	// buildBranch string

	svcRepo *service.ServiceRepo
}

func (s *Server) Init() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.isInitialized {
		return nil
	}
	s.isInitialized = true

	return nil
}

func (s *Server) Run() error {
	defer close(s.shutdownFinished)

	if err := s.Init(); err != nil {
		return err
	}

	// Start background services.
	for _, svc := range s.svcRepo.Services() {

		service := svc
		s.childRoutines.Go(func() error {
			select {
			// 如果已经Done了，就停止启动流程
			case <-s.context.Done():
				return s.context.Err()
			default:
			}

			// start service
			s.logger.Debug("starting background service: %s", service.Name())
			// block!
			err := service.Run(s.context, 10)
			// Do not return context.Canceled error since errgroup.Group only
			// returns the first error to the caller - thus we can miss a more
			// interesting error.
			if err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Error("stopped background service %s error: %s", service.Name(), err)
				return fmt.Errorf("%s stop error: %w", service.Name(), err)
			}
			s.logger.Debug("stopped background service: %s", service.Name())
			return nil
		})

	}

	return s.childRoutines.Wait()
}

// Shutdown initiates Grafana graceful shutdown. This shuts down all
// running background services. Since Run blocks Shutdown supposed to
// be run from a separate goroutine.
func (s *Server) Shutdown(ctx context.Context, reason string) error {
	var err error
	s.shutdownOnce.Do(func() {
		s.logger.Info("shutdown started reason: %s", reason)
		// Call cancel func to stop background services.
		s.shutdownFn()
		// Wait for server to shut down
		select {
		case <-s.shutdownFinished:
			s.logger.Debug("finished waiting for server to shutdown")
		case <-ctx.Done():
			s.logger.Warn("timed out while waiting for server to shutdown")
			err = fmt.Errorf("timeout waiting for shutdown")
		}
	})

	return err
}

func (s *Server) ListenToSystemSignals(ctx context.Context) {
	signalChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sighupChan:
			// if err := log.Reload(); err != nil {
			// 	fmt.Fprintf(os.Stderr, "Failed to reload loggers: %s\n", err)
			// }
		case sig := <-signalChan:
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if err := s.Shutdown(ctx, fmt.Sprintf("system signal -> %s", sig)); err != nil {
				s.logger.Error("timed out waiting for server to shutdown")
			}
			return
		}
	}
}
