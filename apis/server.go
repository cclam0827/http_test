package apis

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appApi "http_test/apis/app"
	mid "http_test/core/api/middlewares"
	"http_test/core/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	apiConfig config.APIConfig
}

// CreateGinServer builds the engine with access logging, panic recovery and CORS
// installed. Logging comes first so every request, CORS preflights included, is logged.
func CreateGinServer(allowOrigins []string, log *slog.Logger) *gin.Engine {
	if config.GetConfig().App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(
		mid.AccessLogger(log),
		mid.Recoverer(log),
		mid.CORS(allowOrigins),
	)
	return r
}

// StartAPI serves until it receives SIGINT/SIGTERM, then drains in-flight
// requests (up to 10s) before returning.
func (s *Server) StartAPI() error {
	log := slog.Default()
	engine := CreateGinServer(s.apiConfig.AllowedOriginsConfig.App, log)
	appApi.WithRouter(engine)

	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", s.apiConfig.App),
		Handler:           engine.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err // the listener failed, for example because the port is in use
	case <-ctx.Done():
	}
	stop() // a second signal now terminates the process immediately

	log.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Info("server stopped")
	return nil
}

func NewServer() *Server {
	return &Server{
		apiConfig: config.GetConfig().API,
	}
}
