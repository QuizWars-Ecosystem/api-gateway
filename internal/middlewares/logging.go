package middlewares

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

func NewLoggingMiddleware(logger *zap.Logger) runtime.Middleware {
	return func(handlerFunc runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			logger.Debug("request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("query", r.URL.RawQuery))
			handlerFunc(w, r, pathParams)
		}
	}
}
