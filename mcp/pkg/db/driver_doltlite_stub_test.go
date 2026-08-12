//go:build !doltlite

package db

import (
	"errors"
	"testing"
)

func TestPrepareDatabaseReportsMissingDoltLiteBuild(t *testing.T) {
	err := PrepareDatabase(Config{DialectType: DialectDoltLite, Path: ":memory:"})
	if !errors.Is(err, ErrDoltLiteNotSupported) {
		t.Fatalf("expected ErrDoltLiteNotSupported, got %v", err)
	}
}
