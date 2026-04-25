package ginx

import "github.com/gin-gonic/gin"

// EngineBuilder creates a gin engine by applying global middleware in order.
type EngineBuilder struct {
	middlewares []gin.HandlerFunc
}

// NewEngineBuilder returns a builder initialized with no middleware.
func NewEngineBuilder() *EngineBuilder {
	return &EngineBuilder{middlewares: make([]gin.HandlerFunc, 0, 4)}
}

// Use appends middleware to the global chain in order.
func (b *EngineBuilder) Use(middlewares ...gin.HandlerFunc) *EngineBuilder {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

// Build creates a new gin engine and applies all middlewares.
func (b *EngineBuilder) Build() *gin.Engine {
	eng := gin.New()
	if len(b.middlewares) > 0 {
		eng.Use(b.middlewares...)
	}
	return eng
}

// NewBareEngine creates a gin engine without default middleware.
func NewBareEngine() *gin.Engine {
	return gin.New()
}

// DefaultEngineConfig controls middleware wiring for NewDefaultEngine.
type DefaultEngineConfig struct {
	Middlewares []gin.HandlerFunc
}

// NewDefaultEngine creates an engine with the default middleware pipeline.
// Middleware order is fixed so logger observes the final response state.
// Logging is backed by logx.Default(), so call logx.Init(...) during app startup.
func NewDefaultEngine(cfg DefaultEngineConfig) *gin.Engine {
	builder := NewEngineBuilder().
		Use(Logger(), ErrorResponder(), Recovery()).
		Use(cfg.Middlewares...)
	return builder.Build()
}
