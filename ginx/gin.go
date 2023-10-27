package ginx

import (
	"github.com/c1emon/gcommon/logx"
	"github.com/gin-gonic/gin"
)

func New(loggerFactory logx.LoggerFactory) *gin.Engine {
	logger := loggerFactory.Get("gin")

	gin.SetMode(gin.DebugMode)
	eng := gin.New()
	eng.Use(LogrusLogger(logger), ErrorHandler(), Recovery(logger))

	return eng
}
