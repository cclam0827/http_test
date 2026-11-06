package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

const (
	// Allow-Methods answer for a preflight that does not name the method it wants.
	corsDefaultMethods = "GET,PUT,POST,PATCH,DELETE,OPTIONS,HEAD"
	corsMaxAge         = "43200" // 12h
)

// CORS accepts cross-origin requests the way httpbin and httpbun do: it reflects
// the request's Origin, Access-Control-Request-Method and
// Access-Control-Request-Headers, so every origin, method and request header is
// allowed, credentials included. A literal "*" is deliberately avoided: browsers
// reject it on requests that carry credentials, and as Allow-Headers it never
// covers Authorization.
//
// "*" in allowOrigins allows every origin; otherwise only exact matches pass and
// any other Origin gets a 403. A request without an Origin header is not a CORS
// request and passes through untouched.
func CORS(allowOrigins []string) gin.HandlerFunc {
	allowAll := slices.Contains(allowOrigins, "*")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			return
		}
		if !allowAll && !slices.Contains(allowOrigins, origin) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Add("Vary", "Origin")

		if c.Request.Method != http.MethodOptions {
			h.Set("Access-Control-Expose-Headers", "Content-Length")
			return
		}

		// Preflight: answer it here, the routes never see it.
		methods := c.GetHeader("Access-Control-Request-Method")
		if methods == "" {
			methods = corsDefaultMethods
		}
		h.Set("Access-Control-Allow-Methods", methods)
		if requested := c.GetHeader("Access-Control-Request-Headers"); requested != "" {
			h.Set("Access-Control-Allow-Headers", requested)
		}
		h.Set("Access-Control-Max-Age", corsMaxAge)
		h.Add("Vary", "Access-Control-Request-Method")
		h.Add("Vary", "Access-Control-Request-Headers")
		c.AbortWithStatus(http.StatusNoContent)
	}
}
