// Package docker defines gantry's own view of the Docker daemon: a set of
// domain types and a Client interface over them. Nothing from the underlying
// SDK (moby's github.com/docker/docker types) is allowed to leak past this
// package. Consumers — the CLI, TUI, and API layers — depend only on the
// types and interface declared here, which keeps them testable against a fake
// and preserves the option of swapping the moby SDK for hand-rolled HTTP.
package docker

import "time"

// Container is the summary view of a container as shown in list output. JSON
// tags define the wire contract the web frontend consumes.
type Container struct {
	ID      string            `json:"id"`
	Names   []string          `json:"names"`
	Image   string            `json:"image"`
	ImageID string            `json:"imageId"`
	Command string            `json:"command"`
	Created time.Time         `json:"created"`
	State   string            `json:"state"` // running, exited, paused, created, restarting, dead
	Status  string            `json:"status"`
	Ports   []Port            `json:"ports"`
	Labels  map[string]string `json:"labels"`
}

// Port describes a single published or exposed port mapping.
type Port struct {
	IP      string `json:"ip"`
	Private uint16 `json:"private"`
	Public  uint16 `json:"public"`
	Type    string `json:"type"` // tcp, udp, sctp
}

// ImageSummary is the summary view of an image as shown in list output.
type ImageSummary struct {
	ID       string            `json:"id"`
	RepoTags []string          `json:"repoTags"`
	Created  time.Time         `json:"created"`
	Size     int64             `json:"size"`
	Labels   map[string]string `json:"labels"`
}

// Network is the summary view of a network.
type Network struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Driver   string            `json:"driver"`
	Scope    string            `json:"scope"`
	Created  time.Time         `json:"created"`
	Internal bool              `json:"internal"`
	Labels   map[string]string `json:"labels"`
}

// Volume is the summary view of a volume.
type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Created    time.Time         `json:"created"`
	Labels     map[string]string `json:"labels"`
}

// Stats is a single sampled resource-usage reading for a container.
type Stats struct {
	ContainerID string  `json:"containerId"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemUsage    uint64  `json:"memUsage"`
	MemLimit    uint64  `json:"memLimit"`
	MemPercent  float64 `json:"memPercent"`
	NetRx       uint64  `json:"netRx"`
	NetTx       uint64  `json:"netTx"`
	BlockRead   uint64  `json:"blockRead"`
	BlockWrite  uint64  `json:"blockWrite"`
	PIDs        uint64  `json:"pids"`
}

// Event is a single message from the daemon's event stream.
type Event struct {
	Type   string    `json:"type"`   // container, image, network, volume
	Action string    `json:"action"` // start, stop, die, create, destroy, ...
	Actor  string    `json:"actor"`  // ID of the object the event concerns
	Name   string    `json:"name"`   // name of the object, when known
	Time   time.Time `json:"time"`
}

// ContainerDetails is the full inspect view of a container: the summary fields
// plus the extra detail only the inspect endpoint returns.
type ContainerDetails struct {
	Container
	Path         string    `json:"path"`
	Args         []string  `json:"args"`
	Env          []string  `json:"env"`
	Platform     string    `json:"platform"`
	RestartCount int       `json:"restartCount"`
	ExitCode     int       `json:"exitCode"`
	Error        string    `json:"error"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
}

// ImageDetails is the full inspect view of an image: the summary fields plus
// the extra detail only the inspect endpoint returns.
type ImageDetails struct {
	ImageSummary
	RepoDigests  []string `json:"repoDigests"`
	Architecture string   `json:"architecture"`
	Os           string   `json:"os"`
	Author       string   `json:"author"`
}

// LogOptions controls how container logs are read.
type LogOptions struct {
	Follow     bool
	Timestamps bool
	Tail       string // number of lines from the end, or "all"
}
