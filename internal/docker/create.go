package docker

import (
	"fmt"
	"sort"
	"strings"
)

// CreateSpec is gantry's renderer-agnostic description of a container to
// create. It covers the common surface (image, name, command, env, ports,
// restart policy, volumes, labels); the web form exposes these with progressive
// disclosure, the CLI reads them from YAML, and --from fills one in from an
// existing container.
type CreateSpec struct {
	Image         string            `json:"image" yaml:"image"`
	Name          string            `json:"name" yaml:"name"`
	Command       []string          `json:"command" yaml:"command"`
	Env           []string          `json:"env" yaml:"env"` // KEY=value
	Ports         []PortMapping     `json:"ports" yaml:"ports"`
	RestartPolicy string            `json:"restartPolicy" yaml:"restart_policy"` // "", no, always, unless-stopped, on-failure
	Volumes       []string          `json:"volumes" yaml:"volumes"`              // docker bind syntax: src:dst[:ro]
	Networks      []string          `json:"networks" yaml:"networks"`
	Labels        map[string]string `json:"labels" yaml:"labels"`
	WorkingDir    string            `json:"workingDir" yaml:"working_dir"`
	User          string            `json:"user" yaml:"user"`
}

// PortMapping is one published port. Host may be empty to only expose the port.
type PortMapping struct {
	Host      string `json:"host" yaml:"host"`
	Container uint16 `json:"container" yaml:"container"`
	Proto     string `json:"proto" yaml:"proto"` // tcp (default) or udp
}

func (p PortMapping) proto() string {
	if p.Proto == "" {
		return "tcp"
	}
	return p.Proto
}

// SpecToDockerRun renders the spec as an equivalent `docker run` command. This
// makes the create form a teaching tool and lets users see exactly what gantry
// would do.
func SpecToDockerRun(s CreateSpec) string {
	args := []string{"docker", "run", "-d"}
	if s.Name != "" {
		args = append(args, "--name", s.Name)
	}
	if s.RestartPolicy != "" && s.RestartPolicy != "no" {
		args = append(args, "--restart", s.RestartPolicy)
	}
	for _, p := range s.Ports {
		if p.Host != "" {
			args = append(args, "-p", fmt.Sprintf("%s:%d/%s", p.Host, p.Container, p.proto()))
		} else {
			args = append(args, "--expose", fmt.Sprintf("%d/%s", p.Container, p.proto()))
		}
	}
	for _, e := range s.Env {
		args = append(args, "-e", shellQuote(e))
	}
	for _, v := range s.Volumes {
		args = append(args, "-v", v)
	}
	for _, k := range sortedKeys(s.Labels) {
		args = append(args, "--label", shellQuote(k+"="+s.Labels[k]))
	}
	for _, n := range s.Networks {
		args = append(args, "--network", n)
	}
	if s.WorkingDir != "" {
		args = append(args, "-w", s.WorkingDir)
	}
	if s.User != "" {
		args = append(args, "-u", s.User)
	}
	args = append(args, s.Image)
	args = append(args, s.Command...)
	return strings.Join(args, " ")
}

// SpecToCompose renders the spec as a docker-compose service fragment.
func SpecToCompose(s CreateSpec) string {
	name := s.Name
	if name == "" {
		name = "app"
	}
	var b strings.Builder
	b.WriteString("services:\n")
	fmt.Fprintf(&b, "  %s:\n", name)
	fmt.Fprintf(&b, "    image: %s\n", s.Image)
	if s.Name != "" {
		fmt.Fprintf(&b, "    container_name: %s\n", s.Name)
	}
	if len(s.Command) > 0 {
		fmt.Fprintf(&b, "    command: %s\n", yamlList(s.Command))
	}
	if len(s.Ports) > 0 {
		ports := make([]string, len(s.Ports))
		for i, p := range s.Ports {
			if p.Host != "" {
				ports[i] = fmt.Sprintf("%s:%d/%s", p.Host, p.Container, p.proto())
			} else {
				ports[i] = fmt.Sprintf("%d/%s", p.Container, p.proto())
			}
		}
		fmt.Fprintf(&b, "    ports: %s\n", yamlList(ports))
	}
	if len(s.Env) > 0 {
		fmt.Fprintf(&b, "    environment: %s\n", yamlList(s.Env))
	}
	if len(s.Volumes) > 0 {
		fmt.Fprintf(&b, "    volumes: %s\n", yamlList(s.Volumes))
	}
	if len(s.Networks) > 0 {
		fmt.Fprintf(&b, "    networks: %s\n", yamlList(s.Networks))
	}
	if s.RestartPolicy != "" && s.RestartPolicy != "no" {
		fmt.Fprintf(&b, "    restart: %s\n", s.RestartPolicy)
	}
	if s.WorkingDir != "" {
		fmt.Fprintf(&b, "    working_dir: %s\n", s.WorkingDir)
	}
	if s.User != "" {
		fmt.Fprintf(&b, "    user: %q\n", s.User)
	}
	if len(s.Labels) > 0 {
		labels := make([]string, 0, len(s.Labels))
		for _, k := range sortedKeys(s.Labels) {
			labels = append(labels, k+"="+s.Labels[k])
		}
		fmt.Fprintf(&b, "    labels: %s\n", yamlList(labels))
	}
	return b.String()
}

func yamlList(items []string) string {
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = fmt.Sprintf("%q", it)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// shellQuote wraps a value in single quotes if it contains shell-significant
// characters, so the rendered command is safe to paste.
func shellQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\"'$&|<>();`\\*?") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
