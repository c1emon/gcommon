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
func New(repo *service.ServiceRepo, logger *slog.Logger, serverOptions ...serverOption) (*Server, error) {
	s, err := newServer(repo, logger)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	fromOptions(s, serverOptions...)
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
		shutdownTimeout:  time.Second * time.Duration(5),
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

	if err := s.preRun(s.context); err != nil {
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
			s.logger.Debug("starting background service", "name", service.Name())
			// block!
			err := service.Run(s.context, s.serviceStopTimeout())
			// Do not return context.Canceled error since errgroup.Group only
			// returns the first error to the caller - thus we can miss a more
			// interesting error.
			if err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Error("stop background service failed", "name", service.Name(), "error", err)
				return fmt.Errorf("%s stop error: %w", service.Name(), err)
			}
			s.logger.Debug("stopped background service", "name", service.Name())
			return nil
		})

	}

	if err := s.postRun(s.context); err != nil {
		return err
	}
	return s.childRoutines.Wait()
}

// serviceStopTimeout is the deadline passed to each service's Stop; kept
// slightly below shutdownTimeout so Run can finish and close shutdownFinished
// before the outer Shutdown context times out.
func (s *Server) serviceStopTimeout() time.Duration {
	const margin = time.Second
	if s.shutdownTimeout <= margin {
		t := s.shutdownTimeout * 4 / 5
		if t <= 0 {
			return s.shutdownTimeout
		}
		return t
	}
	return s.shutdownTimeout - margin
}

// Shutdown begins graceful shutdown: preStop hooks, cancel root context,
// wait for Run to finish background services, then postStop hooks.
// Safe to call from any goroutine once per process shutdown; pairs internally
// with Run's shutdownWG so Run does not exit until Shutdown completes.
func (s *Server) Shutdown(ctx context.Context, reason string) error {
	s.shutdownWG.Add(1)
	defer s.shutdownWG.Done()

	var err error
	s.shutdownOnce.Do(func() {
		s.logger.Info("shutdown started", "reason", reason)
		if e := s.preStop(s.context); e != nil {
			s.logger.Error("server preStop failed", "error", e)
			err = errors.Join(err, e)
		}
		// Call cancel func to stop background services.
		s.shutdownFn()
		// Wait for server to shut down
		select {
		case <-s.shutdownFinished:
			s.logger.Info("server shutdown success")
		case <-ctx.Done():
			s.logger.Error("timed out while waiting for server to shutdown")
			err = errors.Join(err, fmt.Errorf("timeout waiting for shutdown"))
		}
		postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.shutdownTimeout)
		defer cancel()
		if e := s.postStop(postCtx); e != nil {
			s.logger.Error("server postStop failed", "error", e)
			err = errors.Join(err, e)
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
			s.handleShutdownSignal(ctx, sig)
			return
		}
	}
}

func (s *Server) handleShutdownSignal(parent context.Context, sig os.Signal) {
	shutdownCtx, cancel := context.WithTimeout(parent, s.shutdownTimeout)
	defer cancel()
	if err := s.Shutdown(shutdownCtx, fmt.Sprintf("system signal -> %s", sig)); err != nil {
		s.logger.Error("server shutdown finished with errors", "error", err)
	}
}
