package health

import (
	"net/http"

	"github.com/hellofresh/health-go/v5"
)

func Handler(svcName, version string) http.Handler {
	h, _ := health.New(
		health.WithComponent(health.Component{
			Name:    svcName,
			Version: version,
		}))

	return h.Handler()
}
