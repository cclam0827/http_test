package apis

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	appApi "http_test/apis/app"
	mid "http_test/core/api/middlewares"
	"http_test/core/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

type Server struct {
	apiConfig config.APIConfig
}

// CreateGinServer builds the engine with access logging, panic recovery and CORS
// installed. Logging comes first so every request, CORS preflights included, is logged.
func CreateGinServer(allowOrigins []string, l *slog.Logger) *gin.Engine {
	if config.GetConfig().App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// r.SetTrustedProxies([]string{"0.0.0.0::1"})
	r.Use(
		mid.AccessLogger(l),
		mid.Recoverer(l),
		cors.New(cors.Config{
			// AllowAllOrigins:  true,
			AllowOrigins:     allowOrigins,
			AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Origin", "Content-Type"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	return r
}

// https://gin-gonic.com/docs/examples/run-multiple-service/
func (s *Server) StartAPI() {
	addr := fmt.Sprintf("0.0.0.0:%d", s.apiConfig.App)
	appServer := CreateGinServer(s.apiConfig.AllowedOriginsConfig.App, slog.Default())
	appApi.WithRouter(appServer)

	g.Go(func() error {
		slog.Info("server starting", "addr", addr)
		return appServer.Run(addr)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func NewServer() *Server {
	return &Server{
		apiConfig: config.GetConfig().API,
	}
}
