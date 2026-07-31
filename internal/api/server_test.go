package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/api"
	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

func testServer(f *fakedocker.Fake) *httptest.Server {
	return httptest.NewServer(api.NewServer(f).Handler())
}

func TestListContainers(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "a", Names: []string{"web"}, State: "running"},
		{ID: "b", Names: []string{"db"}, State: "exited"},
	}}
	srv := testServer(f)
	defer srv.Close()

	// default: running only
	var running []docker.Container
	getJSON(t, srv.URL+"/api/containers", &running)
	if len(running) != 1 || running[0].ID != "a" {
		t.Fatalf("default list = %+v, want just a", running)
	}
	// all=true
	var all []docker.Container
	getJSON(t, srv.URL+"/api/containers?all=true", &all)
	if len(all) != 2 {
		t.Fatalf("all list = %d, want 2", len(all))
	}
}

func TestInspectContainerNotFound(t *testing.T) {
	srv := testServer(&fakedocker.Fake{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/containers/missing")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestStartContainer(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", State: "exited"}}}
	srv := testServer(f)
	defer srv.Close()

	resp := post(t, srv.URL+"/api/containers/a/start")
	if resp != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp)
	}
	if f.Containers[0].State != "running" {
		t.Fatalf("state = %q, want running", f.Containers[0].State)
	}
}

func TestRemoveContainerConsent(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", State: "exited"}}}
	srv := testServer(f)
	defer srv.Close()

	// Without confirm => 409 Conflict, container survives.
	if code := del(t, srv.URL+"/api/containers/a"); code != http.StatusConflict {
		t.Fatalf("no-confirm delete status = %d, want 409", code)
	}
	if len(f.Containers) != 1 {
		t.Fatal("container removed without confirm")
	}
	// With confirm => 204, container gone.
	if code := del(t, srv.URL+"/api/containers/a?confirm=true"); code != http.StatusNoContent {
		t.Fatalf("confirmed delete status = %d, want 204", code)
	}
	if len(f.Containers) != 0 {
		t.Fatal("container not removed with confirm")
	}
}

func TestPruneDryRun(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "keep", State: "running"},
		{ID: "drop", State: "exited"},
	}}
	srv := testServer(f)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/prune/containers?dryRun=true", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var report docker.PruneReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	if !report.DryRun || len(report.Items) != 1 || report.Items[0] != "drop" {
		t.Fatalf("report = %+v", report)
	}
	if len(f.Containers) != 2 {
		t.Fatal("dry-run removed containers")
	}
}

func TestCreateNetwork(t *testing.T) {
	f := &fakedocker.Fake{}
	srv := testServer(f)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/networks", "application/json", strings.NewReader(`{"name":"mynet"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if len(f.Networks) != 1 || f.Networks[0].Name != "mynet" {
		t.Fatalf("network not created: %+v", f.Networks)
	}
}

func TestEventsSSE(t *testing.T) {
	f := &fakedocker.Fake{EventStream: []docker.Event{
		{Type: "container", Action: "start"},
		{Type: "container", Action: "die"},
	}}
	srv := testServer(f)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}
	body, _ := io.ReadAll(resp.Body) // fake closes the stream, so this returns
	if !strings.Contains(string(body), `"action":"start"`) {
		t.Fatalf("event body missing start event:\n%s", body)
	}
}

func TestLogsSSE(t *testing.T) {
	f := &fakedocker.Fake{Logs: map[string]string{"a": "line one\nline two\n"}}
	srv := testServer(f)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/containers/a/logs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "data: line one") || !strings.Contains(string(body), "data: line two") {
		t.Fatalf("log SSE body unexpected:\n%s", body)
	}
}

func TestHealthAndFrontendPlaceholder(t *testing.T) {
	srv := testServer(&fakedocker.Fake{})
	defer srv.Close()

	var health map[string]bool
	getJSON(t, srv.URL+"/api/health", &health)
	if !health["ok"] {
		t.Fatal("health not ok")
	}

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("frontend status = %d, want 200", resp.StatusCode)
	}
}

// --- helpers ----------------------------------------------------------------

func getJSON(t *testing.T, url string, v any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}

func post(t *testing.T, url string) int {
	t.Helper()
	resp, err := http.Post(url, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func del(t *testing.T, url string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
