# Contributing to Gantry

Thanks for your interest. Gantry is early and moving through a
[phased plan](PLAN.md); the most useful contributions right now align with the
current phase rather than jumping ahead.

## Prerequisites

- **Go 1.26+** — `go version`
- **A running Docker daemon** — needed to run gantry and the integration tests
  (unit tests do not need one)
- **golangci-lint** — matches CI; install from
  [golangci-lint.run](https://golangci-lint.run/welcome/install/)
- **Node.js 20+** — only for the frontend, from Phase 4 onward

## Local development

Common tasks are in the [Makefile](Makefile) — run `make help` for the list.

```bash
# clone
git clone https://github.com/3sarojbhattarai/gantry
cd gantry

# build, test, run
make build            # -> bin/gantry
make test             # unit tests, no daemon
make run ARGS=version # run with args

# raw equivalents, if you prefer
go build ./... && go test ./...
go run ./cmd/gantry --version
```

### The full pre-push check

Run what CI runs before opening a pull request:

```bash
make check   # fmt-check + vet + lint + test
```

Equivalent to `gofmt -l .` (should print nothing), `go vet ./...`,
`golangci-lint run`, and `go test ./...`.

### Integration tests (Phase 1+)

These run against a real daemon and are guarded by a build tag:

```bash
go test -tags integration ./...
```

Fixtures are `alpine sleep 300` containers created in a dedicated
`gantry-test-*` namespace; teardown removes anything matching that prefix. Never
point integration tests at a daemon whose containers you care about.

### Frontend (Phase 4+)

The frontend is React + Vite + **Tailwind CSS**. Keep styling utility-first —
don't add a second CSS framework.

```bash
make web-install   # npm install
make web-dev       # Vite dev server
make web-test      # Vitest
make web-build     # produces web/dist, embedded by the Go binary
```

## Architecture you must respect

Before writing code, read [CLAUDE.md](CLAUDE.md). The one hard rule:

> **moby's SDK types never leak past `internal/docker`.**

`internal/docker` owns gantry's domain types and the `docker.Client` interface.
The moby SDK is used only inside `internal/docker/moby.go`. Everything else — CLI,
TUI, API — depends only on the interface and gantry's structs. If a change makes
`github.com/docker/docker` an import outside `internal/docker`, it will be sent
back: translate into a domain type inside the engine instead.

This is also what makes contributions testable: write logic against the
`docker.Client` interface and test it with `fakedocker.Fake` — no daemon needed.

## Conventions

- **Formatting:** `gofmt`; lint clean under `golangci-lint`.
- **Errors:** wrap with context and a `gantry:` prefix; expose sentinels
  (`docker.ErrNotFound`) for callers to branch on with `errors.Is`.
- **Interface assertions:** add `var _ Client = (*T)(nil)` to every
  implementation.
- **Scope:** keep a change and its commits focused on one concern, within the
  current phase.

## Pull requests

1. Branch off `main` — `git checkout -b phaseN/short-description`.
2. Make the change, with tests. New engine behavior gets a `fakedocker`-backed
   test; log-demux code gets frame-header tests.
3. Run the full pre-push check above; it must be green.
4. Update [PLAN.md](PLAN.md) checkboxes and the [README.md](README.md) roadmap if
   your change completes a planned item.
5. Open the PR with a clear description of what and why, and which phase item it
   advances.

## Reporting bugs and proposing features

Open an issue. For bugs, include your OS, Go version, Docker version
(`docker version`), and the exact `gantry` command and output. For features,
say which phase it fits — or make the case for reordering the plan.

## License

By contributing you agree your contributions are licensed under the
[MIT License](LICENSE).
