# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK library for building Atomist Skills (v2 contract). Provides the core framework for creating Docker-based skills that respond to datalog subscription events, transact entities, and interact with the Atomist platform.

## Project Structure

```
handle.go                 # Core HTTP handler and skill.Start() entry point
types.go                  # Core types (EventIncoming, RequestContext, Status, etc.)
spec.go                   # Skill spec parsing from skill.yaml
subscriptions.go          # Subscription file loading from /datalog
transact.go               # Entity transact functions (ordered and unordered)
http_transact.go          # HTTP-based transact implementation
status.go                 # Status reporting to the platform
log.go                    # Logger implementation (stdout + Cloud Logging)
name.go                   # Event name extraction from incoming events
util/
  decode.go               # EDN decoding utilities
  parameter.go            # Skill parameter access helpers
middleware/
  signing.go              # Request signing middleware
policy/
  policy.go               # Policy evaluation framework
  data/                   # Data sources for policy evaluation
  goals/                  # Goal types
  skills/                 # Skill configuration parsing
  types/                  # GraphQL and policy types
sbom/
  normalization/          # SBOM normalization utilities
environment/
  environment.go          # Runtime environment detection
test/
  simulate.go             # Test simulation utilities
internal/
  types.go, util.go       # Internal helper types
  test_util/              # Test utilities (logger, pointer helpers)
test_data/
  skill.yaml              # Test fixture for skill spec parsing
```

## Common Development Commands

```bash
# Using Taskfile (preferred)
task go:test              # Run all tests
task go:build             # Build binary
task go:fmt               # Format code (goimports, gofmt, go mod tidy)

# Using go directly
go test -v ./... --count=1    # Run all tests
go build -ldflags="-w -s"    # Build optimized binary
```

## Key Patterns

- `skill.Start(handlers)` is the main entry point, starts an HTTP server on PORT (default 8080)
- `Handlers` is a function type mapping subscription/webhook names to `EventHandler` functions
- `EventHandler` signature: `func(ctx context.Context, req RequestContext) Status`
- EDN encoding/decoding used throughout for Atomist platform communication
- `RequestContext` provides access to Event, Logger, Transact, and TransactOrdered
- `Decode[T]` generic function for type-safe EDN decoding of subscription results

## Testing

Tests use `stretchr/testify` for assertions. Run with `go test -v ./... --count=1` or `task go:test`.

## Important Configuration

- Go 1.21.6 minimum
- Uses `olympos.io/encoding/edn` for EDN serialization
- Cloud Logging integration via `cloud.google.com/go/logging`
- Container registry interaction via `google/go-containerregistry`
- SBOM handling via `anchore/syft`
- VEX document support via `openvex/go-vex`
