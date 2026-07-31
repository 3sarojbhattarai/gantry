# Gantry — Build Plan

The phased roadmap from empty repo to `v1.0.0` and release. Each phase ends at a
tagged, shippable point.

## Two decisions that shape everything

Made in Phase 0 because they are expensive to reverse:

1. **Own the domain types.** `internal/docker` is an interface (`docker.Client`)
   over gantry's own structs — `Container`, `ImageSummary`, `Network`,
   `Volume`, `Stats`, `Event`. It is implemented with the official SDK to move
   fast, but moby's `types.ContainerJSON` and friends **never** leak past this
   package into the TUI or API. This buys the option to drop the SDK for
   hand-rolled HTTP later, and makes everything testable against a fake.

2. **Directory layout**, fixed up front so imports don't churn:

   ```
   cmd/gantry/          main + cobra root
   internal/docker/     interface, moby impl, domain types
   internal/cli/        one-shot commands
   internal/tui/        bubbletea
   internal/api/        handlers, SSE, ws exec
   internal/web/        embed.go
   web/                 vite + react
   ```

---

## Phase 0 — Foundations — ✅ done

Repo, module path (`github.com/3sarojbhattarai/gantry`), license (MIT), CI
skeleton, the layout above, and the two decisions locked in.

Ships a compiling skeleton: the `docker.Client` interface, domain types, a
moby-backed implementation whose read methods return `ErrNotImplemented`, an
in-memory `fakedocker.Fake`, the cobra root wired to `--version`, and GitHub
Actions CI.

- [x] `go.mod`, layout, MIT license, `.gitignore`
- [x] `internal/docker`: interface, domain types, `errors.go`, moby skeleton
- [x] `internal/docker/fakedocker`: in-memory fake + pattern-setting test
- [x] `cmd/gantry` + `internal/cli`: cobra root, `version`
- [x] `.github/workflows/ci.yml`: build, vet, test, lint

---

## Phase 1 — The engine + one-shot CLI → `v0.1.0` — ✅ done

Implement the `docker` package read paths, then wire the thinnest possible
renderer on top: one-shot CLI commands. The CLI comes before the TUI because it
proves the engine with no UI complexity in the way — if `gantry ps` is correct,
everything downstream inherits that correctness.

- [x] Read paths in `moby.go`: list/inspect containers, images, networks, volumes
- [x] Logs with demux (own frame-parser in `logs.go`, not the SDK helper)
- [x] `stats` sampling → `docker.Stats` (CPU/mem math in `computeStats`)
- [x] `/events` stream → `docker.Event`
- [x] `gantry ps`, `gantry images`, `gantry logs <id>`
- [x] Table tests against `fakedocker.Fake` (engine + CLI consumers)
- [x] Dedicated log-demux tests with hand-crafted frame headers
- [x] Integration tests (`//go:build integration`) against a real daemon + CI job

---

## Phase 2 — TUI, read-only → `v0.2.0` — ✅ done

Bubbletea. Tabbed list on the left (containers/images/networks/volumes), detail on
the right, log tail at the bottom. Keybindings borrowed from lazydocker so muscle
memory transfers. Bare `gantry` launches it.

The real work is the event loop: `/events` and `/stats` as Bubbletea `tea.Cmd`s
feeding messages into `Update`. Get this right and the UI is live, not polled.

- [x] Tabbed list + detail view + log tail layout
- [x] lazydocker-style keybindings (`j/k`, `tab`/`shift+tab`, `1-4`, `r`, `?`, `q`)
- [x] `/events` and `/stats` as `tea.Cmd`s driving `Update`
- [x] Live refresh (no polling); container log streaming in the bottom pane
- [x] Selection-scoped streams with generation guard against stale messages
- [x] Reducer unit tests against `fakedocker.Fake`

---

## Phase 3 — Mutations → `v0.3.0` — ✅ done

start, stop, restart, kill, remove, prune (images/containers/volumes/networks),
network create and connect/disconnect.

The confirmation layer lives **once, in the engine**: destructive methods take a
`Consent` value (zero value = not granted) and return `ErrConsentRequired`
without it. Each renderer obtains consent its own way — `--force` in the CLI, a
confirmation modal in the TUI, a dialog on the web (Phase 5). `prune` has a
dry-run that previews candidates (containers and images, computed from list
data; volumes/networks report `ErrDryRunUnsupported`).

- [x] Mutation methods on the engine interface (moby impl + fake)
- [x] start/stop/restart/kill; remove container/image/network/volume
- [x] prune (containers/images/volumes/networks) with dry-run
- [x] network create + connect/disconnect
- [x] Shared consent model: `Consent`/`Confirm()`, `ErrConsentRequired`
- [x] CLI commands (`--force` = consent); TUI keybindings + confirmation modal
- [x] Consent-enforcement + mutation tests (engine, fake, CLI, TUI)

---

## Phase 4 — HTTP API + React, read-only → `v0.4.0` — ✅ done

Handlers over the same engine (`internal/api`). REST for lists and inspects, SSE
for the event/stats/log streams. React + Vite + Tailwind CSS + TanStack Query,
with `EventSource` invalidating queries so the UI stays live without polling.
`gantry serve` runs it.

Binds `127.0.0.1` by default; a non-loopback `--addr` prints a loud warning. The
built frontend is embedded via `embed.FS` behind the `embed` build tag, so plain
`go build` works without a Node toolchain and release builds ship one binary.

- [x] REST handlers for lists + inspects (+ health)
- [x] SSE endpoints for events, stats, and logs
- [x] React + Vite + Tailwind CSS + TanStack Query app
- [x] `EventSource` → query invalidation (live, no polling)
- [x] `127.0.0.1` default bind + loud non-loopback warning (`gantry serve`)
- [x] `embed.FS` behind the `embed` build tag; placeholder otherwise
- [x] `httptest` handler tests (fake-backed) + Vitest for frontend helpers

## Phase 5 — Web mutations → `v0.5.0` — ✅ done

The engine already does the work; this is UI. Mutation endpoints, optimistic-feel
refresh (invalidate on success), error toasts, a promise-based confirm dialog
(the web renderer's consent), and bulk selection for "remove selected" /
"remove all stopped".

- [x] Mutation endpoints (start/stop/restart/kill, remove, prune, network ops)
- [x] HTTP-level consent (`?confirm=true` → `docker.Confirm()`; 409 without)
- [x] Toasts + refresh-on-success; confirm dialog for destructive ops
- [x] Bulk selection (remove selected, remove all stopped) + prune with preview

---

## Phase 6 — Exec terminal → `v0.6.0` — ✅ done

`xterm.js` ↔ WebSocket ↔ hijacked `ContainerExecAttach`. The engine exposes an
`ExecSession` (io.ReadWriteCloser + Resize) that hides the SDK's hijacked
connection. `internal/termexec` wires a session to a real terminal (raw mode +
SIGWINCH) and is shared by `gantry exec` and the TUI's suspend-and-attach.

- [x] Engine `ExecSession` + `ContainerExec` (moby + fake)
- [x] `gantry exec <container> <cmd>` — raw mode, resize propagation
- [x] WebSocket exec handler (binary I/O, JSON resize control)
- [x] xterm.js frontend terminal with fit + resize
- [x] TUI exec (`e`) — `tea.Exec` suspend + hand over via `termexec`

---

## Phase 7 — Container create → `v1.0.0` — ✅ core done

The engine owns a renderer-agnostic `CreateSpec`; the CLI reads it from YAML or
`--from`, the web form edits it with progressive disclosure, and both can export
it as a `docker run` command or compose fragment (rendered by the same engine
functions, so they agree).

- [x] Minimal create: image, name, command, ports, env, restart policy
- [x] Progressive disclosure: volumes, networks, labels, working dir, user
- [x] `--from <container>` prefill (`SpecFromContainer`) — clone and tweak
- [x] Export as `docker run` / compose fragment (`SpecToDockerRun`/`SpecToCompose`)
- [x] CLI `gantry create --file <yaml> | --from <container> [--export run|compose]`
- [x] Web create form (modal) with clone-from, export preview, and create
- [ ] Deferred progressive fields: resource limits, capabilities, healthcheck

---

## Phase 8 — Release polish — ✅ mostly done

- [x] goreleaser for multi-platform binaries (`.goreleaser.yaml`, linux/darwin/windows × amd64/arm64)
- [x] Homebrew tap (goreleaser `brews` → `3sarojbhattarai/homebrew-tap`)
- [x] Docker image (multi-stage `Dockerfile` + goreleaser `dockers`; socket-mount security caveat in README) — verified: 22 MB, serves the embedded UI
- [x] Release workflow (`.github/workflows/release.yml`, tag-triggered goreleaser)
- [x] README install section (brew / go / binary / docker) + security warning
- [ ] Terminal GIF + UI screenshot in the README (manual — needs a recording)
- [ ] Docs site (Docusaurus — manual)

Requires two one-time setup steps before the first `git tag vX.Y.Z && git push --tags`:
a `3sarojbhattarai/homebrew-tap` repo, and a `HOMEBREW_TAP_GITHUB_TOKEN` secret
with write access to it.

---

## Testing strategy

Decided now, because retrofitting is miserable:

- **Engine logic** — table tests against `fakedocker.Fake`. Fast, no daemon.
- **Integration** — build tag `//go:build integration`, run against a real
  daemon. Fixtures are `alpine sleep 300` containers in a dedicated
  `gantry-test-*` namespace, with a teardown that nukes anything matching that
  prefix. GitHub Actions runners have Docker, so this runs in CI.
- **Log demux** — dedicated tests with hand-crafted frame headers. Off-by-one
  errors here are subtle and produce garbled output rather than crashes.
- **Frontend** — Vitest for query/reducer logic. Skip E2E; not worth the
  maintenance at this scale.
