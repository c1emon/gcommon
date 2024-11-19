package ginx

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func New(logger *slog.Logger) *gin.Engine {

	gin.SetMode(gin.DebugMode)
	eng := gin.New()
	eng.Use(LogrusLogger(logger), ErrorHandler(), Recovery(logger))

	return eng
}
