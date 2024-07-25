package controllers

import (
	"http_test/core/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RootController struct {
	hostname string
}

func (ctr *RootController) Root(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"message":         "Welcome to the HTTP TEST API",
		"server_hostname": ctr.hostname,
		"method":          c.Request.Method,
		"path":            c.Request.URL.Path,
		"uri":             c.Request.RequestURI,
		"headers":         c.Request.Header,
		"body":            c.Request.Body,
		"query":           c.Request.URL.Query(),
		"params":          c.Params,
		"form":            c.Request.Form,
	})
}

func GetRootController() *RootController {
	return &RootController{
		hostname: config.GetConfig().ServerInfo.Hostname,
	}
}
