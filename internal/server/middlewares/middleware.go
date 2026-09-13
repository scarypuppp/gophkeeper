package middlewares

import (
	"github.com/scarypuppp/gophkeeper/internal/server/config"
	"go.uber.org/zap"
)

// Middleware содержит зависимости, общие для всех HTTP-middleware пакета.
type Middleware struct {
	cfg    *config.Config
	logger *zap.Logger
}

// NewMiddleware создаёт экземпляр Middleware с переданными конфигурацией и логгером.
func NewMiddleware(cfg *config.Config, logger *zap.Logger) Middleware {
	return Middleware{cfg, logger}
}
