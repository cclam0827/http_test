package controllers

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// maxEchoBodyBytes caps how much of a request body is buffered and echoed back.
const maxEchoBodyBytes = 1 << 20 // 1 MiB

type RootController struct {
	hostname string
}

func NewRootController(hostname string) *RootController {
	return &RootController{hostname: hostname}
}

type RequestResponse struct {
	Message        string              `json:"message"`
	ServerHostname string              `json:"server_hostname"`
	Method         string              `json:"method"`
	Path           string              `json:"path"`
	URI            string              `json:"uri"`
	Headers        map[string]string   `json:"headers"`
	Body           any                 `json:"body"`
	BodyTruncated  bool                `json:"body_truncated,omitempty"`
	Query          map[string][]string `json:"query"`
	Params         map[string]string   `json:"params"`
	Form           map[string][]string `json:"form"`
	CurrentTime    time.Time           `json:"current_time"`
}

// Root echoes the request back: method, path, headers, query, body and form.
// It also serves every path and method that has no dedicated route.
func (ctr *RootController) Root(c *gin.Context) {
	raw, truncated, err := readBody(c.Request)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "failed to read request body: " + err.Error(),
		})
		return
	}
	mt := mediaType(c.Request)

	c.IndentedJSON(http.StatusOK, &RequestResponse{
		Message:        "Welcome to the HTTP TEST API",
		ServerHostname: ctr.hostname,
		Method:         c.Request.Method,
		Path:           c.Request.URL.Path,
		URI:            c.Request.RequestURI,
		Headers:        echoHeaders(c.Request),
		Body:           echoBody(mt, raw),
		BodyTruncated:  truncated,
		Query:          c.Request.URL.Query(),
		Params:         map[string]string{},
		Form:           parseForm(c.Request.Method, mt, raw),
		CurrentTime:    time.Now(),
	})
}

// Info returns the server hostname.
func (ctr *RootController) Info(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"message":  "Welcome to the HTTP TEST API",
		"hostname": ctr.hostname,
	})
}

// echoHeaders returns the request headers as httpbin shows them: one string per
// name, with a header sent more than once joined by ", " (the standard way to
// combine repeated fields). Go keeps Host in Request.Host and not in
// Request.Header, so it is added back; it is left out when the request has none.
func echoHeaders(r *http.Request) map[string]string {
	out := make(map[string]string, len(r.Header)+1)
	for name, values := range r.Header {
		out[name] = strings.Join(values, ", ")
	}
	if r.Host != "" {
		out["Host"] = r.Host
	}
	return out
}

// readBody reads at most maxEchoBodyBytes of the request body.
func readBody(r *http.Request) (raw []byte, truncated bool, err error) {
	raw, err = io.ReadAll(io.LimitReader(r.Body, maxEchoBodyBytes+1))
	if err != nil {
		return nil, false, err
	}
	if len(raw) > maxEchoBodyBytes {
		return raw[:maxEchoBodyBytes], true, nil
	}
	return raw, false, nil
}

// echoBody returns the body in a form that encodes sensibly as JSON: a valid
// JSON payload is embedded as-is, anything else becomes a string, and an empty
// body stays `{}` as it always has been.
func echoBody(mt string, raw []byte) any {
	switch {
	case len(raw) == 0:
		return struct{}{}
	case (mt == "application/json" || strings.HasSuffix(mt, "+json")) && json.Valid(raw):
		return json.RawMessage(raw)
	default:
		return string(raw)
	}
}

// parseForm returns the fields of a urlencoded body, or nil when there are none.
func parseForm(method, mt string, raw []byte) map[string][]string {
	if mt != "application/x-www-form-urlencoded" {
		return nil
	}
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return nil
	}
	vals, _ := url.ParseQuery(string(raw)) // malformed pairs are skipped, the rest is kept
	if len(vals) == 0 {
		return nil
	}
	return vals
}

func mediaType(r *http.Request) string {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return ""
	}
	return mt
}
