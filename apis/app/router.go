package router

import (
	ctrls "http_test/apis/app/controllers"

	"github.com/gin-gonic/gin"
)

func WithRouter(r *gin.Engine) {
	grp := r.Group("")

	rootGroup(grp)
}

func rootGroup(r *gin.RouterGroup) {
	statusCtrl := ctrls.GetRootController()
	r.GET("/", statusCtrl.Root)
	r.PUT("/", statusCtrl.Root)
	r.POST("/", statusCtrl.Root)
	r.DELETE("/", statusCtrl.Root)
}
