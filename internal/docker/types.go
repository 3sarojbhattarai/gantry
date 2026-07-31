// Package docker defines gantry's own view of the Docker daemon: a set of
// domain types and a Client interface over them. Nothing from the underlying
// SDK (moby's github.com/docker/docker types) is allowed to leak past this
// package. Consumers — the CLI, TUI, and API layers — depend only on the
// types and interface declared here, which keeps them testable against a fake
// and preserves the option of swapping the moby SDK for hand-rolled HTTP.
package docker

import "time"

// Container is the summary view of a container as shown in list output.
type Container struct {
	ID      string
	Names   []string
	Image   string
	ImageID string
	Command string
	Created time.Time
	State   string // running, exited, paused, created, restarting, dead
	Status  string // human-readable, e.g. "Up 3 hours"
	Ports   []Port
	Labels  map[string]string
}

// Port describes a single published or exposed port mapping.
type Port struct {
	IP      string
	Private uint16
	Public  uint16
	Type    string // tcp, udp, sctp
}

// ImageSummary is the summary view of an image as shown in list output.
type ImageSummary struct {
	ID       string
	RepoTags []string
	Created  time.Time
	Size     int64
	Labels   map[string]string
}

// Network is the summary view of a network.
type Network struct {
	ID       string
	Name     string
	Driver   string
	Scope    string
	Created  time.Time
	Internal bool
	Labels   map[string]string
}

// Volume is the summary view of a volume.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	Created    time.Time
	Labels     map[string]string
}

// Stats is a single sampled resource-usage reading for a container.
type Stats struct {
	ContainerID string
	CPUPercent  float64
	MemUsage    uint64
	MemLimit    uint64
	MemPercent  float64
	NetRx       uint64
	NetTx       uint64
	BlockRead   uint64
	BlockWrite  uint64
	PIDs        uint64
}

// Event is a single message from the daemon's event stream.
type Event struct {
	Type   string // container, image, network, volume
	Action string // start, stop, die, create, destroy, ...
	Actor  string // ID of the object the event concerns
	Name   string // name of the object, when known
	Time   time.Time
}

// ContainerDetails is the full inspect view of a container. It is intentionally
// thin in Phase 0 and fleshed out in Phase 1 when the read paths land; the
// Client.InspectContainer signature is fixed now so the shape is stable.
type ContainerDetails struct {
	Container
}

// ImageDetails is the full inspect view of an image. Thin in Phase 0, fleshed
// out in Phase 1 (see ContainerDetails).
type ImageDetails struct {
	ImageSummary
}

// LogOptions controls how container logs are read.
type LogOptions struct {
	Follow     bool
	Timestamps bool
	Tail       string // number of lines from the end, or "all"
}
