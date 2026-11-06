package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recoverer logs a handler panic through slog and answers 500. A nil writer
// stops gin from also dumping the panic to stderr.
func Recoverer(l *slog.Logger) gin.HandlerFunc {
	if l == nil {
		l = slog.Default()
	}
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err any) {
		l.ErrorContext(c.Request.Context(), "panic recovered",
			"error", err,
			"path", c.Request.URL.Path,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
