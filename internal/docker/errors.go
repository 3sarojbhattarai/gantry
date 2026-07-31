package docker

import "errors"

// ErrNotImplemented is returned by engine methods whose read logic has not yet
// been written. Phase 0 ships the interface and wiring; the moby-backed
// implementation fills these in during Phase 1.
var ErrNotImplemented = errors.New("gantry: not implemented")

// ErrNotFound is returned when an object (container, image, network, volume)
// does not exist. Layers above the engine branch on it via errors.Is.
var ErrNotFound = errors.New("gantry: not found")
