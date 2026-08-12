package db

import (
	"errors"
	"testing"
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
}
