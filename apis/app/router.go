package router

import (
	ctrls "http_test/apis/app/controllers"
	"http_test/core/config"

	"github.com/gin-gonic/gin"
)

func WithRouter(r *gin.Engine) {
	ctrl := ctrls.NewRootController(config.GetConfig().ServerInfo.Hostname)

	r.GET("/info", ctrl.Info)
	r.PUT("/info", ctrl.Info)
	r.POST("/info", ctrl.Info)
	r.DELETE("/info", ctrl.Info)

	// Every other path and method is echoed back.
	r.NoRoute(ctrl.Root)
}
