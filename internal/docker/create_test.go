package docker

import (
	"strings"
	"testing"
)

func sampleSpec() CreateSpec {
	return CreateSpec{
		Image:         "nginx:alpine",
		Name:          "web",
		Command:       []string{"nginx", "-g", "daemon off;"},
		Env:           []string{"TZ=UTC", "GREETING=hello world"},
		Ports:         []PortMapping{{Host: "8080", Container: 80, Proto: "tcp"}},
		RestartPolicy: "unless-stopped",
		Volumes:       []string{"/data:/usr/share/nginx/html:ro"},
		Labels:        map[string]string{"team": "web"},
	}
}

func TestSpecToDockerRun(t *testing.T) {
	got := SpecToDockerRun(sampleSpec())
	for _, want := range []string{
		"docker run -d",
		"--name web",
		"--restart unless-stopped",
		"-p 8080:80/tcp",
		"-v /data:/usr/share/nginx/html:ro",
		"--label team=web",
		"nginx:alpine",
		"nginx -g",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("docker run missing %q in:\n%s", want, got)
		}
	}
	// The env value with a space must be quoted so the command is paste-safe.
	if !strings.Contains(got, "'GREETING=hello world'") {
		t.Errorf("env value not quoted:\n%s", got)
	}
}

func TestSpecToCompose(t *testing.T) {
	got := SpecToCompose(sampleSpec())
	for _, want := range []string{
		"services:",
		"  web:",
		"    image: nginx:alpine",
		"    container_name: web",
		`"8080:80/tcp"`,
		"    restart: unless-stopped",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("compose missing %q in:\n%s", want, got)
		}
	}
}

func TestPortProtoDefault(t *testing.T) {
	p := PortMapping{Container: 80}
	if p.proto() != "tcp" {
		t.Errorf("default proto = %q, want tcp", p.proto())
	}
}

func TestToCreateConfig(t *testing.T) {
	cfg, host, err := toCreateConfig(sampleSpec())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Image != "nginx:alpine" {
		t.Errorf("image = %q", cfg.Image)
	}
	if len(cfg.ExposedPorts) != 1 {
		t.Errorf("exposed ports = %d, want 1", len(cfg.ExposedPorts))
	}
	if len(host.PortBindings) != 1 {
		t.Errorf("port bindings = %d, want 1", len(host.PortBindings))
	}
	if string(host.RestartPolicy.Name) != "unless-stopped" {
		t.Errorf("restart policy = %q", host.RestartPolicy.Name)
	}
	if len(host.Binds) != 1 {
		t.Errorf("binds = %d, want 1", len(host.Binds))
	}
}
