package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/c1emon/gcommon/logx"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type BackgroundService interface {
	Run(ctx context.Context) error
}

// from https://github.com/grafana/grafana/blob/4cc72a22ad03132295ab3428ed9877ba2cb42eb2/pkg/server/server.go
func New() (*Server, error) {
	s, err := newServer()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

func newServer() (*Server, error) {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	s := &Server{
		context:          childCtx,
		childRoutines:    childRoutines,
		shutdownFn:       shutdownFn,
		shutdownFinished: make(chan any),
		log:              logx.GetLogger(),
		// cfg:                cfg,
		backgroundServices: make([]BackgroundService, 0),
	}

	return s, nil
}

type Server struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
	log           *logrus.Logger
	// cfg              *setting.Config
	shutdownOnce     sync.Once
	shutdownFinished chan any
	isInitialized    bool
	mtx              sync.Mutex

	// pidFile     string
	// version     string
	// commit      string
	// buildBranch string

	backgroundServices []BackgroundService
}

func (s *Server) RegistSvc(svc BackgroundService) {
	s.backgroundServices = append(s.backgroundServices, svc)
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

	services := s.backgroundServices
	// Start background services.
	for _, svc := range services {

		service := svc
		svcName := reflect.TypeOf(service).String()

		s.childRoutines.Go(func() error {
			select {
			case <-s.context.Done():
				return s.context.Err()
			default:
			}

			// start service
			s.log.Debugf("Starting background service: %s", svcName)
			err := service.Run(s.context)
			// Do not return context.Canceled error since errgroup.Group only
			// returns the first error to the caller - thus we can miss a more
			// interesting error.
			if err != nil && !errors.Is(err, context.Canceled) {
				s.log.Errorf("Stopped background service: %s for %s", "http server", err)
				return fmt.Errorf("%s run error: %w", "http server", err)
			}
			s.log.Debugf("Stopped background service %s for %s", svcName, err)
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
		s.log.Infof("Shutdown started: %s", reason)
		// Call cancel func to stop background services.
		s.shutdownFn()
		// Wait for server to shut down
		select {
		case <-s.shutdownFinished:
			s.log.Debug("Finished waiting for server to shut down")
		case <-ctx.Done():
			s.log.Warn("Timed out while waiting for server to shut down")
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
			if err := s.Shutdown(ctx, fmt.Sprintf("System signal: %s", sig)); err != nil {
				fmt.Fprintf(os.Stderr, "Timed out waiting for server to shut down\n")
			}
			return
		}
	}
}
