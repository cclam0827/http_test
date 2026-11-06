package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

// colorHandler prints one colorized line per record, meant for local development.
type colorHandler struct {
	mu     *sync.Mutex // shared by every handler derived from this one
	out    io.Writer
	level  slog.Level
	attrs  []slog.Attr // from WithAttrs, keys already group-qualified
	groups string      // "a.b." prefix built by WithGroup
}

func newColorHandler(out io.Writer, level slog.Level) *colorHandler {
	return &colorHandler{mu: &sync.Mutex{}, out: out, level: level}
}

func (h *colorHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *colorHandler) Handle(_ context.Context, r slog.Record) error {
	b := make([]byte, 0, 256)
	b = append(b, colorGray...)
	b = r.Time.AppendFormat(b, "2006/01/02 15:04:05")
	b = append(b, colorReset...)
	b = append(b, ' ')
	b = append(b, levelColor(r.Level)...)
	b = append(b, r.Level.String()...)
	b = append(b, colorReset...)
	b = append(b, ' ')
	b = append(b, r.Message...)
	for _, a := range h.attrs {
		b = appendAttr(b, "", a)
	}
	r.Attrs(func(a slog.Attr) bool {
		b = appendAttr(b, h.groups, a)
		return true
	})
	b = append(b, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(b)
	return err
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := *h
	qualified := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		a.Key = h.groups + a.Key
		qualified[i] = a
	}
	c.attrs = slices.Concat(h.attrs, qualified) // new backing array: never shared with h
	return &c
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	c := *h
	c.groups = h.groups + name + "."
	return &c
}

// appendAttr writes " key=value", flattening nested groups as "group.key".
func appendAttr(b []byte, prefix string, a slog.Attr) []byte {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return b
	}
	if a.Value.Kind() == slog.KindGroup {
		if a.Key != "" {
			prefix += a.Key + "."
		}
		for _, ga := range a.Value.Group() {
			b = appendAttr(b, prefix, ga)
		}
		return b
	}

	b = append(b, ' ')
	b = append(b, prefix...)
	b = append(b, a.Key...)
	b = append(b, '=')

	switch a.Key {
	case "method":
		b = append(b, colorCyan...)
	case "status":
		b = append(b, statusColor(a.Value)...)
	case "duration":
		b = append(b, colorYellow...)
	default:
		return appendValue(b, a.Value)
	}
	b = appendValue(b, a.Value)
	return append(b, colorReset...)
}

func appendValue(b []byte, v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		if s := v.String(); needsQuote(s) {
			return strconv.AppendQuote(b, s)
		} else {
			return append(b, s...)
		}
	case slog.KindInt64:
		return strconv.AppendInt(b, v.Int64(), 10)
	case slog.KindUint64:
		return strconv.AppendUint(b, v.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.AppendFloat(b, v.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.AppendBool(b, v.Bool())
	case slog.KindDuration:
		return append(b, v.Duration().String()...)
	case slog.KindTime:
		return v.Time().AppendFormat(b, time.RFC3339)
	default:
		return strconv.AppendQuote(b, fmt.Sprint(v.Any()))
	}
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			return true
		}
	}
	return false
}

func levelColor(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return colorRed
	case l >= slog.LevelWarn:
		return colorYellow
	case l >= slog.LevelInfo:
		return colorGreen
	default:
		return colorGray
	}
}

func statusColor(v slog.Value) string {
	if v.Kind() != slog.KindInt64 {
		return ""
	}
	switch code := v.Int64(); {
	case code >= 500:
		return colorRed
	case code >= 400:
		return colorYellow
	case code >= 300:
		return colorCyan
	default:
		return colorGreen
	}
}
