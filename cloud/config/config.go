package config

import "io"

type RemoteResponse struct {
	Value []byte
	Error error
}

// https://github.com/spf13/viper/blob/c4dcd31f68e5d77ce447c0091dd1ca6d7e169807/viper.go#L413
type Configer interface {
	Get(path string) (io.Reader, error)
	Watch(path string) (io.Reader, error)
	WatchChannel(path string) (<-chan *RemoteResponse, chan bool)
}
