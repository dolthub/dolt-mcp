package main

import (
	"testing"
)

func TestGetTLSConfigWithValidCertAndKey(t *testing.T) {
	certFile := "testdata/server.crt"
	keyFile := "testdata/server.key"

	tlsConfig, err := getTLSConfig(certFile, keyFile, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("expected tlsConfig, got nil")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}
}

func TestGetTLSConfigWithMissingCertOrKey(t *testing.T) {
	_, err := getTLSConfig("", "testdata/server.key", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	_, err = getTLSConfig("testdata/server.crt", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTLSConfigWithInvalidCertOrKey(t *testing.T) {
	_, err := getTLSConfig("invalid.crt", "invalid.key", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTLSConfigWithValidCA(t *testing.T) {
	certFile := "testdata/server.crt"
	keyFile := "testdata/server.key"
	caFile := "testdata/ca.crt"

	tlsConfig, err := getTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tlsConfig.ClientCAs == nil {
		t.Fatal("expected ClientCAs, got nil")
	}
}

func TestParseClaimsMapWithValidInput(t *testing.T) {
	input := "key1=value1,key2=value2"
	expected := map[string]string{"key1": "value1", "key2": "value2"}

	result, err := parseClaimsMap(&input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d claims, got %d", len(expected), len(result))
	}

	for k, v := range expected {
		if result[k] != v {
			t.Fatalf("expected %s=%s, got %s=%s", k, v, k, result[k])
		}
	}
}

func TestParseClaimsMapWithInvalidInput(t *testing.T) {
	input := "key1=value1,key2"

	_, err := parseClaimsMap(&input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseClaimsMapWithEmptyInput(t *testing.T) {
	input := ""

	result, err := parseClaimsMap(&input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestValidateArgsWithValidArgs(t *testing.T) {
	serveHTTP = boolPtr(true)
	mcpPort = intPtr(8080)

	err := validateArgs("localhost", "user", 3306)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateArgsWithMissingHost(t *testing.T) {
	serveHTTP = boolPtr(true)
	mcpPort = intPtr(8080)

	err := validateArgs("", "user", 3306)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateArgsWithMissingUser(t *testing.T) {
	serveHTTP = boolPtr(true)
	mcpPort = intPtr(8080)

	err := validateArgs("localhost", "", 3306)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateArgsWithMissingPort(t *testing.T) {
	serveHTTP = boolPtr(true)
	mcpPort = intPtr(8080)

	err := validateArgs("localhost", "user", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateArgsWithMissingMCPPort(t *testing.T) {
	serveHTTP = boolPtr(true)
	mcpPort = intPtr(0)

	err := validateArgs("localhost", "user", 3306)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
