package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orch8-io/cli/internal/output"
	orch8 "github.com/orch8-io/sdk-go"
)

func TestInstanceListCommand(t *testing.T) {
	instances := []orch8.TaskInstance{
		{ID: "i1", SequenceID: "s1", State: "pending", TenantID: "t1", Namespace: "default", CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z"},
		{ID: "i2", SequenceID: "s2", State: "running", TenantID: "t1", Namespace: "default", CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances" {
			q := r.URL.Query()
			if q.Get("tenant_id") != "t1" {
				t.Errorf("expected tenant_id=t1, got %s", q.Get("tenant_id"))
			}
			if q.Get("limit") != "10" {
				t.Errorf("expected limit=10, got %s", q.Get("limit"))
			}
			json.NewEncoder(w).Encode(instances)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = true

	buf := new(bytes.Buffer)
	output.Out = buf
	defer func() { output.Out = nil }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "list", "--limit", "10"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out []orch8.TaskInstance
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 instances, got %d", len(out))
	}
}

func TestHealthCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health/ready" {
			json.NewEncoder(w).Encode(orch8.HealthResponse{Status: "ok"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = false

	buf := new(bytes.Buffer)
	output.Out = buf
	defer func() { output.Out = nil }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"health"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != "Status: ok\n" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestCommandFlagErrors(t *testing.T) {
	flagURL = "http://localhost"
	flagTenantID = "t1"

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "list", "--limit", "not-a-number"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --limit")
	}
}
