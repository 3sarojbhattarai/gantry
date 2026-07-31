package docker

import "io"

// ExecOptions configures a command run inside a running container.
type ExecOptions struct {
	Cmd        []string
	TTY        bool
	Env        []string
	WorkingDir string
	User       string
}

// ExecSession is a live, bidirectional connection to a process running inside a
// container. Callers write stdin to it, read stdout/stderr from it, resize the
// TTY as the viewport changes, and Close it when done. It hides the SDK's
// hijacked-connection type from the layers above.
type ExecSession interface {
	io.ReadWriteCloser
	// Resize informs the process of a new TTY size (rows, cols).
	Resize(rows, cols uint) error
}
