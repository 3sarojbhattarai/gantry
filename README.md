# Gantry

A single-binary Docker manager with a terminal UI and a web UI, built in Go.

Gantry is a docker tool for inspecting and managing a local Docker daemon — 
list and inspect containers, images, networks and volumes; tail logs; watch 
live stats and events; start, stop, prune and create containers — from either 
your terminal or a browser, backed by the same engine.

> **Status:** early development. Phase 0 (foundations) is complete: the engine
> interface, domain types, a moby-backed skeleton, a test fake, the cobra root,
> and CI. Working commands land in Phase 1. See [PLAN.md](PLAN.md) for the full
> roadmap.

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
| TUI        | [Bubbletea](https://github.com/charmbracelet/bubbletea) (Phase 2) |
| Web API    | Go stdlib `net/http`, Server-Sent Events, WebSocket (Phase 4)     |
| Frontend   | React + Vite + Tailwind CSS + TanStack Query, xterm.js (Phase 4/6) |
| Release    | goreleaser, Homebrew tap, Docker image (Phase 8)                  |
| CI         | GitHub Actions (build, vet, test, golangci-lint)                  |

## Getting started

Requires Go 1.26+ and a running Docker daemon.

```bash
git clone https://github.com/3sarojbhattarai/gantry
cd gantry
go run ./cmd/gantry --version
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full local-development workflow.

## Roadmap

Full detail and rationale in [PLAN.md](PLAN.md). At a glance:

- [x] **Phase 0 — Foundations.** Module, layout, engine interface + domain
      types, moby skeleton, test fake, cobra root, CI.
- [ ] **Phase 1 — Engine + one-shot CLI** (`v0.1.0`). Read paths
      (list/inspect/logs/stats/events); `gantry ps`, `images`, `logs`.
- [ ] **Phase 2 — TUI, read-only** (`v0.2.0`). Bubbletea panes + live event loop.
- [ ] **Phase 3 — Mutations** (`v0.3.0`). start/stop/restart/kill/remove/prune,
      network create/connect; shared confirmation layer.
- [ ] **Phase 4 — HTTP API + React, read-only** (`v0.4.0`). REST + SSE, embedded
      frontend, 127.0.0.1 by default.
- [ ] **Phase 5 — Web mutations** (`v0.5.0`). Optimistic updates, toasts, bulk
      select.
- [ ] **Phase 6 — Exec terminal** (`v0.6.0`). xterm.js ↔ WebSocket ↔ exec, TUI
      exec, resize propagation.
- [ ] **Phase 7 — Container create** (`v1.0.0`). Minimal → progressive
      disclosure, `--from <container>`, export as `docker run`/compose.
- [ ] **Phase 8 — Release polish.** goreleaser, Homebrew, Docker image, docs.

## License

[MIT](LICENSE) © 2026 Saroj Bhattarai
