# CLAUDE.md

Guidance for Claude Code (and any AI agent) working in this repository.

## What this is

Gantry is a single-binary Docker manager with a terminal UI (Bubbletea) and a web
UI (React), both driven by one Go engine. It is lazydocker-shaped: list/inspect
containers, images, networks, volumes; logs; live stats/events; mutations; and
container create. See [README.md](README.md) for the overview and
[PLAN.md](PLAN.md) for the phased roadmap and current status.

## The one architectural rule

**moby's SDK types never leak past `internal/docker`.**

`internal/docker` defines gantry's own domain types (`Container`, `ImageSummary`,
`Network`, `Volume`, `Stats`, `Event`) and a `Client` interface over them. The
moby SDK (`github.com/docker/docker/client`) is used only inside
`internal/docker/moby.go` to implement that interface. Every other package — CLI,
TUI, API — depends solely on the `docker.Client` interface and gantry's structs.

Why it matters:

- The engine stays swappable (hand-rolled HTTP is a future option).
- Everything above the engine is testable against `fakedocker.Fake` with no
  daemon running.

If you find yourself importing anything from `github.com/docker/docker` outside
`internal/docker`, stop — translate it into a gantry domain type inside the
engine instead.

## Layout

```
cmd/gantry/          main + cobra root (thin; delegates to internal/cli)
internal/docker/     engine: Client interface (client.go), domain types
                     (types.go), sentinel errors (errors.go), moby impl (moby.go)
  fakedocker/        in-memory Client for daemon-free tests
internal/cli/        one-shot cobra commands
internal/tui/        Bubbletea UI (Phase 2)
internal/api/        HTTP handlers, SSE, WebSocket exec (Phase 4)
internal/web/        embed.FS for the built frontend (Phase 4)
web/                 React + Vite + Tailwind CSS frontend (Phase 4)
```

## Conventions

- **Go 1.26.** Standard `gofmt`; CI runs `golangci-lint`.
- **Error wrapping.** Wrap with context and the `gantry:` prefix, e.g.
  `fmt.Errorf("gantry: pinging docker daemon: %w", err)`. Layers above branch on
  sentinels (`docker.ErrNotFound`) via `errors.Is`.
- **Interface assertions.** Every `Client` implementation carries a compile-time
  `var _ Client = (*T)(nil)` so signature drift fails the build, not a test.
- **Streaming reads** return a channel that closes when the context is cancelled
  or the stream ends. Callers own cancellation via context.
- **Version** is injected at build time via
  `-ldflags "-X .../internal/cli.version=vX.Y.Z"`; it defaults to `dev`.
- **Consent for destructive ops** lives in the engine (an explicit flag), not per
  renderer. Renderers choose how to obtain it — `--force`, TUI modal, web dialog.
  (Arrives in Phase 3; keep this in mind when touching mutation code.)

## Working with the plan

Work proceeds phase by phase (see [PLAN.md](PLAN.md)); each phase ends at a tagged,
shippable point. When implementing, stay within the current phase's scope — the
plan deliberately defers work to keep each step reviewable. Update the checklist
in `PLAN.md` and the roadmap in `README.md` when a phase item lands.

## Build, test, run

Prefer the Makefile (`make help` lists targets); the raw commands are equivalent.

```bash
make build                     # -> bin/gantry, version from git describe
make test                      # unit tests (fake-backed, no daemon)
make test-integration          # integration tests (needs a real daemon; Phase 1+)
make check                     # fmt-check + vet + lint + test (matches CI)
make run ARGS="version"        # go run with args

# raw equivalents
go build ./... && go vet ./... && go test ./...
go run ./cmd/gantry --version
```

Styling in the frontend is **Tailwind CSS** (Phase 4); keep styles utility-first
rather than introducing a separate CSS framework.

## Testing expectations

- **Engine + consumers:** table tests against `fakedocker.Fake`. No daemon.
- **Integration:** `//go:build integration`, real daemon, fixtures namespaced
  `gantry-test-*` with teardown removing anything matching that prefix.
- **Log demux:** dedicated tests with hand-crafted frame headers — off-by-one
  bugs there garble output silently.
- **Frontend:** Vitest for query/reducer logic; no E2E.

## Git

Do not commit or push unless asked. Branch off `main` before committing. This is a
solo project; keep history clean and commits scoped to a single concern.
