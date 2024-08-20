package router

import (
	"fmt"
	ctrls "http_test/apis/app/controllers"
	"http_test/core/config"
	"http_test/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func WithRouter(r *gin.Engine) {
	grp := r.Group("")

	rootGroup(grp)
}

func rootGroup(r *gin.RouterGroup) {
	apiConfig := config.GetConfig().API
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", "localhost", apiConfig.App)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler,
		// ginSwagger.InstanceName("swagger"), // default
		// ginSwagger.URL("http://localhost:5000/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	statusCtrl := ctrls.GetRootController()
	r.GET("/", statusCtrl.Root)
	r.PUT("/", statusCtrl.Root)
	r.POST("/", statusCtrl.Root)
	r.DELETE("/", statusCtrl.Root)
	r.GET("/info", statusCtrl.Info)
	r.PUT("/info", statusCtrl.Info)
	r.POST("/info", statusCtrl.Info)
	r.DELETE("/info", statusCtrl.Info)
}
