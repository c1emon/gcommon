// Package httphealth exposes a net/http handler backed by hellofresh/health-go.
package httphealth

import (
	"net/http"

	"github.com/c1emon/gcommon/health/v2"
	"github.com/c1emon/gcommon/health/v2/bridge"
)

// Handler returns an [http.Handler] that serves JSON health status including
// component fields and, when checks pass, basic Go runtime metrics (via
// hellofresh [github.com/hellofresh/health-go/v5] WithSystemInfo).
func Handler(cfg health.Config) (http.Handler, error) {
	h, err := bridge.NewHealth(cfg)
	if err != nil {
		return nil, err
	}
	return h.Handler(), nil
}

// MustHandler is like [Handler] but panics if cfg is invalid or health setup fails.
func MustHandler(cfg health.Config) http.Handler {
	h, err := Handler(cfg)
	if err != nil {
		panic(err)
	}
	return h
}
