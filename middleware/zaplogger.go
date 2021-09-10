package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func NewLogger(log *zap.Logger) *Logger {
	return &Logger{log}
}

type Logger struct {
	*zap.Logger
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.Info("Before request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))
	next(rw, r)
	l.Info("After request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("headers", rw.Header()))
}
