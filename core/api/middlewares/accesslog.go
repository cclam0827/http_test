package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLogger emits one structured access log line per request, after the
// response has been written.
func AccessLogger(l *slog.Logger) gin.HandlerFunc {
	if l == nil {
		l = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.Int("bytes", max(c.Writer.Size(), 0)), // Size() is -1 when nothing was written
			slog.String("client_ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}
		l.LogAttrs(c.Request.Context(), slog.LevelInfo, "http_request", attrs...)
	}
}
