package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LogRequest middleware логирует метод, URI и длительность каждого входящего запроса.
// Если в context присутствует UserID, он также включается в лог.
func (m *Middleware) LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		handler.ServeHTTP(w, r)

		fields := []zap.Field{
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
		}

		if userID, ok := UserIDFromContext(r.Context()); ok {
			fields = append(fields, zap.Int64("user_id", userID))
		}

		m.logger.Info("request", fields...)
	})
}
