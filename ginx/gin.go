package ginx

import (
	"github.com/c1emon/gcommon/util"
	"github.com/gin-gonic/gin"
)

func New(opts ...*util.FuncOption[gin.Engine]) *gin.Engine {

	eng := gin.New()
	for _, opt := range opts {
		opt.Apply(eng)
	}

	return eng
}

func WithMiddleware(h gin.HandlerFunc) *util.FuncOption[gin.Engine] {
	return util.WrapFuncOption(func(eng *gin.Engine) {
		eng.Use(h)
	})
}
