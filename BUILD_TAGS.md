# Build Tags

This package supports optional build tags:

## Fiber Support

To use Fiber middleware, build with `-tags fiber`:

```bash
# Build
go build -tags fiber

# Run
go run -tags fiber main.go

# Test
go test -tags fiber
```

## FastHTTP Support

To use FastHTTP middleware, build with `-tags fasthttp`:

```bash
# Build
go build -tags fasthttp

# Run
go run -tags fasthttp main.go

# Test
go test -tags fasthttp
```

## Multiple Tag Support

For both Fiber and FastHTTP support:

```bash
go build -tags "fiber fasthttp"
```

## Examples

### Standard HTTP (default)
```bash
go build
```

### With Fiber support
```bash
go build -tags fiber
```

### With FastHTTP support
```bash
go build -tags fasthttp
```

### With all middlewares
```bash
go build -tags "fiber fasthttp"
```

Projects that don't use these tags will have no extra dependencies, keeping the package lightweight.