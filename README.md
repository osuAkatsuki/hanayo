the front-facing portion of https://akatsuki.gg

## Development

```bash
# Install dependencies (first time only)
go mod download
cd web && npm install && cd ..

# Run development server (builds frontend, Go binary, and starts server)
./run-server.sh
```

## Manual Build

```bash
# Build frontend assets
cd web && gulp && cd ..

# Build Go binary
go build -o hanayo .

# Run server
./hanayo
```

## Docker

```bash
make build    # Build Docker image (includes npm/gulp build)
make run      # Run container on port 46221
```
