# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Traefik middleware plugin written in Go that removes query parameters from HTTP requests using regex patterns. The plugin allows filtering query parameters while optionally preserving certain paths from modification.

## Key Components

- `traefikremovequeryparametersbyregex.go`: Main plugin implementation with `QueryParameterRemover` middleware
- `traefikremovequeryparametersbyregex_test.go`: Comprehensive test suite
- `.traefik.yml`: Traefik plugin configuration and metadata

## Development Commands

```bash
# Run linting (uses golangci-lint with strict configuration)
make lint
# or: golangci-lint run

# Run tests with coverage
make test
# or: go test -v -cover ./...

# Run yaegi tests (Traefik's Go interpreter)
make yaegi_test
# or: yaegi test -v .

# Run both lint and test (default target)
make

# Create vendor directory
make vendor

# Clean vendor directory
make clean
```

## Architecture

The plugin implements a single middleware type `deleteexcept` that:
1. Compiles regex patterns for allowed query parameters (`AllowedValuesRegex`) and URI exceptions (`ExceptURIRegex`)
2. Processes HTTP requests by filtering query parameters that don't match the allowed pattern
3. Optionally adds `Plugin-Original-Uri` header when `AddOriginalHostnameHeader` is enabled
4. Preserves original URLs for paths matching the exception pattern

Configuration validation ensures either `AllowedValuesRegex` or `ExceptURIRegex` is provided.

## Testing

Tests cover various scenarios including parameter filtering, URI exceptions, header addition, and error conditions. Use `go test -v -cover ./...` for running tests with coverage information.