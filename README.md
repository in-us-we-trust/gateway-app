## Scripts Reference

All commands go through `scripts.go` — no `make`, no shell, just Go:

| Command                          | What it does                               |
| -------------------------------- | ------------------------------------------ |
| `go run scripts.go run`          | Start the API server                       |
| `go run scripts.go test`         | Run integration tests with verbose output  |
| `go run scripts.go build`        | Compile binary to `bin/server`             |
| `go run scripts.go docker-build` | Build Docker image `gateway-app`           |
| `go run scripts.go swagger`      | Regenerate `docs/` from source annotations |

P
