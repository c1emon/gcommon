// Package healthgin exposes a Gin handler backed by hellofresh/health-go.
package healthgin

import (
	"github.com/c1emon/gcommon/health/v2"
	"github.com/c1emon/gcommon/health/v2/bridge"
	"github.com/gin-gonic/gin"
)

// Handler returns a [gin.HandlerFunc] that serves the same JSON payload as
// [github.com/c1emon/gcommon/health/v2/http.Handler].
func Handler(cfg health.Config) (gin.HandlerFunc, error) {
	h, err := bridge.NewHealth(cfg)
	if err != nil {
		return nil, err
	}
	return gin.WrapH(h.Handler()), nil
}

// MustHandler is like [Handler] but panics if cfg is invalid or health setup fails.
func MustHandler(cfg health.Config) gin.HandlerFunc {
	hf, err := Handler(cfg)
	if err != nil {
		panic(err)
	}
	return hf
}
