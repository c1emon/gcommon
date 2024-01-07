package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/c1emon/gcommon/pinger"
)

func TestChecker(t *testing.T) {
	pinger := pinger.NewTcpinger("baidu.com", 80, pinger.WithTimeout(time.Second))
	st := pinger.Ping(context.Background())
	fmt.Printf("%v", st)
}
