package ginx

import (
	"sync"

	"github.com/c1emon/gcommon/logx"
	"github.com/gin-gonic/gin"
)

var eng *gin.Engine
var once = &sync.Once{}

func GetGinEng() *gin.Engine {
	once.Do(func() {

		gin.SetMode(gin.DebugMode)
		eng = gin.New()
		eng.Use(LogrusLogger(logx.GetLogger()), ErrorHandler(), Recovery(logx.GetLogger()))

	})
	return eng
}
