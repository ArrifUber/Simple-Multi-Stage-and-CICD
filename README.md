# Simple Multi-Stage & CI/CD — Go REST API

A small CRUD REST API written in Go using only the standard library (no database —
data lives in memory). The project shows how to package a Go app with a
**multi-stage Docker build** and ship it with a **simple GitHub Actions pipeline**:
test → build & push to Docker Hub → deploy on a self-hosted runner.

## Project structure

```
.
├── app/                     # Go source code (API + unit tests)
├── Dockerfile               # multi-stage build
├── deploy/
│   └── docker-compose.yaml  # runs the published image
└── .github/workflows/
    └── cicd.yml             # test → build & push → deploy
```

## How it works

### Multi-stage build

| Stage | Base image | Purpose |
| --- | --- | --- |
| `builder` | `golang:1.23-alpine` | Downloads modules and compiles a static binary (`CGO_ENABLED=0`) |
| final | `alpine:latest` | Copies only the binary + CA certificates and runs it |

The Go toolchain and source code are left behind, so the final image is tiny:
**~27 MB** compared to **~481 MB** for a single-stage build of the same app.

### Pipeline (`.github/workflows/cicd.yml`)

Triggered on every push to `main`:

```
push to main
      │
      ▼
[build-and-push]  runs-on: ubuntu-latest
  checkout → setup Go 1.23 → go test ./app/... → docker login
  → build & push  $DOCKERHUB_USERNAME/simple-golang-multi-stage:latest
      │
      ▼
[deploy]          runs-on: self-hosted
  checkout → docker pull → docker compose up -d → prune old images
```

## Prerequisites

- Go 1.23+
- Docker + Docker Compose (for container runs and the self-hosted runner)
- A Docker Hub account
- A machine that can stay online for the self-hosted runner

## Run locally (no Docker)

```bash
cd app
go run .
```

The API listens on `http://localhost:8080` (override with `PORT`).

## Run with Docker

```bash
# Build the image from the project root
docker build -t simple-golang-multi-stage .

# Run it
docker run -d --name go_backend -p 8080:8080 simple-golang-multi-stage
```

Or use the compose file (it pulls the image published by the pipeline):

```bash
docker compose -f deploy/docker-compose.yaml up -d
```

## API

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/health` | Health check |
| GET | `/items` | List items |
| POST | `/items` | Create an item |
| GET | `/items/{id}` | Get one item |
| PUT | `/items/{id}` | Update an item |
| DELETE | `/items/{id}` | Delete an item |

```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Learn Go","description":"Build a small API"}'
```

Full details and more examples: [`app/README.md`](app/README.md).

## Tests

```bash
cd app
go test ./...        # run all tests
go test -race ./...  # concurrency check
```

The pipeline runs `go test -C ./app ./...` before building, so a failing test
stops the deployment.

## CI/CD setup

### 1. Add repository secrets

Go to **Settings → Secrets and variables → Actions → New repository secret** and add:

| Secret | Value |
| --- | --- |
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub **access token** (Account Settings → Security → New Access Token, permission **Read & Write**) |

Use an access token, not your Docker Hub password.

### 2. Set up the self-hosted runner

The `deploy` job runs on your own machine (label `self-hosted`). GitHub provides
the exact download URL and token in **Settings → Actions → Runners → New
self-hosted runner**. Choose your OS, then run the commands it shows.

**Linux:**

```bash
mkdir actions-runner && cd actions-runner
curl -o actions-runner.tar.gz -L <download-url-from-github>
tar xzf actions-runner.tar.gz
./config.sh --url https://github.com/ArrifUber/Simple-Multi-Stage-and-CICD- --token <RUNNER_TOKEN>
./run.sh                                    # test in foreground
sudo ./svc.sh install && sudo ./svc.sh start # then run as a service
```

**Windows (PowerShell):**

```powershell
mkdir actions-runner; cd actions-runner
Invoke-WebRequest -Uri <download-url-from-github> -OutFile actions-runner.zip
Expand-Archive actions-runner.zip -DestinationPath .
./config.cmd --url https://github.com/ArrifUber/Simple-Multi-Stage-and-CICD- --token <RUNNER_TOKEN>
./run.cmd
```

Notes:

- The runner host needs Docker installed and running; on Linux, add your user to
  the `docker` group (`sudo usermod -aG docker $USER`) so `docker` works without `sudo`.
- The runner token is short-lived — get a fresh one from the GitHub UI if it expires.
- Keep the runner online as a service so queued deployments can run.

### 3. Match the image name

The pipeline pushes to:

```
$DOCKERHUB_USERNAME/simple-golang-multi-stage:latest
```

`deploy/docker-compose.yaml` currently references
`arifbudianto/simple-golang-multi-stage:latest`. If your Docker Hub username is
different, update the `image:` line there.

### 4. Deploy

Push to `main`. When the workflow finishes, the app is rebuilt, pushed, pulled on
the runner, and restarted via Docker Compose at `http://localhost:8080`.

## Troubleshooting

- **Job stuck on "Waiting for a runner"** — the self-hosted runner is offline; start it.
- **`unauthorized` when pushing** — check `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN`; the token needs Read & Write.
- **Port 8080 already in use** — stop the conflicting container or change the port mapping in `deploy/docker-compose.yaml`.
- **Data disappears after restart** — expected: storage is in memory, there is no database.
