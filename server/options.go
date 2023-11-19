package server

import (
	"github.com/c1emon/gcommon/util"
)

type serverOption util.Option[serverOptions]

type serverOptions struct {
	preRunFunc   func() error
	postRunFunc  func() error
	preStopFunc  func() error
	postStopFunc func() error
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

func PreRunFunc(f func() error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.preRunFunc = f
		})
}

func PostRunFunc(f func() error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.postRunFunc = f
		})
}

func PreStopFunc(f func() error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.preStopFunc = f
		})
}

func PostStopFunc(f func() error) serverOption {
	return util.WrapFuncOption[serverOptions](
		func(so *serverOptions) {
			so.postStopFunc = f
		})
}
