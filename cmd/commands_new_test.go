package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/orch8-io/cli/internal/output"
	orch8 "github.com/orch8-io/sdk-go"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// setupTest configures flags and output capture for a single test.
// It returns the output buffer and a cleanup function.
func setupTest(srvURL string, jsonMode bool) (*bytes.Buffer, func()) {
	flagURL = srvURL
	flagTenantID = "t1"
	flagJSON = jsonMode
	flagAPIKey = ""

	buf := new(bytes.Buffer)
	output.Out = buf

	return buf, func() { output.Out = nil }
}

// run sets up rootCmd and executes with the given args, writing into buf.
func run(t *testing.T, buf *bytes.Buffer, args ...string) {
	t.Helper()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// captureStdout redirects os.Stdout to a pipe so that bare fmt.Println
// calls can be captured. Returns a function that restores stdout and
// returns everything written.
func captureStdout(t *testing.T) func() string {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	return func() string {
		w.Close()
		captured, _ := io.ReadAll(r)
		r.Close()
		os.Stdout = origStdout
		return string(captured)
	}
}

// writeTempJSON writes data as JSON to a temp file and returns its path.
// The caller must remove the file when done.
func writeTempJSON(t *testing.T, data any) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		t.Fatalf("encoding JSON: %v", err)
	}
	f.Close()
	return f.Name()
}

// decodeBody reads the request body into a generic map.
func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}
	return m
}

// ---------------------------------------------------------------------------
// Sequence commands
// ---------------------------------------------------------------------------

func TestSequenceList(t *testing.T) {
	sequences := []orch8.SequenceDefinition{
		{
			ID:         "s1",
			Name:       "my-seq",
			Namespace:  "default",
			Version:    1,
			Deprecated: false,
			CreatedAt:  "2025-01-01T00:00:00Z",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sequences" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(sequences)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "sequence", "list", "--json")

	var out []orch8.SequenceDefinition
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 sequence, got %d", len(out))
	}
	if out[0].ID != "s1" {
		t.Errorf("expected ID=s1, got %s", out[0].ID)
	}
	if out[0].Name != "my-seq" {
		t.Errorf("expected Name=my-seq, got %s", out[0].Name)
	}
}

func TestSequenceDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sequences/s1" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, false)
	defer cleanup()

	getStdout := captureStdout(t)

	run(t, buf, "sequence", "delete", "s1")

	got := strings.TrimSpace(getStdout())
	if got != "Deleted: s1" {
		t.Errorf("expected %q, got %q", "Deleted: s1", got)
	}
}

func TestSequenceMigrateInstance(t *testing.T) {
	result := orch8.TaskInstance{
		ID:         "i1",
		SequenceID: "s2",
		TenantID:   "t1",
		Namespace:  "default",
		State:      "pending",
		CreatedAt:  "2025-01-01T00:00:00Z",
		UpdatedAt:  "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sequences/migrate-instance" && r.Method == http.MethodPost {
			body := decodeBody(t, r)
			if body["instance_id"] != "i1" {
				t.Errorf("expected instance_id=i1, got %v", body["instance_id"])
			}
			if body["target_sequence_id"] != "s2" {
				t.Errorf("expected target_sequence_id=s2, got %v", body["target_sequence_id"])
			}
			json.NewEncoder(w).Encode(result)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "sequence", "migrate-instance", "--instance", "i1", "--target-sequence", "s2", "--json")

	var out orch8.TaskInstance
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != "i1" {
		t.Errorf("expected ID=i1, got %s", out.ID)
	}
	if out.SequenceID != "s2" {
		t.Errorf("expected SequenceID=s2, got %s", out.SequenceID)
	}
}

// ---------------------------------------------------------------------------
// Rollback policy commands
// ---------------------------------------------------------------------------

func TestRollbackPolicyList(t *testing.T) {
	policies := []orch8.RollbackPolicy{
		{
			ID:                     1,
			TenantID:               "t1",
			SequenceName:           "my-seq",
			ErrorRateThreshold:     0.5,
			TimeWindowSecs:         300,
			Enabled:                true,
			CooldownSecs:           60,
			ConfirmationWindowSecs: 30,
			CreatedAt:              "2025-01-01T00:00:00Z",
			UpdatedAt:              "2025-01-01T00:00:00Z",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rollback-policies" && r.Method == http.MethodGet {
			if q := r.URL.Query().Get("tenant_id"); q != "t1" {
				t.Errorf("expected tenant_id=t1, got %s", q)
			}
			json.NewEncoder(w).Encode(policies)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "rollback-policy", "list", "--json")

	var out []orch8.RollbackPolicy
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(out))
	}
	if out[0].SequenceName != "my-seq" {
		t.Errorf("expected SequenceName=my-seq, got %s", out[0].SequenceName)
	}
	if out[0].ErrorRateThreshold != 0.5 {
		t.Errorf("expected ErrorRateThreshold=0.5, got %f", out[0].ErrorRateThreshold)
	}
}

func TestRollbackPolicyGet(t *testing.T) {
	policy := orch8.RollbackPolicy{
		ID:                     1,
		TenantID:               "t1",
		SequenceName:           "my-seq",
		ErrorRateThreshold:     0.5,
		TimeWindowSecs:         300,
		Enabled:                true,
		CooldownSecs:           60,
		ConfirmationWindowSecs: 30,
		CreatedAt:              "2025-01-01T00:00:00Z",
		UpdatedAt:              "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rollback-policies/my-seq" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(policy)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "rbp", "get", "my-seq", "--json")

	var out orch8.RollbackPolicy
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.SequenceName != "my-seq" {
		t.Errorf("expected SequenceName=my-seq, got %s", out.SequenceName)
	}
	if out.Enabled != true {
		t.Errorf("expected Enabled=true, got %v", out.Enabled)
	}
}

func TestRollbackPolicyCreate(t *testing.T) {
	created := orch8.RollbackPolicy{
		ID:                 1,
		TenantID:           "t1",
		SequenceName:       "my-seq",
		ErrorRateThreshold: 0.5,
		TimeWindowSecs:     300,
		Enabled:            true,
		CreatedAt:          "2025-01-01T00:00:00Z",
		UpdatedAt:          "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rollback-policies" && r.Method == http.MethodPost {
			body := decodeBody(t, r)
			if body["tenant_id"] != "t1" {
				t.Errorf("expected tenant_id=t1, got %v", body["tenant_id"])
			}
			if body["sequence_name"] != "my-seq" {
				t.Errorf("expected sequence_name=my-seq, got %v", body["sequence_name"])
			}
			if body["error_rate_threshold"] != 0.5 {
				t.Errorf("expected error_rate_threshold=0.5, got %v", body["error_rate_threshold"])
			}
			tw, ok := body["time_window_secs"].(float64)
			if !ok || int(tw) != 300 {
				t.Errorf("expected time_window_secs=300, got %v", body["time_window_secs"])
			}
			json.NewEncoder(w).Encode(created)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "rbp", "create", "--sequence", "my-seq", "--threshold", "0.5", "--time-window", "300", "--json")

	var out orch8.RollbackPolicy
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != 1 {
		t.Errorf("expected ID=1, got %d", out.ID)
	}
	if out.SequenceName != "my-seq" {
		t.Errorf("expected SequenceName=my-seq, got %s", out.SequenceName)
	}
}

func TestRollbackPolicyDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rollback-policies/my-seq" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, false)
	defer cleanup()

	getStdout := captureStdout(t)

	run(t, buf, "rbp", "delete", "my-seq")

	got := strings.TrimSpace(getStdout())
	if got != "Deleted: my-seq" {
		t.Errorf("expected %q, got %q", "Deleted: my-seq", got)
	}
}

// ---------------------------------------------------------------------------
// Session commands
// ---------------------------------------------------------------------------

func TestSessionUpdateData(t *testing.T) {
	session := orch8.Session{
		ID:         "sess1",
		TenantID:   "t1",
		SessionKey: "sk1",
		State:      "active",
		CreatedAt:  "2025-01-01T00:00:00Z",
		UpdatedAt:  "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sessions/sess1/data" && r.Method == http.MethodPatch {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			var m map[string]any
			if err := json.Unmarshal(body, &m); err != nil {
				t.Fatalf("decoding body: %v", err)
			}
			if m["key"] != "val" {
				t.Errorf("expected key=val, got %v", m["key"])
			}
			json.NewEncoder(w).Encode(session)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "session", "update-data", "sess1", "--data", `{"key":"val"}`, "--json")

	var out orch8.Session
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != "sess1" {
		t.Errorf("expected ID=sess1, got %s", out.ID)
	}
	if out.State != "active" {
		t.Errorf("expected State=active, got %s", out.State)
	}
}

// ---------------------------------------------------------------------------
// Plugin commands
// ---------------------------------------------------------------------------

func TestPluginUpdate(t *testing.T) {
	plugin := orch8.PluginDef{
		Name:       "my-plugin",
		PluginType: "webhook",
		Source:     "https://example.com/hook",
		TenantID:   "t1",
		Enabled:    false,
		CreatedAt:  "2025-01-01T00:00:00Z",
		UpdatedAt:  "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/plugins/my-plugin" && r.Method == http.MethodPatch {
			body := decodeBody(t, r)
			if body["enabled"] != false {
				t.Errorf("expected enabled=false, got %v", body["enabled"])
			}
			json.NewEncoder(w).Encode(plugin)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile := writeTempJSON(t, map[string]any{"enabled": false})
	defer os.Remove(tmpFile)

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "plugin", "update", "my-plugin", "--file", tmpFile, "--json")

	var out orch8.PluginDef
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.Name != "my-plugin" {
		t.Errorf("expected Name=my-plugin, got %s", out.Name)
	}
	if out.Enabled != false {
		t.Errorf("expected Enabled=false, got %v", out.Enabled)
	}
}

// ---------------------------------------------------------------------------
// Credential commands
// ---------------------------------------------------------------------------

func TestCredentialUpdate(t *testing.T) {
	cred := orch8.Credential{
		ID:             "cred1",
		TenantID:       "t1",
		Name:           "new-name",
		CredentialType: "api_key",
		CreatedAt:      "2025-01-01T00:00:00Z",
		UpdatedAt:      "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/credentials/cred1" && r.Method == http.MethodPatch {
			body := decodeBody(t, r)
			if body["name"] != "new-name" {
				t.Errorf("expected name=new-name, got %v", body["name"])
			}
			json.NewEncoder(w).Encode(cred)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile := writeTempJSON(t, map[string]any{"name": "new-name"})
	defer os.Remove(tmpFile)

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "credential", "update", "cred1", "--file", tmpFile, "--json")

	var out orch8.Credential
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != "cred1" {
		t.Errorf("expected ID=cred1, got %s", out.ID)
	}
	if out.Name != "new-name" {
		t.Errorf("expected Name=new-name, got %s", out.Name)
	}
}

// ---------------------------------------------------------------------------
// Pool resource commands
// ---------------------------------------------------------------------------

func TestPoolAddResource(t *testing.T) {
	resource := orch8.PoolResource{
		ID:          "res1",
		PoolID:      "pool1",
		ResourceKey: "worker-1",
		State:       "available",
		CreatedAt:   "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pools/pool1/resources" && r.Method == http.MethodPost {
			body := decodeBody(t, r)
			if body["resource_key"] != "worker-1" {
				t.Errorf("expected resource_key=worker-1, got %v", body["resource_key"])
			}
			json.NewEncoder(w).Encode(resource)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile := writeTempJSON(t, map[string]any{
		"resource_key": "worker-1",
		"state":        "available",
	})
	defer os.Remove(tmpFile)

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "pool", "add-resource", "pool1", "--file", tmpFile, "--json")

	var out orch8.PoolResource
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != "res1" {
		t.Errorf("expected ID=res1, got %s", out.ID)
	}
	if out.PoolID != "pool1" {
		t.Errorf("expected PoolID=pool1, got %s", out.PoolID)
	}
}

func TestPoolUpdateResource(t *testing.T) {
	resource := orch8.PoolResource{
		ID:          "res1",
		PoolID:      "pool1",
		ResourceKey: "worker-1",
		State:       "locked",
		CreatedAt:   "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-02T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pools/pool1/resources/res1" && r.Method == http.MethodPut {
			body := decodeBody(t, r)
			if body["state"] != "locked" {
				t.Errorf("expected state=locked, got %v", body["state"])
			}
			json.NewEncoder(w).Encode(resource)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile := writeTempJSON(t, map[string]any{
		"state": "locked",
	})
	defer os.Remove(tmpFile)

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "pool", "update-resource", "pool1", "res1", "--file", tmpFile, "--json")

	var out orch8.PoolResource
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.ID != "res1" {
		t.Errorf("expected ID=res1, got %s", out.ID)
	}
	if out.State != "locked" {
		t.Errorf("expected State=locked, got %s", out.State)
	}
}

func TestPoolDeleteResource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pools/pool1/resources/res1" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, false)
	defer cleanup()

	getStdout := captureStdout(t)

	run(t, buf, "pool", "delete-resource", "pool1", "res1")

	got := strings.TrimSpace(getStdout())
	if got != "Deleted: res1" {
		t.Errorf("expected %q, got %q", "Deleted: res1", got)
	}
}

// ---------------------------------------------------------------------------
// Circuit breaker commands (per-tenant)
// ---------------------------------------------------------------------------

func TestCircuitBreakerListTenant(t *testing.T) {
	breakers := []orch8.CircuitBreakerState{
		{
			Handler:      "my-handler",
			State:        "closed",
			FailureCount: 0,
			LastFailure:  "",
		},
		{
			Handler:      "other-handler",
			State:        "open",
			FailureCount: 5,
			LastFailure:  "2025-01-01T00:00:00Z",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tenants/t1/circuit-breakers" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(breakers)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "cb", "list", "--json")

	var out []orch8.CircuitBreakerState
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 circuit breakers, got %d", len(out))
	}
	if out[0].Handler != "my-handler" {
		t.Errorf("expected Handler=my-handler, got %s", out[0].Handler)
	}
	if out[1].State != "open" {
		t.Errorf("expected State=open, got %s", out[1].State)
	}
}

func TestCircuitBreakerGetTenant(t *testing.T) {
	breaker := orch8.CircuitBreakerState{
		Handler:      "my-handler",
		State:        "open",
		FailureCount: 3,
		LastFailure:  "2025-01-01T00:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tenants/t1/circuit-breakers/my-handler" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(breaker)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "cb", "get", "my-handler", "--json")

	var out orch8.CircuitBreakerState
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.Handler != "my-handler" {
		t.Errorf("expected Handler=my-handler, got %s", out.Handler)
	}
	if out.FailureCount != 3 {
		t.Errorf("expected FailureCount=3, got %d", out.FailureCount)
	}
}

func TestCircuitBreakerResetTenant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tenants/t1/circuit-breakers/my-handler/reset" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, false)
	defer cleanup()

	getStdout := captureStdout(t)

	run(t, buf, "cb", "reset", "my-handler")

	got := strings.TrimSpace(getStdout())
	if got != "Reset: my-handler" {
		t.Errorf("expected %q, got %q", "Reset: my-handler", got)
	}
}

// ---------------------------------------------------------------------------
// Approval commands
// ---------------------------------------------------------------------------

func TestApprovalListRich(t *testing.T) {
	response := orch8.ApprovalsResponse{
		Items: []orch8.ApprovalItem{
			{
				InstanceID:   "i1",
				TenantID:     "t1",
				Namespace:    "default",
				SequenceID:   "s1",
				SequenceName: "my-seq",
				BlockID:      "step-approve",
				Prompt:       "Please approve this deployment",
				Choices:      []orch8.HumanChoice{},
				WaitingSince: "2025-01-01T00:00:00Z",
				Deadline:     nil,
				AllowComment: false,
			},
		},
		Total: 1,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/approvals" && r.Method == http.MethodGet {
			if q := r.URL.Query().Get("tenant_id"); q != "t1" {
				t.Errorf("expected tenant_id=t1, got %s", q)
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	buf, cleanup := setupTest(srv.URL, true)
	defer cleanup()

	run(t, buf, "approval", "list", "--json")

	var out orch8.ApprovalsResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("expected Total=1, got %d", out.Total)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	item := out.Items[0]
	if item.InstanceID != "i1" {
		t.Errorf("expected InstanceID=i1, got %s", item.InstanceID)
	}
	if item.SequenceName != "my-seq" {
		t.Errorf("expected SequenceName=my-seq, got %s", item.SequenceName)
	}
	if item.BlockID != "step-approve" {
		t.Errorf("expected BlockID=step-approve, got %s", item.BlockID)
	}
	if item.Prompt != "Please approve this deployment" {
		t.Errorf("expected Prompt=%q, got %q", "Please approve this deployment", item.Prompt)
	}
	if item.AllowComment != false {
		t.Errorf("expected AllowComment=false, got %v", item.AllowComment)
	}
}
