# Golang API Server Template

A minimal Go HTTP API server template with Cobra CLI, structured logging (slog), graceful shutdown, and Kubernetes-friendly health checks. Use it as a starting point for new services or as a reference for patterns (middleware, config, versioning).

## Features

- **HTTP server** — Configurable port, readiness-aware shutdown
- **Endpoints** — `GET /healthz` (liveness/readiness), `GET /version` (JSON build info)
- **CLI** — [Cobra](https://github.com/spf13/cobra) with global and server flags
- **Config** — [Viper](https://github.com/spf13/viper): flags, env vars, optional `project.env` / `.env`
- **Logging** — [slog](https://pkg.go.dev/log/slog) JSON output, request ID and level configurable
- **Graceful shutdown** — SIGINT/SIGTERM, readiness drain delay for k8s
- **Container** — Multi-stage Dockerfile (distroless), optional Task-based build/run

## Prerequisites

- **Go 1.26+** (see `go.mod`)
- Optional: [Task](https://taskfile.dev/) for `task build`, `task run/server`, etc.
- Optional: Docker for container image and `task kind/deploy`

## Creating a Project from This Template

Using [gonew](https://pkg.go.dev/golang.org/x/tools/cmd/gonew):

```bash
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/nikitamarchenko/golang-template-api-server@main your.domain/myprog
```

Then set your module name and image in `project.env` (and update any references in Dockerfile/taskfile if you rename the binary).

## Configuration

Settings can be provided by:

- **Flags** — e.g. `--http-port 8080`, `--log-level INFO`
- **Environment** — same keys with `_` instead of `-` (e.g. `HTTP_PORT`, `LOG_LEVEL`)
- **Files** — when using Task, `project.env` and `.env` are loaded (see `taskfile.yaml`)

### Main options

| Flag / Env              | Default | Description |
|-------------------------|---------|-------------|
| `--http-port` / `HTTP_PORT` | 8080 | HTTP listen port |
| `--log-level` / `LOG_LEVEL` | INFO | Log level (DEBUG, INFO, WARN, ERROR) |
| `--log-show-source`      | false | Add source file/line to log lines |
| `--http-readiness-probe-period-seconds` | 10 | Used for shutdown drain delay (match k8s probe) |
| `--allow-root-user`     | false | Allow running as root (e.g. in some containers) |

Example `project.env` (used by Task; adjust for your repo):

```env
PROJECT_NAME=golang-template-api-server
IMG=localhost/golang-template-api-server
```

## Building and Running

### Run with Go

```bash
go build -o bin/app .
./bin/app server
```

With options:

```bash
./bin/app server --http-port 8482 --log-level DEBUG
```

### Run with Task

If you use [Task](https://taskfile.dev/):

```bash
task run/server
```

With extra args:

```bash
task run/server -- --http-port 8482 --log-level DEBUG
```

Other useful tasks:

- `task build` — build binary to `/tmp/bin/$(PROJECT_NAME)`
- `task run/server/live` — run with live reload (air)
- `task test` — run tests
- `task lint` — golangci-lint
- `task build/image` — build Docker image

## API Endpoints

| Method | Path      | Description |
|--------|-----------|-------------|
| GET    | `/healthz` | Health/readiness. Returns 200 when ready, 503 when shutting down. |
| GET    | `/version` | JSON version info (git version, commit, build date, Go version, platform). |

Example:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/version | jq
```

## Docker

Build the image (Task uses `project.env` for `IMG` and version LDFLAGS):

```bash
task build/image
```

Or with Docker directly:

```bash
docker build -t myapp:latest .
docker run -p 8080:8080 myapp:latest
```

The Dockerfile produces a minimal image (distroless). Default command is `server`; the app listens on port 8080.

## Project Layout

```
.
├── cmd/
│   ├── root.go      # Cobra root, global flags (log-level, log-show-source)
│   ├── server.go    # server subcommand and config binding
│   └── cmd.go       # exit code constants
├── internal/
│   ├── server/      # HTTP server, routes, middleware, handlers
│   └── version/     # version info (injected at build via ldflags)
├── main.go
├── go.mod / go.sum
├── Dockerfile       # multi-stage, distroless
├── taskfile.yaml    # build, lint, test, run, Docker, kind deploy
└── project.env      # PROJECT_NAME, IMG (for Task and deployment)
```

## Development

- **Format and deps:** `task tidy` or `go fmt ./...` and `go mod tidy`
- **Lint:** `task lint` (installs golangci-lint if needed)
- **Tests:** `task test` or `go test -race -buildvcs ./...`
- **Run:** `task run/server` or `./bin/your-binary server`
- **Live reload:** `task run/server/live` (uses air)

Version info for `/version` is set at build time; see `hack/version.sh` and `task build/release` / `task build/image`.

## License

See repository license file.
