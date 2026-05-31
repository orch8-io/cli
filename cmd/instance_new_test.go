package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/orch8-io/cli/internal/output"
	orch8 "github.com/orch8-io/sdk-go"
)

func TestInstanceUpdateContext(t *testing.T) {
	var gotBody map[string]any
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = false

	buf := new(bytes.Buffer)
	output.Out = buf
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "update-context", "abc123", "--context", `{"foo":"bar"}`})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", gotMethod)
	}
	if gotPath != "/instances/abc123/context" {
		t.Errorf("expected /instances/abc123/context, got %s", gotPath)
	}

	ctxVal, ok := gotBody["context"].(map[string]any)
	if !ok {
		t.Fatalf("expected context to be a map, got %T: %v", gotBody["context"], gotBody)
	}
	if ctxVal["foo"] != "bar" {
		t.Errorf("expected context.foo=bar, got %v", ctxVal["foo"])
	}
}

func TestInstanceBatchCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/batch" && r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(orch8.BatchCreateResponse{Created: 3})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "batch-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	batchData := []map[string]any{
		{"sequence_id": "seq1", "tenant_id": "t1"},
		{"sequence_id": "seq2", "tenant_id": "t1"},
		{"sequence_id": "seq3", "tenant_id": "t1"},
	}
	json.NewEncoder(tmpFile).Encode(batchData)
	tmpFile.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = true

	buf := new(bytes.Buffer)
	output.Out = buf
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "batch-create", "--file", tmpFile.Name()})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out orch8.BatchCreateResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if out.Created != 3 {
		t.Errorf("expected Created=3, got %d", out.Created)
	}
}

func TestInstanceCheckpoints(t *testing.T) {
	checkpoints := []orch8.Checkpoint{
		{ID: "cp1", InstanceID: "i1", CreatedAt: "2025-01-01T00:00:00Z"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/i1/checkpoints" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(checkpoints)
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "checkpoints", "i1", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out []orch8.Checkpoint
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(out))
	}
	if out[0].ID != "cp1" {
		t.Errorf("expected ID=cp1, got %s", out[0].ID)
	}
	if out[0].InstanceID != "i1" {
		t.Errorf("expected InstanceID=i1, got %s", out[0].InstanceID)
	}
}

func TestInstanceCheckpointSave(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/i1/checkpoints" && r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
			json.NewEncoder(w).Encode(orch8.Checkpoint{
				ID:         "cp-new",
				InstanceID: "i1",
				CreatedAt:  "2025-06-01T00:00:00Z",
			})
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "checkpoint-save", "i1", "--data", `{"state":"saved"}`, "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out orch8.Checkpoint
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if out.ID != "cp-new" {
		t.Errorf("expected ID=cp-new, got %s", out.ID)
	}
	if out.InstanceID != "i1" {
		t.Errorf("expected InstanceID=i1, got %s", out.InstanceID)
	}
}

func TestInstanceCheckpointLatest(t *testing.T) {
	cp := orch8.Checkpoint{
		ID:             "cp-latest",
		InstanceID:     "i1",
		CheckpointData: map[string]any{"step": "final"},
		CreatedAt:      "2025-06-01T12:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/i1/checkpoints/latest" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(cp)
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// checkpoint-latest always outputs JSON regardless of --json flag
	rootCmd.SetArgs([]string{"instance", "checkpoint-latest", "i1"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out orch8.Checkpoint
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if out.ID != "cp-latest" {
		t.Errorf("expected ID=cp-latest, got %s", out.ID)
	}
	if out.InstanceID != "i1" {
		t.Errorf("expected InstanceID=i1, got %s", out.InstanceID)
	}
}

func TestInstanceCheckpointPrune(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/i1/checkpoints/prune" && r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "checkpoint-prune", "i1", "--keep", "5"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keepVal, ok := gotBody["keep"]
	if !ok {
		t.Fatalf("expected keep in request body, got %v", gotBody)
	}
	// JSON numbers decode as float64
	if keepVal.(float64) != 5 {
		t.Errorf("expected keep=5, got %v", keepVal)
	}
}

func TestInstanceInjectBlocks(t *testing.T) {
	var gotBody any
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Path == "/instances/i1/inject-blocks" && r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "blocks-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	blocks := []map[string]any{
		{"type": "step", "handler": "send_email"},
		{"type": "step", "handler": "notify_slack"},
	}
	json.NewEncoder(tmpFile).Encode(blocks)
	tmpFile.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = false

	buf := new(bytes.Buffer)
	output.Out = buf
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"instance", "inject-blocks", "i1", "--file", tmpFile.Name()})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/instances/i1/inject-blocks" {
		t.Errorf("expected /instances/i1/inject-blocks, got %s", gotPath)
	}
	if gotBody == nil {
		t.Fatal("expected request body, got nil")
	}
}

func TestInstanceBulkState(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/bulk/state" && r.Method == http.MethodPatch {
			json.NewDecoder(r.Body).Decode(&gotBody)
			json.NewEncoder(w).Encode(orch8.BulkResponse{Updated: 10})
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"instance", "bulk-state",
		"--state", "Cancelled",
		"--filter-tenant", "t1",
		"--filter-states", "Failed,Pending",
		"--json",
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify request body
	if gotBody["new_state"] != "Cancelled" {
		t.Errorf("expected new_state=Cancelled, got %v", gotBody["new_state"])
	}
	filter, ok := gotBody["filter"].(map[string]any)
	if !ok {
		t.Fatalf("expected filter to be a map, got %T: %v", gotBody["filter"], gotBody)
	}
	if filter["tenant_id"] != "t1" {
		t.Errorf("expected filter.tenant_id=t1, got %v", filter["tenant_id"])
	}
	states, ok := filter["states"].([]any)
	if !ok {
		t.Fatalf("expected filter.states to be an array, got %T: %v", filter["states"], filter["states"])
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	if states[0] != "Failed" || states[1] != "Pending" {
		t.Errorf("expected states [Failed, Pending], got %v", states)
	}

	// Verify output
	var out orch8.BulkResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if out.Updated != 10 {
		t.Errorf("expected Updated=10, got %d", out.Updated)
	}
}

func TestInstanceBulkReschedule(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/bulk/reschedule" && r.Method == http.MethodPatch {
			json.NewDecoder(r.Body).Decode(&gotBody)
			json.NewEncoder(w).Encode(orch8.BulkResponse{Updated: 5})
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
	defer func() { output.Out = os.Stdout }()

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"instance", "bulk-reschedule",
		"--offset", "3600",
		"--filter-namespace", "prod",
		"--json",
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify request body
	if gotBody["offset_secs"].(float64) != 3600 {
		t.Errorf("expected offset_secs=3600, got %v", gotBody["offset_secs"])
	}
	filter, ok := gotBody["filter"].(map[string]any)
	if !ok {
		t.Fatalf("expected filter to be a map, got %T: %v", gotBody["filter"], gotBody)
	}
	if filter["namespace"] != "prod" {
		t.Errorf("expected filter.namespace=prod, got %v", filter["namespace"])
	}

	// Verify output
	var out orch8.BulkResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parsing output: %v (raw: %q)", err, buf.String())
	}
	if out.Updated != 5 {
		t.Errorf("expected Updated=5, got %d", out.Updated)
	}
}

func TestInstanceStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/i1/stream" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)

			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("expected ResponseWriter to be a Flusher")
				return
			}

			fmt.Fprintf(w, "data: {\"state\":\"running\"}\n\n")
			flusher.Flush()
			fmt.Fprintf(w, "data: {\"state\":\"completed\"}\n\n")
			flusher.Flush()
			// Server closes the connection by returning
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	flagURL = srv.URL
	flagTenantID = "t1"
	flagJSON = false

	// The stream command writes directly to os.Stdout via json.NewEncoder(os.Stdout),
	// so we need to capture os.Stdout with a pipe.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	outBuf := new(bytes.Buffer)
	output.Out = outBuf
	defer func() {
		os.Stdout = origStdout
		output.Out = os.Stdout
	}()

	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"instance", "stream", "i1"})

	var wg sync.WaitGroup
	var captured bytes.Buffer

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Read all output from the pipe until it is closed
		captured.ReadFrom(r)
	}()

	execErr := rootCmd.Execute()

	// Close the write end so the reader goroutine finishes
	w.Close()
	wg.Wait()

	// Restore stdout before any further t.Log/t.Error calls
	os.Stdout = origStdout

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	lines := strings.Split(strings.TrimSpace(captured.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 JSON lines, got %d: %q", len(lines), captured.String())
	}

	var event1, event2 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event1); err != nil {
		t.Fatalf("parsing first event: %v (line: %q)", err, lines[0])
	}
	if err := json.Unmarshal([]byte(lines[1]), &event2); err != nil {
		t.Fatalf("parsing second event: %v (line: %q)", err, lines[1])
	}

	if event1["state"] != "running" {
		t.Errorf("expected first event state=running, got %v", event1["state"])
	}
	if event2["state"] != "completed" {
		t.Errorf("expected second event state=completed, got %v", event2["state"])
	}
}
