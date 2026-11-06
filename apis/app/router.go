package router

import (
	ctrls "http_test/apis/app/controllers"

	"github.com/gin-gonic/gin"
)

func WithRouter(r *gin.Engine) {
	grp := r.Group("")

	rootGroup(grp)
	r.NoRoute(ctrls.GetRootController().Root)
}

func rootGroup(r *gin.RouterGroup) {
	statusCtrl := ctrls.GetRootController()
	r.GET("/info", statusCtrl.Info)
	r.PUT("/info", statusCtrl.Info)
	r.POST("/info", statusCtrl.Info)
	r.DELETE("/info", statusCtrl.Info)
}
