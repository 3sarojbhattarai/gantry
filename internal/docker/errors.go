package docker

import "errors"

// ErrNotImplemented is returned by engine methods whose read logic has not yet
// been written. Phase 0 ships the interface and wiring; the moby-backed
// implementation fills these in during Phase 1.
var ErrNotImplemented = errors.New("gantry: not implemented")

// ErrNotFound is returned when an object (container, image, network, volume)
// does not exist. Layers above the engine branch on it via errors.Is.
var ErrNotFound = errors.New("gantry: not found")

// ErrConsentRequired is returned by destructive operations that were called
// without explicit consent. Renderers obtain consent their own way (CLI
// --force, TUI modal, web dialog) and pass docker.Confirm() to proceed.
var ErrConsentRequired = errors.New("gantry: destructive operation requires explicit consent")

// ErrDryRunUnsupported is returned when a prune dry-run is requested for a
// resource whose candidates cannot be previewed from list data alone
// (volumes and networks, which need daemon-side reference tracking).
var ErrDryRunUnsupported = errors.New("gantry: dry-run not supported for this resource")
