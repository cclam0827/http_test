package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewHandlerFormats(t *testing.T) {
	tests := []struct {
		format string
		check  func(out string) bool
	}{
		{"json", func(o string) bool { return strings.HasPrefix(o, `{"time":`) && strings.Contains(o, `"msg":"hello"`) }},
		{"text", func(o string) bool { return strings.Contains(o, "msg=hello") && !strings.Contains(o, "\x1b[") }},
		{"color", func(o string) bool { return strings.Contains(o, "\x1b[") && strings.Contains(o, "hello k=v") }},
		{"nonsense", func(o string) bool { return strings.HasPrefix(o, "{") }}, // falls back to json
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			var buf bytes.Buffer
			slog.New(NewHandler(&buf, tt.format, slog.LevelInfo)).Info("hello", "k", "v")
			if !tt.check(buf.String()) {
				t.Errorf("unexpected %s output: %q", tt.format, buf.String())
			}
		})
	}
}

func TestLevels(t *testing.T) {
	if ParseLevel("DEBUG") != slog.LevelDebug || ParseLevel("info") != slog.LevelInfo || ParseLevel("") != slog.LevelInfo {
		t.Error("ParseLevel mapping is wrong")
	}
	for _, format := range []string{"json", "text", "color"} {
		var buf bytes.Buffer
		l := slog.New(NewHandler(&buf, format, slog.LevelInfo))
		l.Debug("hidden")
		if buf.Len() != 0 {
			t.Errorf("%s: debug record leaked at info level: %q", format, buf.String())
		}
		slog.New(NewHandler(&buf, format, slog.LevelDebug)).Debug("shown")
		if !strings.Contains(buf.String(), "shown") {
			t.Errorf("%s: debug record missing at debug level", format)
		}
	}
}

// Handlers derived from the same parent must not share attribute storage.
func TestColorHandlerDerivedLoggersDoNotAlias(t *testing.T) {
	var buf bytes.Buffer
	// Three chained With calls leave a naive append-based implementation with spare capacity.
	base := slog.New(newColorHandler(&buf, slog.LevelInfo)).With("a", 1).With("b", 2).With("c", 3)
	x := base.With("who", "x")
	y := base.With("who", "y")
	x.Info("m")
	y.Info("m")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "who=x") || strings.Contains(lines[0], "who=y") ||
		!strings.Contains(lines[1], "who=y") || strings.Contains(lines[1], "who=x") {
		t.Errorf("derived loggers interfered with each other:\n%s", buf.String())
	}
	for _, l := range lines {
		if !strings.Contains(l, "a=1 b=2 c=3") {
			t.Errorf("inherited attributes lost: %q", l)
		}
	}
}

func TestColorHandlerGroups(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(newColorHandler(&buf, slog.LevelInfo))
	l.WithGroup("g").With("x", 1).Info("m", "k", 2, slog.Group("n", "z", 3))
	out := buf.String()
	for _, want := range []string{"g.x=1", "g.k=2", "g.n.z=3"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}
