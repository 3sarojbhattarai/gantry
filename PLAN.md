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

## Phase 1 — The engine + one-shot CLI → `v0.1.0`

Implement the `docker` package read paths, then wire the thinnest possible
renderer on top: one-shot CLI commands. The CLI comes before the TUI because it
proves the engine with no UI complexity in the way — if `gantry ps` is correct,
everything downstream inherits that correctness.

- [ ] Read paths in `moby.go`: list/inspect containers, images, networks, volumes
- [ ] Logs with `stdcopy` demux
- [ ] `stats` sampling → `docker.Stats`
- [ ] `/events` stream → `docker.Event`
- [ ] `gantry ps`, `gantry images`, `gantry logs <id>`
- [ ] Table tests against `fakedocker.Fake`
- [ ] Dedicated log-demux tests with hand-crafted frame headers

---

## Phase 2 — TUI, read-only → `v0.2.0`

Bubbletea. Pane list on the left (containers/images/networks/volumes), detail on
the right, log tail at the bottom. Keybindings borrowed from lazydocker so muscle
memory transfers.

The real work is the event loop: `/events` and `/stats` as Bubbletea `tea.Cmd`s
feeding messages into `Update`. Get this right and the UI is live, not polled.
Budget most of the phase here — it's the first genuinely unfamiliar thing if you
haven't used Charm before. At this point the tool is useful daily, which matters
for motivation.

- [ ] Pane list + detail view + log tail layout
- [ ] lazydocker-style keybindings
- [ ] `/events` and `/stats` as `tea.Cmd`s driving `Update`
- [ ] Live refresh (no polling)

---

## Phase 3 — Mutations → `v0.3.0`

start, stop, restart, kill, remove, prune (images/containers/volumes/networks),
network create and connect/disconnect.

Build the confirmation layer **once, in the engine**, not per-renderer:
destructive ops take an explicit flag, and each renderer decides how to obtain
consent — `--force` in CLI, modal in TUI, dialog in web. `prune` gets a dry-run
that reports what would be reclaimed. This is the first real "usable release"
moment and a good public checkpoint — a feature-complete lazydocker equivalent.

- [ ] Mutation methods on the engine interface
- [ ] start/stop/restart/kill/remove
- [ ] prune (containers/images/volumes/networks) with dry-run
- [ ] network create + connect/disconnect
- [ ] Shared consent model: explicit destructive flag, per-renderer consent

---

## Phase 4 — HTTP API + React, read-only → `v0.4.0`

Handlers over the same engine. REST for lists and inspects, SSE for the event and
stats streams. React + Vite + Tailwind CSS + TanStack Query, with `EventSource`
invalidating queries so the UI stays live without polling.

Bind `127.0.0.1` by default; overriding it requires an explicit flag with a loud
warning. Embed `web/dist` via `embed.FS` behind a build tag so `go run` still
works without a frontend build.

- [ ] REST handlers for lists + inspects
- [ ] SSE endpoints for events + stats
- [ ] React + Vite + Tailwind CSS + TanStack Query app
- [ ] `EventSource` → query invalidation
- [ ] `127.0.0.1` default bind + explicit-override warning
- [ ] `embed.FS` behind a build tag

---

## Phase 5 — Web mutations → `v0.5.0`

Mostly mechanical — the engine already does the work. The real content is UI:
optimistic updates, error toasts, confirm dialogs, and bulk selection for
"remove all stopped".

- [ ] Mutation endpoints
- [ ] Optimistic updates + error toasts
- [ ] Confirm dialogs (reusing the engine consent model)
- [ ] Bulk selection

---

## Phase 6 — Exec terminal → `v0.6.0`

`xterm.js` ↔ WebSocket ↔ hijacked `ContainerExecAttach`. Do the TUI version too —
Bubbletea can suspend and hand the terminal over, which is simpler than the
browser path. Don't forget resize propagation; it's the thing everyone ships
broken.

- [ ] WebSocket exec handler over hijacked attach
- [ ] xterm.js frontend terminal
- [ ] TUI exec (suspend + hand over terminal)
- [ ] Resize propagation (both surfaces)

---

## Phase 7 — Container create → `v1.0.0`

The hard part. Sequence it:

1. **Minimal path first:** image + name + ports + env + restart policy. Covers
   80% of real use.
2. **Progressive disclosure** for the rest: mounts, networks, resource limits,
   capabilities, healthcheck, labels.
3. **`--from <container>`:** read an existing container's config as the form's
   starting point. Cheap to build, disproportionately useful — makes "clone and
   tweak" the primary workflow instead of filling a giant form from scratch.
4. **Export as `docker run` / compose fragment.** Turns the form into a teaching
   tool and sidesteps the trust problem — users see exactly what gantry would do.

For the CLI, don't rebuild `docker run`'s flag surface. Accept a small YAML file
or the `--from` path.

- [ ] Minimal create: image, name, ports, env, restart policy
- [ ] Progressive disclosure: mounts, networks, limits, caps, healthcheck, labels
- [ ] `--from <container>` prefill
- [ ] Export as `docker run` / compose fragment
- [ ] CLI create via YAML file / `--from`

---

## Phase 8 — Release polish

- [ ] goreleaser for multi-platform binaries
- [ ] Homebrew tap
- [ ] Docker image (mount the socket; document the security caveat prominently)
- [ ] README with a terminal GIF and a UI screenshot
- [ ] Docs site

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
