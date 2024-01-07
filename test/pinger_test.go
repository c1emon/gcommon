package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/c1emon/gcommon/pinger"
)

func TestTcpinger(t *testing.T) {
	pinger := pinger.NewTcpinger("baidu.com", 80, pinger.WithTimeout(time.Second), pinger.WithCount(10))
	st := pinger.Ping(context.Background())
	fmt.Printf("%v", st)
}

func TestPinger(t *testing.T) {
	pinger := pinger.NewICMPPinger("baidu.com", pinger.WithTimeout(time.Second))
	st := pinger.Ping(context.Background())
	fmt.Printf("%v", st)
}
