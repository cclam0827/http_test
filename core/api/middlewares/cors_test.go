package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// corsEngine serves GET and OPTIONS on /x with a handler that records whether it ran.
func corsEngine(allowOrigins []string, ran *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(allowOrigins))
	h := func(c *gin.Context) { *ran = true; c.String(http.StatusOK, "ok") }
	r.GET("/x", h)
	r.OPTIONS("/x", h)
	return r
}

func send(r *gin.Engine, method string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/x", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCORSPreflightReflectsTheRequest(t *testing.T) {
	var ran bool
	rec := send(corsEngine([]string{"*"}, &ran), http.MethodOptions, map[string]string{
		"Origin":                         "https://a.example",
		"Access-Control-Request-Method":  "PATCH",
		"Access-Control-Request-Headers": "authorization, X-Anything-At-All",
	})

	if rec.Code != http.StatusNoContent || ran {
		t.Fatalf("preflight must be answered with 204 without reaching the route, got %d (route ran: %v)", rec.Code, ran)
	}
	want := map[string]string{
		"Access-Control-Allow-Origin":      "https://a.example", // never "*": it fails with credentials
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "PATCH",
		"Access-Control-Allow-Headers":     "authorization, X-Anything-At-All", // verbatim, Authorization included
		"Access-Control-Max-Age":           "43200",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	vary := strings.Join(rec.Header().Values("Vary"), ",")
	for _, v := range []string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"} {
		if !strings.Contains(vary, v) {
			t.Errorf("Vary = %q, missing %s", vary, v)
		}
	}
}

func TestCORSPreflightWithoutSpecifics(t *testing.T) {
	var ran bool
	rec := send(corsEngine([]string{"*"}, &ran), http.MethodOptions, map[string]string{"Origin": "https://a.example"})

	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "PATCH") || !strings.Contains(got, "DELETE") {
		t.Errorf("default Allow-Methods = %q, want the full method list including PATCH", got)
	}
	if _, has := rec.Header()["Access-Control-Allow-Headers"]; has {
		t.Error("Allow-Headers must not be sent when no headers were requested")
	}
}

func TestCORSActualRequest(t *testing.T) {
	var ran bool
	rec := send(corsEngine([]string{"*"}, &ran), http.MethodGet, map[string]string{"Origin": "https://a.example"})

	if !ran || rec.Code != http.StatusOK {
		t.Fatalf("the route must run for a normal request, got %d (ran: %v)", rec.Code, ran)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://a.example" ||
		rec.Header().Get("Access-Control-Allow-Credentials") != "true" ||
		rec.Header().Get("Access-Control-Expose-Headers") != "Content-Length" ||
		!strings.Contains(strings.Join(rec.Header().Values("Vary"), ","), "Origin") {
		t.Errorf("missing CORS headers on the response: %v", rec.Header())
	}
}

// With "*" configured nothing is hard-coded: whatever Origin the browser sends is echoed back.
func TestCORSAnyOriginIsReflected(t *testing.T) {
	for _, origin := range []string{
		"https://a.example",
		"http://localhost:3000",
		"http://192.168.1.10:8080",
		"https://sub.domain.example:8443",
		"null", // sandboxed iframes and file:// pages
	} {
		var ran bool
		rec := send(corsEngine([]string{"*"}, &ran), http.MethodGet, map[string]string{"Origin": origin})
		if !ran || rec.Header().Get("Access-Control-Allow-Origin") != origin || rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("origin %q was not reflected: %d %v", origin, rec.Code, rec.Header())
		}
	}
}

func TestCORSIgnoresRequestsWithoutOrigin(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodOptions} {
		var ran bool
		rec := send(corsEngine([]string{"*"}, &ran), method, nil)
		if !ran || rec.Code != http.StatusOK {
			t.Errorf("%s without Origin must reach the route, got %d (ran: %v)", method, rec.Code, ran)
		}
		for k := range rec.Header() {
			if strings.HasPrefix(k, "Access-Control-") {
				t.Errorf("%s without Origin got a CORS header: %s", method, k)
			}
		}
	}
}

func TestCORSOriginAllowList(t *testing.T) {
	var ran bool
	r := corsEngine([]string{"https://ok.example"}, &ran)

	if rec := send(r, http.MethodGet, map[string]string{"Origin": "https://ok.example"}); !ran ||
		rec.Header().Get("Access-Control-Allow-Origin") != "https://ok.example" {
		t.Errorf("an allowed origin must be accepted: %d %v", rec.Code, rec.Header())
	}

	ran = false
	rec := send(r, http.MethodGet, map[string]string{"Origin": "https://evil.example"})
	if rec.Code != http.StatusForbidden || ran || rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("any other origin must get a bare 403: %d (route ran: %v) %v", rec.Code, ran, rec.Header())
	}
}
