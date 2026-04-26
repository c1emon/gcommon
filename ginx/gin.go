package ginx

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

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
	Logger *slog.Logger

	Middlewares []gin.HandlerFunc
}

// NewDefaultEngine creates an engine with the default middleware pipeline.
// Middleware order is fixed so logger observes the final response state.
// cfg.Logger must be non-nil; it is passed to [Logger] and [Recovery].
func NewDefaultEngine(cfg DefaultEngineConfig) *gin.Engine {
	if cfg.Logger == nil {
		panic("ginx: DefaultEngineConfig.Logger must be non-nil")
	}
	SetGinSlogWriters(cfg.Logger)
	builder := NewEngineBuilder().
		Use(Logger(cfg.Logger), ErrorResponder(), Recovery(cfg.Logger)).
		Use(cfg.Middlewares...)
	return builder.Build()
}
