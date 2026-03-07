# Quick Start Guide

## Before You Begin

This SDK is ready to use, but you need to update the module path to match your repository.

## Step 1: Update Module Path

Replace `github.com/yourusername/gofoundation` with your actual repository path in:

1. **go.mod** - Line 1
2. All import statements in:
   - `gateway/*.go`
   - `middleware/*.go`
   - `response/*.go`
   - `logger/*.go`
   - `examples/*/*.go`

Quick find and replace:
```bash
find . -type f -name "*.go" -exec sed -i '' 's|github.com/yourusername/gofoundation|github.com/YOUR_USERNAME/gofoundation|g' {} +
sed -i '' 's|github.com/yourusername/gofoundation|github.com/YOUR_USERNAME/gofoundation|g' go.mod
```

## Step 2: Tidy Dependencies

```bash
go mod tidy
```

## Step 3: Run Tests

```bash
go test ./...
```

## Step 4: Try the Basic Example

```bash
cd examples/basic
go run main.go
```

In another terminal:
```bash
curl http://localhost:8080/api/users
```

You should see:
```json
{
  "data": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ],
  "trace_id": "..."
}
```

## Step 5: Start Building!

Create your own service:

```go
package main

import (
    "log"
    "net/http"
    "github.com/YOUR_USERNAME/gofoundation/gateway"
    "github.com/YOUR_USERNAME/gofoundation/response"
)

func main() {
    gw, _ := gateway.New(gateway.DefaultConfig("my-service"))

    http.HandleFunc("/api/hello", gw.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := w.(response.Writer)
        rw.WriteJSON(http.StatusOK, map[string]string{
            "message": "Hello, World!",
        })
    }))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Optional: Jaeger Integration

To see traces in Jaeger:

1. Start Jaeger:
```bash
docker run -d \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

2. Run the full-featured example:
```bash
cd examples/full-featured
go run main.go
```

3. Make some requests:
```bash
curl http://localhost:8080/api/users
curl http://localhost:8080/api/orders
```

4. View traces at http://localhost:16686

## Configuration Options

### Minimal (Dev Mode)
```go
gw, _ := gateway.New(gateway.DefaultConfig("my-service"))
```

### Production with File Logging
```go
config := gateway.DefaultConfig("my-service")
config.Logger = logger.Config{
    Level:      logger.InfoLevel,
    DevMode:    false,
    FilePath:   "/var/log/my-service.log",
    MaxSize:    100,  // MB
    MaxAge:     7,    // days
    MaxBackups: 3,
    Compress:   true,
}
gw, _ := gateway.New(config)
```

### With Custom Middleware
```go
config := gateway.DefaultConfig("my-service")
config.Middlewares = []gateway.Middleware{
    authMiddleware,
    rateLimitMiddleware,
}
gw, _ := gateway.New(config)
```

## Documentation

- **README.md** - Full documentation
- **CONTRIBUTING.md** - How to contribute
- **IMPLEMENTATION.md** - Implementation details
- **examples/** - Working examples

## Support

If you encounter issues:
1. Check the examples directory
2. Read the documentation
3. Open an issue on GitHub

Happy coding! 🚀
