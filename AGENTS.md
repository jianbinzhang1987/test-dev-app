# Repository Guidelines

## Project Structure & Module Organization
- `deploymaster-pro-wails/` contains the Wails (Go + Vue 3) desktop app.
  - Go backend entry points: `main.go`, `app.go`.
  - Core backend modules: `internal/node`, `internal/ssh`, `internal/topology`.
  - Frontend UI: `deploymaster-pro-wails/frontend/src/` (Vue + Vite).
- `assets/` holds design or placeholder assets.
- `bin/` contains local tooling artifacts.
- Docs: `需求文档.md` (product requirements) and `架构文档.md` (architecture).

## Build, Test, and Development Commands
Run from repo root unless noted.
- `cd deploymaster-pro-wails/frontend && npm install` — install frontend dependencies.
- `cd deploymaster-pro-wails && wails dev` — start Wails live development (hot reload; web dev port 34115).
- `cd deploymaster-pro-wails && wails build` — build a production desktop bundle (output in `deploymaster-pro-wails/build/bin/`).
- `cd deploymaster-pro-wails && go test ./...` — run Go unit tests (none currently in tree).
- `cd deploymaster-pro-wails/frontend && npm run build` — type-check and build the Vue frontend.

## Coding Style & Naming Conventions
- Go: standard `gofmt` formatting; package names are lower-case, short, and domain-specific (e.g., `internal/node`).
- Vue/TS: follow Vite + Vue 3 defaults; prefer `PascalCase` for components and `camelCase` for composables (e.g., `useNodeService.ts`).
- Filenames: use descriptive, lowercase names for Go files; avoid abbreviations unless conventional.

## Testing Guidelines
- Go tests should live next to source with `_test.go` suffix (e.g., `internal/node/store_test.go`).
- No JS test runner is configured; add one if frontend unit tests are required.
- When adding tests, keep them deterministic and free of external network dependencies.

## Commit & Pull Request Guidelines
- Commit messages follow a conventional style (e.g., `feat: ...`, `fix: ...`). Keep the subject short and action-oriented.
- PRs should include: a concise summary, steps to test (commands), and screenshots for UI changes.
- Link relevant docs or issues when touching `需求文档.md` or `架构文档.md`.

## Configuration & Data Notes
- Project settings live in `deploymaster-pro-wails/wails.json`.
- Node data persists at `~/.deploymaster/nodes.json` (macOS/Linux) or `%USERPROFILE%\.deploymaster\nodes.json` (Windows).
