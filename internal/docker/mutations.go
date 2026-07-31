package docker

// This file defines gantry's mutation surface: the options types and the
// consent model that gates destructive operations. The moby-backed
// implementations live in moby.go; the in-memory versions live in fakedocker.

// Consent is an explicit confirmation that a destructive operation may proceed.
// Destructive engine methods refuse to run unless consent is granted, so the
// rule "confirm before you destroy" lives in the engine's contract rather than
// being reinvented by each renderer. Renderers construct one with Confirm()
// after obtaining the user's approval — `--force` in the CLI, a modal in the
// TUI, a dialog on the web.
//
// The zero value is *not* granted, so a caller cannot accidentally destroy data
// by leaving an options struct blank.
type Consent struct{ ok bool }

// Confirm returns granted consent.
func Confirm() Consent { return Consent{ok: true} }

// Granted reports whether consent was given.
func (c Consent) Granted() bool { return c.ok }

// StopOptions controls how a container is stopped or restarted.
type StopOptions struct {
	// Timeout is the grace period, in seconds, before the container is killed.
	// nil uses the daemon default (10s).
	Timeout *int
}

// RemoveContainerOptions controls container removal. It is destructive and
// carries Consent.
type RemoveContainerOptions struct {
	Consent
	Force         bool // remove even if running (kill first)
	RemoveVolumes bool // remove anonymous volumes attached to the container
}

// RemoveImageOptions controls image removal. It is destructive and carries
// Consent.
type RemoveImageOptions struct {
	Consent
	Force         bool // remove even if tagged in multiple repos / used
	PruneChildren bool // remove untagged parent images
}

// RemoveVolumeOptions controls volume removal. It is destructive and carries
// Consent.
type RemoveVolumeOptions struct {
	Consent
	Force bool
}

// RemoveNetworkOptions controls network removal. It carries Consent.
type RemoveNetworkOptions struct {
	Consent
}

// PruneOptions controls a prune. A prune is destructive and carries Consent,
// except when DryRun is set — a dry run only reports what would be removed and
// needs no consent.
type PruneOptions struct {
	Consent
	DryRun bool
}

// PruneReport summarizes what a prune removed, or (on a dry run) what it would
// remove.
type PruneReport struct {
	Items          []string `json:"items"`          // IDs or names removed / to be removed
	SpaceReclaimed uint64   `json:"spaceReclaimed"` // bytes; may be 0 when unknown (e.g. container dry run)
	DryRun         bool     `json:"dryRun"`
}

// CreateNetworkOptions describes a network to create.
type CreateNetworkOptions struct {
	Name     string
	Driver   string // defaults to "bridge" when empty
	Internal bool
	Labels   map[string]string
}
