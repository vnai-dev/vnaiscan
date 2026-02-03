# vnaiscan project

## Build Commands

```bash
# Build binary
go build -o vnaiscan ./cmd/vnaiscan

# Run tests
go test ./...

# Build Docker image
docker build -t vnaiscan .

# Run with Docker
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock vnaiscan scan image:tag
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions small and focused
- Add comments for exported functions

## Project Structure

```
vnaiscan/
├── cmd/vnaiscan/       # CLI entrypoint
├── internal/
│   ├── scanner/        # Core scanning logic
│   ├── aggregator/     # Result aggregation
│   ├── reporter/       # Report generation (JSON, HTML, SARIF)
│   ├── cache/          # Image/DB caching
│   └── config/         # Configuration handling
├── pkg/models/         # Shared data models
├── scripts/            # Install scripts
└── docs/               # Documentation
```
