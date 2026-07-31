package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

func (m *mobyClient) CreateContainer(ctx context.Context, spec CreateSpec, start bool) (string, error) {
	cfg, hostCfg, err := toCreateConfig(spec)
	if err != nil {
		return "", fmt.Errorf("gantry: building container config: %w", err)
	}
	resp, err := m.cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, spec.Name)
	if err != nil {
		return "", fmt.Errorf("gantry: creating container: %w", mapErr(err))
	}
	if start {
		if err := m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return resp.ID, fmt.Errorf("gantry: starting created container %s: %w", resp.ID, mapErr(err))
		}
	}
	return resp.ID, nil
}

func toCreateConfig(spec CreateSpec) (*container.Config, *container.HostConfig, error) {
	cfg := &container.Config{
		Image:      spec.Image,
		Cmd:        spec.Command,
		Env:        spec.Env,
		Labels:     spec.Labels,
		WorkingDir: spec.WorkingDir,
		User:       spec.User,
	}
	hostCfg := &container.HostConfig{
		Binds: spec.Volumes,
	}

	if len(spec.Ports) > 0 {
		cfg.ExposedPorts = nat.PortSet{}
		hostCfg.PortBindings = nat.PortMap{}
		for _, p := range spec.Ports {
			port, err := nat.NewPort(p.proto(), strconv.Itoa(int(p.Container)))
			if err != nil {
				return nil, nil, fmt.Errorf("invalid port %d/%s: %w", p.Container, p.proto(), err)
			}
			cfg.ExposedPorts[port] = struct{}{}
			if p.Host != "" {
				hostCfg.PortBindings[port] = []nat.PortBinding{{HostPort: p.Host}}
			}
		}
	}

	if spec.RestartPolicy != "" {
		hostCfg.RestartPolicy = container.RestartPolicy{
			Name: container.RestartPolicyMode(spec.RestartPolicy),
		}
	}
	// A single network is set as the network mode; multiple networks are a
	// progressive-disclosure extension for later.
	if len(spec.Networks) > 0 {
		hostCfg.NetworkMode = container.NetworkMode(spec.Networks[0])
	}
	return cfg, hostCfg, nil
}

// SpecFromContainer reads an existing container's configuration into a
// CreateSpec — the basis of the --from "clone and tweak" workflow.
func (m *mobyClient) SpecFromContainer(ctx context.Context, id string) (CreateSpec, error) {
	r, err := m.cli.ContainerInspect(ctx, id)
	if err != nil {
		return CreateSpec{}, fmt.Errorf("gantry: reading config of %s: %w", id, mapErr(err))
	}
	spec := CreateSpec{}
	if r.ContainerJSONBase != nil {
		spec.Name = strings.TrimPrefix(r.Name, "/")
		if r.HostConfig != nil {
			spec.RestartPolicy = string(r.HostConfig.RestartPolicy.Name)
			spec.Volumes = r.HostConfig.Binds
			for port, bindings := range r.HostConfig.PortBindings {
				pm := PortMapping{Container: uint16(port.Int()), Proto: port.Proto()}
				if len(bindings) > 0 {
					pm.Host = bindings[0].HostPort
				}
				spec.Ports = append(spec.Ports, pm)
			}
		}
	}
	if r.Config != nil {
		spec.Image = r.Config.Image
		spec.Command = r.Config.Cmd
		spec.Env = r.Config.Env
		spec.Labels = r.Config.Labels
		spec.WorkingDir = r.Config.WorkingDir
		spec.User = r.Config.User
	}
	return spec, nil
}
