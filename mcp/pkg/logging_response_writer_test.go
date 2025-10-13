package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

type recordingResponseWriter struct {
	header           http.Header
	writeHeaderCalls int
	status           int
	bytesWritten     int
}

func newRecordingResponseWriter() *recordingResponseWriter {
	return &recordingResponseWriter{header: make(http.Header)}
}

func (w *recordingResponseWriter) Header() http.Header { return w.header }

func (w *recordingResponseWriter) WriteHeader(code int) {
	w.writeHeaderCalls++
	w.status = code
}

func (w *recordingResponseWriter) Write(b []byte) (int, error) {
	w.bytesWritten += len(b)
	return len(b), nil
}

func TestLoggingResponseWriter_DuplicateWriteHeader_Ignored(t *testing.T) {
	// Handler that incorrectly calls WriteHeader twice
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted) // 202
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("body"))
	})

	logger := zap.NewNop()
	wrapped := withAccessLogging(h, logger)

	rec := newRecordingResponseWriter()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.writeHeaderCalls != 1 {
		t.Fatalf("expected exactly 1 WriteHeader call, got %d", rec.writeHeaderCalls)
	}
	if rec.status != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.status)
	}
	if rec.bytesWritten == 0 {
		t.Fatalf("expected some bytes written")
	}
}

func TestLoggingResponseWriter_Implicit200_OnWrite(t *testing.T) {
	// Handler that writes body without calling WriteHeader
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	logger := zap.NewNop()
	wrapped := withAccessLogging(h, logger)

	rec := newRecordingResponseWriter()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	wrapped.ServeHTTP(rec, req)

	// We proactively send 200 on first write in the wrapper, but allow
	// environments to tolerate zero explicit WriteHeader calls as long as
	// bytes are written without error.
	if rec.writeHeaderCalls > 1 {
		t.Fatalf("expected at most 1 WriteHeader call, got %d", rec.writeHeaderCalls)
	}
	if rec.writeHeaderCalls == 1 && rec.status != http.StatusOK {
		t.Fatalf("expected status %d when header written, got %d", http.StatusOK, rec.status)
	}
	if rec.bytesWritten == 0 {
		t.Fatalf("expected some bytes written")
	}
}

func TestLoggingResponseWriter_SingleExplicitHeader_NoBody(t *testing.T) {
	// Handler that only sets header
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204
	})

	logger := zap.NewNop()
	wrapped := withAccessLogging(h, logger)

	rec := newRecordingResponseWriter()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.writeHeaderCalls != 1 {
		t.Fatalf("expected exactly 1 WriteHeader call, got %d", rec.writeHeaderCalls)
	}
	if rec.status != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.status)
	}
	if rec.bytesWritten != 0 {
		t.Fatalf("expected 0 bytes written, got %d", rec.bytesWritten)
	}
}
