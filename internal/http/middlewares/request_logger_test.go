package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLoggerWritesErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	r := gin.New()
	r.Use(TraceID(), RequestLogger())
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/conflict", func(c *gin.Context) {
		_ = c.Error(errors.New("недопустимый переход статуса"))
		c.Status(http.StatusConflict)
	})
	r.GET("/boom", func(c *gin.Context) {
		_ = c.Error(errors.New("db down"))
		c.Status(http.StatusInternalServerError)
	})

	cases := []struct {
		path, level, err string
	}{
		{"/ok", "INFO", ""},
		{"/conflict", "WARN", "недопустимый переход статуса"},
		{"/boom", "ERROR", "db down"},
	}
	for _, tc := range cases {
		buf.Reset()
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, tc.path, nil))

		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("%s: bad log line %q: %v", tc.path, buf.String(), err)
		}
		if rec["level"] != tc.level {
			t.Errorf("%s: level = %v, want %s", tc.path, rec["level"], tc.level)
		}
		if got, _ := rec["err"].(string); got != tc.err {
			t.Errorf("%s: err = %q, want %q", tc.path, got, tc.err)
		}
		if rec["trace_id"] == "" || rec["trace_id"] == nil {
			t.Errorf("%s: no trace_id", tc.path)
		}
	}
}
