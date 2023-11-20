package server

import (
	"context"

	"github.com/c1emon/gcommon/util"
)

type serverOption util.Option[serverOptions]

type serverOptions struct {
	preRunFunc   func(context.Context) error
	postRunFunc  func(context.Context) error
	preStopFunc  func(context.Context) error
	postStopFunc func(context.Context) error
}

func fromOptions(server *Server, opts ...serverOption) {
	o := &serverOptions{
		preRunFunc:   nil,
		postRunFunc:  nil,
		preStopFunc:  nil,
		postStopFunc: nil,
	}
	for _, v := range opts {
		v.Apply(o)
	}
	server.preRunFunc = o.preRunFunc
	server.postRunFunc = o.postRunFunc
	server.preStopFunc = o.preStopFunc
	server.postStopFunc = o.postStopFunc
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
