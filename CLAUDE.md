# CodeCanary

AI-powered code review for GitHub pull requests.

## Project structure

```
cmd/             # CLI commands (cobra)
  root.go        # Root "codecanary" command
  init.go        # codecanary init — setup wizard
  review.go      # codecanary review <pr> — run a review
  generate.go    # codecanary review generate — generate config from repo
internal/
  review/        # Review engine (prompts, GitHub API, triage, formatting)
  auth/          # OAuth PKCE flow, GitHub App installation
worker/          # Cloudflare Worker — OIDC token proxy (TypeScript)
main.go          # Entry point
```

## Build & run

```sh
go build -o codecanary .
go run .
```

Version is set via ldflags: `-X main.version=v{version}`

## Key dependencies

- `spf13/cobra` — CLI framework
- `bmatcuk/doublestar` — glob pattern matching for ignore rules
- `gopkg.in/yaml.v3` — config parsing

## Architecture notes

- **Config** is `.codecanary.yml` at the repo root (flat file, no directory)
- **Review flow**: fetch PR data via `gh` CLI → build prompt → call Claude via `claude` CLI → parse findings → post as PR review
- **Incremental reviews**: on re-push, triage existing threads (Go-driven classifier), only re-evaluate changed threads via Claude (haiku), then review only new code
- **Dual marker detection**: reads both `codecanary:review` and legacy `clanopy:review` HTML markers for backward compatibility
- **Anti-hallucination**: explicit file allowlist, line validation against diff, max finding distance threshold
- **Worker** (`worker/`): OIDC token exchange proxy on Cloudflare Workers — verifies GitHub Actions OIDC token, returns GitHub App installation token

## Rules

- **Minimize shell code.** The GitHub Action (`alansikora/codecanary-action`) should be kept thin. All logic must live in Go.
- No automated tests exist yet (only config unit tests). Be careful with refactors.
