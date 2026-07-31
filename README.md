# Gantry

A single-binary Docker manager with a terminal UI and a web UI, built in Go.

Gantry is a docker tool for inspecting and managing a local Docker daemon — 
list and inspect containers, images, networks and volumes; tail logs; watch 
live stats and events; start, stop, prune and create containers — from either 
your terminal or a browser, backed by the same engine.

## Why

- **One tool, two surfaces.** A TUI for quick terminal work and a web UI for a
  richer view — both driven by one Go engine, so behavior never diverges.
- **Live, not polled.** The daemon's `/events` and `/stats` streams feed the UI,
  so the view reflects reality instead of a stale snapshot.
- **Own the domain.** Gantry defines its own `Container`, `Image`, `Network`,
  and `Volume` types behind an interface. The Docker SDK is an implementation
  detail, never a dependency the UI layers see.

## Architecture

```
cmd/gantry/          main + cobra root
internal/docker/     engine: Client interface, domain types, moby impl
  fakedocker/        in-memory Client for daemon-free tests
internal/cli/        one-shot commands (ps, images, logs, …)
internal/tui/        Bubbletea terminal UI
internal/api/        HTTP handlers, SSE streams, WebSocket exec
internal/web/        embed.FS for the built frontend
web/                 Vite + React frontend
```

The one rule that shapes everything: **moby's SDK types never leak past
`internal/docker`.** Every renderer — CLI, TUI, API — depends only on the
`docker.Client` interface and gantry's own structs. This keeps the engine
swappable (hand-rolled HTTP later, if wanted) and everything testable against
`fakedocker.Fake` with no daemon running.

## Tech stack

| Layer      | Choice                                                             |
| ---------- | ----------------------------------------------------------------- |
| Language   | Go 1.26                                                            |
| CLI        | [cobra](https://github.com/spf13/cobra)                           |
| Docker API | [moby SDK](https://github.com/moby/moby) (`docker/docker/client`) |
| TUI        | [Bubbletea](https://github.com/charmbracelet/bubbletea)           |
| Web API    | Go stdlib `net/http`, Server-Sent Events, WebSocket              |
| Frontend   | React + Vite + Tailwind CSS + TanStack Query, xterm.js            |
| Release    | goreleaser, Homebrew tap, GHCR Docker image                      |
| CI         | GitHub Actions (build, vet, test, lint, frontend, release)        |

## Install

Prebuilt binaries, Homebrew, and a Docker image are published on each tagged
release (via goreleaser).

**Homebrew:**

```bash
brew install 3sarojbhattarai/tap/gantry
```

**Go:**

```bash
go install github.com/3sarojbhattarai/gantry/cmd/gantry@latest
```

**Prebuilt binary:** download the archive for your platform from the
[releases page](https://github.com/3sarojbhattarai/gantry/releases) and put
`gantry` on your `PATH`.

**Docker:**

```bash
docker run --rm -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/3sarojbhattarai/gantry:latest
```

> ⚠️ **Security.** The Docker image mounts your host's Docker socket, which grants
> the container — and anyone who can reach the published port — **full,
> unauthenticated control of your Docker daemon** (start/stop/delete containers,
> read images, run `exec`). Only publish the port on a trusted network, prefer
> binding it to `127.0.0.1` (`-p 127.0.0.1:8080:8080`), and never expose it to
> the public internet.

## Build from source

Requires Go 1.26+ (and Node 20+ only if you want the frontend embedded) and a
running Docker daemon.

```bash
git clone https://github.com/3sarojbhattarai/gantry
cd gantry
make build          # dev binary (frontend not embedded)
make build-embed    # self-contained binary with the web UI
```

Then:

```bash
./bin/gantry            # launch the terminal UI
./bin/gantry ps -a      # list all containers
./bin/gantry images     # list images
./bin/gantry logs -f <container>   # follow a container's logs

# mutations (destructive ones need --force to confirm)
./bin/gantry start|stop|restart|kill <container>...
./bin/gantry rm -f <container>...
./bin/gantry rmi -f <image>...
./bin/gantry prune containers --dry-run      # preview
./bin/gantry prune images -f                 # reclaim
./bin/gantry network create <name>

# exec + create
./bin/gantry exec <container> sh             # interactive shell
./bin/gantry create --from <container> --export run   # see the docker run equiv
./bin/gantry create --file spec.yaml --start          # create from YAML
```

In the TUI: `1-4` or `tab` switch panes, `j`/`k` move, `s`/`S`/`R`/`K`
start/stop/restart/kill, `d` removes (with a confirmation prompt), `e` opens a
shell in the selected container, `r` refreshes, `?` toggles help, `q` quits.

For the web UI:

```bash
./bin/gantry serve                 # http://127.0.0.1:8080 (frontend + API)
```

`make build` produces a fast dev binary with the frontend *not* embedded (it
serves a placeholder pointing at the Vite dev server). `make build-embed` builds
the frontend and bakes it in for a single self-contained binary. During frontend
development, run the API and the Vite dev server side by side:

```bash
./bin/gantry serve      # terminal 1: API on :8080
make web-dev            # terminal 2: Vite on :5173, proxies /api to :8080
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full local-development workflow.

## Roadmap

Full detail and rationale in [PLAN.md](PLAN.md). At a glance:

- [x] **Phase 0 — Foundations.** Module, layout, engine interface + domain
      types, moby skeleton, test fake, cobra root, CI.
- [x] **Phase 1 — Engine + one-shot CLI** (`v0.1.0`). Read paths
      (list/inspect/logs/stats/events); `gantry ps`, `images`, `logs`.
- [x] **Phase 2 — TUI, read-only** (`v0.2.0`). Bubbletea panes + live event loop.
- [x] **Phase 3 — Mutations** (`v0.3.0`). start/stop/restart/kill/remove/prune,
      network create/connect; shared confirmation layer.
- [x] **Phase 4 — HTTP API + React, read-only** (`v0.4.0`). REST + SSE, embedded
      frontend, 127.0.0.1 by default.
- [x] **Phase 5 — Web mutations** (`v0.5.0`). Toasts, confirm dialogs, bulk
      select.
- [x] **Phase 6 — Exec terminal** (`v0.6.0`). xterm.js ↔ WebSocket ↔ exec, TUI
      exec, resize propagation.
- [x] **Phase 7 — Container create** (`v1.0.0`). Minimal → progressive
      disclosure, `--from <container>`, export as `docker run`/compose.
- [x] **Phase 8 — Release polish.** goreleaser, Homebrew, Docker image, release
      workflow, install docs. (Terminal GIF + docs site still to do.)

## License

[MIT](LICENSE) © 2026 Saroj Bhattarai
