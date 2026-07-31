// Package termexec wires a docker.ExecSession to a real terminal: raw mode,
// bidirectional copy, and TTY resize propagation. Both the CLI (`gantry exec`)
// and the TUI (suspend-and-attach) use it, so the interactive attach lives in
// one place.
package termexec

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"golang.org/x/term"
)

// Attach connects sess to in/out. When in is a terminal it is put in raw mode
// for the duration and SIGWINCH resizes are forwarded to the session. It
// returns when either side of the stream closes.
func Attach(sess docker.ExecSession, in, out *os.File) error {
	fd := int(in.Fd())
	if term.IsTerminal(fd) {
		old, err := term.MakeRaw(fd)
		if err == nil {
			defer func() { _ = term.Restore(fd, old) }()
		}
		resize(sess, out)

		winch := make(chan os.Signal, 1)
		signal.Notify(winch, syscall.SIGWINCH)
		defer signal.Stop(winch)
		go func() {
			for range winch {
				resize(sess, out)
			}
		}()
	}

	go func() {
		_, _ = io.Copy(sess, in)
		_ = sess.Close()
	}()
	_, err := io.Copy(out, sess)
	return err
}

func resize(sess docker.ExecSession, out *os.File) {
	if w, h, err := term.GetSize(int(out.Fd())); err == nil {
		_ = sess.Resize(uint(h), uint(w))
	}
}
