# AGENTS.md - iperf-rpc

Go + Vue 3 iperf3 network speedtest RPC, deployed on K8s for pod-to-pod throughput testing across clusters.

## Key commands

```bash
# Go build
go build ./...

# Tests (only iperf package; avoid go test ./... which may hang on network)
go test ./internal/iperf/ -v

# Swagger docs generation (required before commit if API changes)
swag init -g cmd/server/main.go -o docs

# Frontend build
cd web && pnpm install --frozen-lockfile && pnpm run build

# goreleaser local snapshot (binary only)
goreleaser build --snapshot --clean

# goreleaser local snapshot with Docker (needs web/dist pre-built)
cd web && pnpm run build && cd .. && goreleaser release --snapshot --clean --skip=validate
```

## GOPROXY

Network may not reach `proxy.golang.org`. Set `GOPROXY=https://goproxy.cn,direct`.

## Architecture

Go backend (gin + swaggo) at `cmd/server/`, Vue 3 frontend at `web/`. The binary serves the frontend from `./web/dist/` via Gin static file routing.

Entrypoint: `cmd/server/main.go`
Router: `internal/api/router.go` (5 API endpoints + swagger + SPA fallback)
Handler: `internal/api/handler.go`
iperf3 exec: `internal/iperf/executor.go`

## Docker / CI gotchas

- **Dockerfile is COPY-only**: no build stages inside it. CI (and local dev) must pre-build the Go binary and `web/dist` before `docker build`. goreleaser provides the binary; `extra_files: web/dist` in `.goreleaser.yaml` copies the frontend into the Docker context.
- **Frontend build in CI**: the release workflow (`release.yml`) runs `pnpm run build` in `web/` BEFORE goreleaser. Without this, the Docker context will lack `web/dist` and goreleaser fails.
- **riscv64**: Go binary only. No Docker image (alpine lacks riscv64 support).
- **Version tags**: goreleaser's `{{ .Version }}` strips the `v` prefix, so Git tag `v0.1.2` produces image tag `0.1.2` (not `v0.1.2`).
- `latest` manifest auto-updates on every release.

## Branch & release workflow

```
dev → no-ff merge → master → git tag vX.Y.Z → push tags → CI builds & pushes to ghcr.io
```

- Commits: conventional (`feat:`, `fix:`, `docs:`, `chore:`)
- Commit messages append: `Assisted by: DeepSeek V4 Pro + OpenCode 1.14.30`
- All development on `dev`; `master` is stable/release only

## K8s

Manifests at `k8s/manifests.yaml`: two Deployments + two NodePort Services:
- `iperf-rpc-server`: API (8080:30080) + iperf3 (5201:30081)
- `iperf-rpc-client`: API only (8080:30082)

## Docker Compose (local dev)

`docker compose up -d` starts two instances (server-a on :8080/:5201, server-b on :8081/:5202). Requires `web/dist` pre-built for the Dockerfile to succeed.
