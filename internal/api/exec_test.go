package api_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
)

func TestCreateContainerEndpoint(t *testing.T) {
	f := &fakedocker.Fake{}
	srv := testServer(f)
	defer srv.Close()

	resp, err := http.Post(
		srv.URL+"/api/containers?start=true",
		"application/json",
		strings.NewReader(`{"image":"nginx","name":"web"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if len(f.Containers) != 1 || f.Containers[0].State != "running" {
		t.Fatalf("container not created/started: %+v", f.Containers)
	}
}

func TestCreateContainerRequiresImage(t *testing.T) {
	srv := testServer(&fakedocker.Fake{})
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/containers", "application/json", strings.NewReader(`{"name":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestContainerSpecEndpoint(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{
		{ID: "a", Names: []string{"web"}, Image: "nginx"},
	}}
	srv := testServer(f)
	defer srv.Close()

	var spec docker.CreateSpec
	getJSON(t, srv.URL+"/api/containers/a/spec", &spec)
	if spec.Image != "nginx" || spec.Name != "web" {
		t.Fatalf("spec = %+v", spec)
	}
}

func TestExportEndpoint(t *testing.T) {
	srv := testServer(&fakedocker.Fake{})
	defer srv.Close()

	resp, err := http.Post(
		srv.URL+"/api/export/run",
		"application/json",
		strings.NewReader(`{"image":"nginx","name":"web"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body["text"], "docker run") || !strings.Contains(body["text"], "nginx") {
		t.Fatalf("export text = %q", body["text"])
	}
}
