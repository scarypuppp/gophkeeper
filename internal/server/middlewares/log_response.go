package middlewares

import (
	"net/http"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (response *loggingResponseWriter) Write(bytes []byte) (int, error) {
	size, err := response.ResponseWriter.Write(bytes)
	response.responseData.size += size
	return size, err
}

func (response *loggingResponseWriter) WriteHeader(statusCode int) {
	response.ResponseWriter.WriteHeader(statusCode)
	response.responseData.status = statusCode
}

// LogResponse middleware логирует HTTP-статус и размер тела каждого ответа.
// Если в context присутствует UserID, он также включается в лог.
func (m *Middleware) LogResponse(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		handler.ServeHTTP(&lw, r)

		fields := []zap.Field{
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		}

		if userID, ok := UserIDFromContext(r.Context()); ok {
			fields = append(fields, zap.Int64("user_id", userID))
		}

		m.logger.Info("response", fields...)
	})
}
