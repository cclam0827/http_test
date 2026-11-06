package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http_test/apis"
	appApi "http_test/apis/app"
	"http_test/core/config"
)

// These tests wire the real router, middleware and CORS setup together. They
// live in package main so the config package finds ./etc/app/config.json
// relative to the repository root when it loads on import.

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := apis.CreateGinServer(config.GetConfig().API.AllowedOriginsConfig.App, log)
	appApi.WithRouter(engine)
	srv := httptest.NewServer(engine.Handler())
	t.Cleanup(srv.Close)
	return srv
}

func request(t *testing.T, method, url string, header http.Header, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	maps.Copy(req.Header, header)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	return m
}

func TestInfoRoutes(t *testing.T) {
	srv := newTestServer(t)
	for _, method := range []string{"GET", "PUT", "POST", "DELETE"} {
		resp := request(t, method, srv.URL+"/info", nil, "")
		got := decode(t, resp)
		if resp.StatusCode != 200 || got["hostname"] == "" || got["message"] == nil {
			t.Errorf("%s /info: status %d body %v", method, resp.StatusCode, got)
		}
	}
}

func TestCatchAllEchoesEverythingElse(t *testing.T) {
	srv := newTestServer(t)
	for _, tc := range []struct{ method, path string }{
		{"GET", "/"},
		{"GET", "/anything/at/all?x=1"},
		{"PATCH", "/info"}, // /info exists, but not for PATCH
		{"POST", "/swagger/index.html"},
	} {
		resp := request(t, tc.method, srv.URL+tc.path, nil, "")
		got := decode(t, resp)
		if resp.StatusCode != 200 || got["method"] != tc.method {
			t.Errorf("%s %s: status %d, echoed method %v", tc.method, tc.path, resp.StatusCode, got["method"])
		}
	}
}

func TestEchoPostedJSON(t *testing.T) {
	srv := newTestServer(t)
	resp := request(t, "POST", srv.URL+"/anything", http.Header{"Content-Type": {"application/json"}}, `{"hello":"world"}`)
	body, _ := decode(t, resp)["body"].(map[string]any)
	if body["hello"] != "world" {
		t.Errorf("posted JSON was not echoed: %v", body)
	}
}

// The middleware's own tests cover the CORS rules; these check it is wired in with the shipped config.
func TestCORSPreflight(t *testing.T) {
	srv := newTestServer(t)
	resp := request(t, "OPTIONS", srv.URL+"/info", http.Header{
		"Origin":                         {"http://example.com"},
		"Access-Control-Request-Method":  {"PATCH"},
		"Access-Control-Request-Headers": {"authorization, x-anything"},
	}, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", resp.StatusCode)
	}
	want := map[string]string{
		"Access-Control-Allow-Origin":      "http://example.com",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "PATCH",
		"Access-Control-Allow-Headers":     "authorization, x-anything",
		"Access-Control-Max-Age":           "43200",
	}
	for k, v := range want {
		if got := resp.Header.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestCORSHeadersOnlyWithOrigin(t *testing.T) {
	srv := newTestServer(t)

	with := request(t, "GET", srv.URL+"/info", http.Header{"Origin": {"http://example.com"}}, "")
	if with.Header.Get("Access-Control-Allow-Origin") != "http://example.com" || with.Header.Get("Access-Control-Expose-Headers") != "Content-Length" {
		t.Errorf("missing CORS headers on a cross-origin request: %v", with.Header)
	}

	without := request(t, "GET", srv.URL+"/info", nil, "")
	if v := without.Header.Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("CORS header leaked onto a same-origin request: %q", v)
	}
}
