package cmd

import (
	"bytes"
	"testing"

	orch8 "github.com/orch8-io/sdk-go"
)

func TestNewClient(t *testing.T) {
	flagURL = "http://localhost:8080"
	flagTenantID = "test-tenant"
	flagAPIKey = "secret-key"

	c := newClient()
	if c == nil {
		t.Fatal("expected client, got nil")
	}

	// newClient returns an unexported *orch8.Client; we can only verify it
	// doesn't panic and has the right type via the package boundary.
	var _ *orch8.Client = c
}

func TestEnvOr(t *testing.T) {
	if got := envOr("NONEXISTENT_VAR_XYZ", "fallback"); got != "fallback" {
		t.Errorf("expected fallback, got %q", got)
	}
}

func TestVersionCmd(t *testing.T) {
	version = "0.2.0-test"
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "orch8 0.2.0-test\n" {
		t.Errorf("unexpected output: %q", out)
	}
}
