// Package bridge wires github.com/hellofresh/health-go/v5 for sibling packages
// health/http and health/gin. Most callers should use those packages instead.
package bridge

import (
	"errors"
	"strings"

	"github.com/c1emon/gcommon/health/v2"
	hf "github.com/hellofresh/health-go/v5"
)

// NewHealth builds a hellofresh Health instance with component metadata and
// system/runtime metrics enabled (Go version, goroutines, heap stats).
func NewHealth(cfg health.Config) (*hf.Health, error) {
	name := strings.TrimSpace(cfg.ServiceName)
	if name == "" {
		return nil, errors.New("health: Config.ServiceName is required")
	}
	return hf.New(
		hf.WithComponent(hf.Component{
			Name:    name,
			Version: strings.TrimSpace(cfg.Version),
		}),
		hf.WithSystemInfo(),
	)
}
