package server

import (
	"context"
	"time"

	"github.com/c1emon/gcommon/util"
)

type serverOption util.Option[serverOptions]

type serverOptions struct {
	preRunFunc      func(context.Context) error
	postRunFunc     func(context.Context) error
	preStopFunc     func(context.Context) error
	postStopFunc    func(context.Context) error
	shutdownTimeout time.Duration
}

func fromOptions(server *Server, opts ...serverOption) {
	o := &serverOptions{
		preRunFunc:      server.preRunFunc,
		postRunFunc:     server.postRunFunc,
		preStopFunc:     server.preStopFunc,
		postStopFunc:    server.postStopFunc,
		shutdownTimeout: server.shutdownTimeout,
	}
	for _, v := range opts {
		v.Apply(o)
	}
	server.preRunFunc = o.preRunFunc
	server.postRunFunc = o.postRunFunc
	server.preStopFunc = o.preStopFunc
	server.postStopFunc = o.postStopFunc
	server.shutdownTimeout = o.shutdownTimeout
}

func PreRunFunc(f func(context.Context) error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.preRunFunc = f
		})
}

func PostRunFunc(f func(context.Context) error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.postRunFunc = f
		})
}

func PreStopFunc(f func(context.Context) error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.preStopFunc = f
		})
}

func PostStopFunc(f func(context.Context) error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.postStopFunc = f
		})
}

func WithShutdownTimeout(timeout time.Duration) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			if timeout < time.Duration(2)*time.Second {
				timeout = time.Duration(2) * time.Second
			}
			so.shutdownTimeout = timeout
		})
}
