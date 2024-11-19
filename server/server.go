package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/c1emon/gcommon/service"
	"golang.org/x/sync/errgroup"
)

// from https://github.com/grafana/grafana/blob/4cc72a22ad03132295ab3428ed9877ba2cb42eb2/pkg/server/server.go
func New(repo *service.ServiceRepo, logger *slog.Logger, servserverOptions ...serverOption) (*Server, error) {
	s, err := newServer(repo, logger)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	fromOptions(s, servserverOptions...)
	return s, nil
}

func newServer(repo *service.ServiceRepo, logger *slog.Logger) (*Server, error) {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	s := &Server{
		context:          childCtx,
		childRoutines:    childRoutines,
		shutdownFn:       shutdownFn,
		shutdownFinished: make(chan any),
		shutdownTimeout:  time.Second * time.Duration(12),
		shutdownWG:       &sync.WaitGroup{},

		logger: logger,

		preRunFunc:   nil,
		postRunFunc:  nil,
		preStopFunc:  nil,
		postStopFunc: nil,
		svcRepo:      repo,
	}

	return s, nil
}

type Server struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
	logger        *slog.Logger

	shutdownOnce     sync.Once
	shutdownFinished chan any
	shutdownTimeout  time.Duration
	isInitialized    bool
	mtx              sync.Mutex

	shutdownWG *sync.WaitGroup

	// pidFile     string
	// version     string
	// commit      string
	// buildBranch string

	preRunFunc   func(context.Context) error
	postRunFunc  func(context.Context) error
	preStopFunc  func(context.Context) error
	postStopFunc func(context.Context) error

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

func (s *Server) preRun(ctx context.Context) error {
	if s.preRunFunc != nil {
		s.logger.Debug("start server pre run task")
		return s.preRunFunc(ctx)
	}
	return nil
}

func (s *Server) postRun(ctx context.Context) error {
	if s.postRunFunc != nil {
		s.logger.Debug("start server post run task")
		return s.postRunFunc(ctx)
	}
	return nil
}

func (s *Server) preStop(ctx context.Context) error {
	if s.preStopFunc != nil {
		s.logger.Debug("start server pre stop task")
		return s.preStopFunc(ctx)
	}
	return nil
}

func (s *Server) postStop(ctx context.Context) error {
	if s.postStopFunc != nil {
		s.logger.Debug("start server post stop task")
		return s.postStopFunc(ctx)
	}
	return nil
}

func (s *Server) Run() error {
	defer s.shutdownWG.Wait()
	defer close(s.shutdownFinished)

	if err := s.Init(); err != nil {
		return err
	}

	s.preRun(s.context)

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
			s.logger.Debug("starting background service", "name", service.Name())
			// block!
			err := service.Run(s.context, 10)
			// Do not return context.Canceled error since errgroup.Group only
			// returns the first error to the caller - thus we can miss a more
			// interesting error.
			if err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Error("stopp background service failed", "name", service.Name(), "error", err)
				return fmt.Errorf("%s stop error: %w", service.Name(), err)
			}
			s.logger.Debug("stopped background service", "name", service.Name())
			return nil
		})

	}

	s.postRun(s.context)
	return s.childRoutines.Wait()
}

// Shutdown initiates Grafana graceful shutdown. This shuts down all
// running background services. Since Run blocks Shutdown supposed to
// be run from a separate goroutine.
func (s *Server) Shutdown(ctx context.Context, reason string) error {
	defer s.shutdownWG.Done()
	var err error
	s.shutdownOnce.Do(func() {
		s.logger.Info("shutdown started", "reason", reason)
		s.preStop(s.context)
		// Call cancel func to stop background services.
		s.shutdownFn()
		// Wait for server to shut down
		select {
		case <-s.shutdownFinished:
			s.logger.Info("server shutdown success")
		case <-ctx.Done():
			s.logger.Error("timed out while waiting for server to shutdown")
			err = fmt.Errorf("timeout waiting for shutdown")
		}
		s.postStop(s.context)
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
			ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
			defer cancel()
			s.shutdownWG.Add(1)
			if err := s.Shutdown(ctx, fmt.Sprintf("system signal -> %s", sig)); err != nil {
				s.logger.Error("timed out waiting for server to shutdown")
			}
			return
		}
	}
}
