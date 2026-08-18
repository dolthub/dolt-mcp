package db

import (
	"errors"
	"testing"
	"time"
)

func TestDoltLiteConfigValidation(t *testing.T) {
	config := Config{DialectType: DialectDoltLite}
	if err := config.Validate(); !errors.Is(err, ErrNoDatabaseFileDefined) {
		t.Fatalf("expected ErrNoDatabaseFileDefined, got %v", err)
	}

	config.Path = "/tmp/test.db"
	if err := config.Validate(); err != nil {
		t.Fatalf("expected valid DoltLite config, got %v", err)
	}

	config = Config{DialectType: DialectDoltLite, DSN: ":memory:"}
	if err := config.Validate(); err != nil {
		t.Fatalf("expected raw DSN to bypass path validation, got %v", err)
	}

	config.BusyTimeout = -time.Millisecond
	if err := config.Validate(); !errors.Is(err, ErrInvalidDoltLiteBusyTimeout) {
		t.Fatalf("expected ErrInvalidDoltLiteBusyTimeout, got %v", err)
	}

	config.BusyTimeout = time.Microsecond
	if err := config.Validate(); !errors.Is(err, ErrInvalidDoltLiteBusyTimeout) {
		t.Fatalf("expected ErrInvalidDoltLiteBusyTimeout, got %v", err)
	}
}
