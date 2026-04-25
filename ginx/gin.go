package ginx

import (
	"github.com/c1emon/gcommon/util"
	"github.com/gin-gonic/gin"
)

// New builds a gin engine with optional configuration applied in order.
func New(opts ...util.Option[gin.Engine]) *gin.Engine {
	eng := gin.New()
	for _, opt := range opts {
		opt.Apply(eng)
	}
	return eng
}

// WithMiddleware registers a handler in the engine's global middleware chain.
func WithMiddleware(h gin.HandlerFunc) util.Option[gin.Engine] {
	return util.WrapFuncOption(func(eng *gin.Engine) {
		eng.Use(h)
	})
}
