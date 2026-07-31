package cli

import (
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// newClient constructs the engine client. It is a package var so tests can
// substitute a fakedocker.Fake without a running daemon.
var newClient = docker.New

// withClient builds a client, verifies the daemon is reachable, and runs fn,
// closing the client afterwards. Commands funnel through it so daemon setup
// and teardown live in one place.
func withClient(ctx context.Context, fn func(context.Context, docker.Client) error) error {
	cli, err := newClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Ping(ctx); err != nil {
		return err
	}
	return fn(ctx, cli)
}
