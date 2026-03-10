# hCTF2 CLI Design

**Date**: 2026-03-10
**Status**: Implemented (v0.7.0)

## Overview

Add a full CLI to the hCTF2 binary, covering both admin operations (challenge/competition/user management) and participant operations (browsing challenges, submitting flags, checking scoreboard). The existing server becomes a `serve` subcommand. All CLI commands communicate with a running hCTF2 server via its existing REST API — no direct DB access.

## Command Tree

```
hctf2
├── serve                          # current server (all existing flags move here)
├── login                          # prompt/flags → store JWT to config
├── logout
├── status                         # show configured server + auth state
│
├── challenge
│   ├── list                       # table: id, title, category, points, solved?
│   ├── get <id>                   # detail + questions (glamour markdown)
│   ├── browse                     # bubbletea interactive picker → submit
│   ├── create                     # admin: huh form or --flags
│   ├── update <id>
│   └── delete <id>
│
├── flag submit <question-id> <flag>
├── hint
│   ├── list <question-id>
│   └── unlock <id>
│
├── team
│   ├── list / get <id>
│   ├── create / join <invite-code>
│   └── delete <id>                # admin only
│
├── competition
│   ├── list / get <id>
│   ├── create / start <id> / end <id>
│   └── scoreboard <id>
│
├── user                           # admin only
│   ├── list / promote <id> / delete <id>
│
└── version / info                 # replaces --version / --info flags
```

Running bare `hctf2` (no subcommand) prints help — no silent default behavior.

## Auth & Config

Config file: `~/.config/hctf2/config.yaml` (overridable via `HCTF2_CONFIG` env var)

```yaml
server: http://localhost:8090
token: eyJhbGci...
token_expires: 2026-03-11T10:00:00Z
```

- `hctf2 login` — prompts via `huh` if `--server`/`--email`/`--password` flags are missing. Calls `POST /api/auth/login`, stores JWT locally.
- Token sent as `Cookie: auth_token=<token>` on every request (mirrors browser behavior).
- `hctf2 status` — shows server URL, logged-in user (from token claims), and expiry. No network call.
- `--server` global flag — overrides config for a single invocation.
- Expired token: commands fail with clear message; no silent refresh.
- Admin detection: server returns 403; CLI surfaces `error: admin privileges required`.

## Output & TUI

**TTY detection** at startup via `term.IsTerminal(os.Stdout.Fd())`.

| Flag | Behavior |
|------|----------|
| (TTY detected) | Rich output: lipgloss tables, glamour markdown, huh prompts |
| `--json` | JSON to stdout (scriptable, CI-friendly) |
| `--quiet` | Minimal: IDs on create, "ok" on success |

**Libraries and usage:**

- **lipgloss** — tables for `challenge list`, `team list`, `scoreboard`, `user list`. Width-adaptive columns.
- **glamour** — renders markdown descriptions in `challenge get <id>`.
- **huh** — interactive forms for `login` (missing flags), `challenge create` (missing flags). Skipped when not a TTY — missing required args → error instead.
- **bubbletea** — one dedicated interactive mode: `hctf2 challenge browse`. Arrow keys to navigate, `/` to filter, `enter` to view, `s` to submit flag inline. Auto-disabled when not a TTY.

Errors always go to stderr. Exit code 1 on any error.

## Package Structure

```
hctf2/
├── main.go                  # entry: cobra root command dispatch only
├── cmd/
│   ├── root.go              # cobra root, global flags (--server, --json, --quiet)
│   ├── serve.go             # wraps current server logic (flags move here)
│   ├── login.go
│   ├── status.go
│   ├── challenge.go
│   ├── flag.go
│   ├── hint.go
│   ├── team.go
│   ├── competition.go
│   └── user.go
│
├── internal/
│   ├── client/              # HTTP client wrapping existing REST API
│   │   ├── client.go        # base: server URL, token, do()
│   │   ├── challenges.go
│   │   ├── teams.go
│   │   ├── competitions.go
│   │   └── auth.go
│   ├── tui/                 # charmbracelet components
│   │   ├── table.go         # lipgloss table renderer
│   │   ├── browse.go        # bubbletea challenge browser
│   │   └── theme.go         # shared lipgloss styles
│   ├── config/              # config read/write
│   │   └── config.go
│   │
│   ├── auth/                # unchanged
│   ├── database/            # unchanged
│   ├── handlers/            # unchanged
│   ├── models/              # unchanged
│   └── views/               # unchanged
```

**Key constraints:**
- `internal/client/` speaks HTTP only — zero knowledge of server internals.
- `cmd/serve.go` is the current server setup extracted into a cobra `RunE`.
- New dependencies: `cobra`, `bubbletea`, `lipgloss`, `huh`, `glamour` — all pure Go, no CGO.

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | Subcommand structure, --help, shell completion |
| `github.com/charmbracelet/bubbletea` | Interactive challenge browser |
| `github.com/charmbracelet/lipgloss` | Table and styled output |
| `github.com/charmbracelet/huh` | Interactive forms for prompts |
| `github.com/charmbracelet/glamour` | Markdown rendering in terminal |
| `golang.org/x/term` | TTY detection |

## Non-Goals

- Direct DB access from CLI (HTTP only)
- Refresh token / session renewal (out of scope for now)
- Real-time WebSocket feeds from CLI
- Shell completion generation (cobra provides it for free via `hctf2 completion`)
