package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

// TestWithBearerAuth_InvalidToken_DoesNotCallNext demonstrates that when JWT
// token validation fails, the auth middleware must respond 401 AND stop
// processing the request. Prior to the fix, the middleware wrote the 401 but
// forgot to return, so the wrapped handler ran anyway — a full auth bypass.
func TestWithBearerAuth_InvalidToken_DoesNotCallNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	logger := zap.NewNop()
	// The URL is never fetched: an unparseable token fails before any JWKS
	// lookup, so this test is deterministic and makes no network calls.
	claims := map[string]string{"iss": "test-issuer"}
	handler, err := withBearerAuth(logger, next, claims, "http://127.0.0.1:1/jwks")
	if err != nil {
		t.Fatalf("unexpected error building bearer auth handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if nextCalled {
		t.Fatalf("wrapped handler was called despite failed authentication; the request was not rejected")
	}
}

// TestWithBearerAuth_MissingToken_DoesNotCallNext is a sanity check that the
// no-token path (which already returned correctly) still rejects the request.
func TestWithBearerAuth_MissingToken_DoesNotCallNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	logger := zap.NewNop()
	claims := map[string]string{"iss": "test-issuer"}
	handler, err := withBearerAuth(logger, next, claims, "http://127.0.0.1:1/jwks")
	if err != nil {
		t.Fatalf("unexpected error building bearer auth handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if nextCalled {
		t.Fatalf("wrapped handler was called despite missing authentication")
	}
}
