package controllers

import (
	"http_test/core/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RootController struct {
	hostname string
}

type RequestResponse struct {
	Message        string              `json:"message"`
	ServerHostname string              `json:"server_hostname"`
	Method         string              `json:"method"`
	Path           string              `json:"path"`
	URI            string              `json:"uri"`
	Headers        map[string][]string `json:"headers"`
	Body           interface{}         `json:"body"`
	Query          map[string][]string `json:"query"`
	Params         map[string]string   `json:"params"`
	Form           map[string][]string `json:"form"`
	CurrentTime    time.Time           `json:"current_time"`
}

// Default godoc
// @Summary show server information
// @Schemes
// @Description Returns server information and request details
// @Tags default
// @Accept json
// @Produce json
// @Success 200 {object} RequestResponse
// @Router / [get]
func (ctr *RootController) Root(c *gin.Context) {
	response := &RequestResponse{
		Message:        "Welcome to the HTTP TEST API",
		ServerHostname: ctr.hostname,
		Method:         c.Request.Method,
		Path:           c.Request.URL.Path,
		URI:            c.Request.RequestURI,
		Headers:        c.Request.Header,
		Body:           c.Request.Body,
		Query:          c.Request.URL.Query(),
		Params:         map[string]string{},
		Form:           c.Request.Form,
		CurrentTime:    time.Now(), // Add current time field
	}
	c.IndentedJSON(http.StatusOK, response)
}

type InfoResponse struct {
	Message  string `json:"message"`
	Hostname string `json:"hostname"`
}

// Default godoc
// @Summary show server information
// @Schemes
// @Description Returns server information
// @Tags default
// @Accept json
// @Produce json
// @Success 200 {object} InfoResponse
// @Router /info [get]
func (ctr *RootController) Info(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"message":  "Welcome to the HTTP TEST API",
		"hostname": ctr.hostname,
	})
}

func GetRootController() *RootController {
	return &RootController{
		hostname: config.GetConfig().ServerInfo.Hostname,
	}
}
