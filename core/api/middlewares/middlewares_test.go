package middlewares

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// records decodes the JSON log lines written to buf.
func records(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	sc := bufio.NewScanner(buf)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("log line is not JSON: %v: %s", err, sc.Text())
		}
		out = append(out, m)
	}
	return out
}

func serve(r *gin.Engine, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestAccessLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	r := gin.New()
	r.Use(AccessLogger(slog.New(slog.NewJSONHandler(&buf, nil))))
	r.GET("/hello", func(c *gin.Context) { c.String(http.StatusCreated, "hi") })
	r.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("boom"))
		c.Status(http.StatusInternalServerError)
	})

	serve(r, "GET", "/hello?secret=1")
	serve(r, "GET", "/fail")

	recs := records(t, &buf)
	if len(recs) != 2 {
		t.Fatalf("got %d log lines, want one per request: %v", len(recs), recs)
	}
	ok := recs[0]
	if ok["msg"] != "http_request" || ok["method"] != "GET" || ok["path"] != "/hello" ||
		ok["status"] != float64(201) || ok["bytes"] != float64(2) || ok["client_ip"] != "192.0.2.1" {
		t.Errorf("unexpected access log: %v", ok)
	}
	if _, has := ok["error"]; has {
		t.Error("error attribute must be absent on a clean request")
	}
	if _, has := ok["duration"]; !has {
		t.Error("duration is missing")
	}
	if got := recs[1]["error"]; got == nil || recs[1]["status"] != float64(500) {
		t.Errorf("failing request should log status 500 and its error, got %v", recs[1])
	}
}

func TestRecoverer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	r := gin.New()
	r.Use(AccessLogger(log), Recoverer(log))
	r.GET("/panic", func(c *gin.Context) { panic("kaboom") })

	rec := serve(r, "GET", "/panic")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	recs := records(t, &buf)
	if len(recs) != 2 {
		t.Fatalf("want a panic record and an access record, got %v", recs)
	}
	if recs[0]["msg"] != "panic recovered" || recs[0]["error"] != "kaboom" || recs[0]["level"] != "ERROR" {
		t.Errorf("unexpected panic record: %v", recs[0])
	}
	if s, _ := recs[0]["stack"].(string); s == "" {
		t.Error("panic record has no stack")
	}
	if recs[1]["status"] != float64(500) {
		t.Errorf("access log should see the recovered 500, got %v", recs[1])
	}
}
