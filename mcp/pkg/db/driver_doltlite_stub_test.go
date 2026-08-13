//go:build !doltlite

package db

import (
	"errors"
	"testing"
)

func TestPrepareDatabaseReportsMissingDoltLiteBuild(t *testing.T) {
	config := Config{DialectType: DialectDoltLite, Path: ":memory:"}
	err := PrepareDatabase(&config)
	if !errors.Is(err, ErrDoltLiteNotSupported) {
		t.Fatalf("expected ErrDoltLiteNotSupported, got %v", err)
	}
}
