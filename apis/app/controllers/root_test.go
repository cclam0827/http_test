package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewRootController("test-host")
	r := gin.New()
	r.GET("/info", ctrl.Info)
	r.NoRoute(ctrl.Root)
	return r
}

// do sends a request through the engine and returns the decoded JSON response.
func do(t *testing.T, method, target, contentType, body string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	newEngine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not JSON: %v\n%s", err, rec.Body.String())
	}
	return got
}

func TestInfo(t *testing.T) {
	got := do(t, http.MethodGet, "/info", "", "")
	if got["hostname"] != "test-host" || got["message"] != "Welcome to the HTTP TEST API" {
		t.Fatalf("unexpected /info response: %v", got)
	}
}

func TestRootEchoesRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/a/b?x=1&x=2", nil)
	req.Header.Set("X-Test", "v")
	rec := httptest.NewRecorder()
	newEngine().ServeHTTP(rec, req)

	var got RequestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Method != "PUT" || got.Path != "/a/b" || got.URI != "/a/b?x=1&x=2" || got.ServerHostname != "test-host" {
		t.Errorf("request line not echoed: %+v", got)
	}
	if q := got.Query["x"]; len(q) != 2 || q[0] != "1" || q[1] != "2" {
		t.Errorf("query = %v", got.Query)
	}
	if h := got.Headers["X-Test"]; len(h) != 1 || h[0] != "v" {
		t.Errorf("headers = %v", got.Headers)
	}
	if got.CurrentTime.IsZero() {
		t.Error("current_time is missing")
	}
}

func TestRootBody(t *testing.T) {
	tests := []struct {
		name, method, contentType, body string
		want                            any
	}{
		{"no body keeps the historical {}", "GET", "", "", map[string]any{}},
		{"json is embedded", "POST", "application/json", `{"a":1}`, map[string]any{"a": float64(1)}},
		{"json suffix and charset", "POST", "application/vnd.api+json; charset=utf-8", `[1,2]`, []any{float64(1), float64(2)}},
		{"invalid json stays a string", "POST", "application/json", `{nope`, `{nope`},
		{"json-looking text stays a string", "POST", "text/plain", `{"a":1}`, `{"a":1}`},
		{"plain text", "POST", "text/plain", "hello", "hello"},
		{"urlencoded body is echoed raw too", "POST", "application/x-www-form-urlencoded", "a=1", "a=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := do(t, tt.method, "/echo", tt.contentType, tt.body)["body"]
			if gj, wj := mustJSON(t, got), mustJSON(t, tt.want); gj != wj {
				t.Errorf("body = %s, want %s", gj, wj)
			}
		})
	}
}

func TestRootForm(t *testing.T) {
	got := do(t, "POST", "/form", "application/x-www-form-urlencoded", "a=1&a=2&b=x")["form"]
	if j := mustJSON(t, got); j != `{"a":["1","2"],"b":["x"]}` {
		t.Errorf("form = %s", j)
	}
	for _, c := range []struct{ name, method, ct string }{
		{"GET never parses a body", "GET", "application/x-www-form-urlencoded"},
		{"other content types are not forms", "POST", "text/plain"},
	} {
		if f := do(t, c.method, "/form", c.ct, "a=1")["form"]; f != nil {
			t.Errorf("%s: form = %v, want null", c.name, f)
		}
	}
}

func TestRootBodyLimit(t *testing.T) {
	exact := do(t, "POST", "/big", "text/plain", strings.Repeat("a", maxEchoBodyBytes))
	if s := exact["body"].(string); len(s) != maxEchoBodyBytes {
		t.Errorf("exact-limit body length = %d", len(s))
	}
	if _, ok := exact["body_truncated"]; ok {
		t.Error("a body exactly at the limit must not be reported as truncated")
	}

	over := do(t, "POST", "/big", "text/plain", strings.Repeat("a", maxEchoBodyBytes+10))
	if s := over["body"].(string); len(s) != maxEchoBodyBytes {
		t.Errorf("truncated body length = %d, want %d", len(s), maxEchoBodyBytes)
	}
	if over["body_truncated"] != true {
		t.Errorf("body_truncated = %v, want true", over["body_truncated"])
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
