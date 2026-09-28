# Contributing to gometa

Thank you for your interest in contributing to `gometa`! We welcome contributions from the community, whether they are bug reports, documentation improvements, feature requests, or pull requests.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git

### Cloning the Repository

```bash
git clone https://github.com/yvvlee/gometa.git
cd gometa
```

## Running Tests and Linter

Before submitting a pull request, ensure all tests and linters pass cleanly:

```bash
# Run all tests with race detector
go test -v -race ./...

# Run static checks
go vet ./...

# Run the gometa static analyzer on example packages
go install ./cmd/gometa
go vet -vettool="$(command -v gometa)" ./examples/...
```

## Project Structure

- `metadata.go`: Core data structures (`TypeMetadata`, `FieldMetadata`, `MethodMetadata`, `Annotation`, etc.) and lookup helpers (`FindAnnotation`, `AllAnnotations`).
- `target.go`: Declaration targets (`TargetType`, `TargetField`, `TargetMethod`, etc.) and bitmask logic.
- `builder.go`: Fluent constructors (`Field`, `Method`, `Param`, `Result`) and boot-time reflection validation.
- `registry.go`: Thread-safe global metadata registry (`Register`, `MetadataOf`, `Registrations`).
- `traversal.go`: Deterministic iteration helpers (`RangeFields`, `RangeMethods`, `Walk`, `WalkAnnotations`).
- `analyzer/`: Static analyzer implementation integrating with `golang.org/x/tools/go/analysis`.
- `cmd/gometa/`: Standalone CLI binary for `go vet -vettool`.
- `examples/`: Runnable real-world examples.

## Pull Request Guidelines

1. **Keep PRs focused**: Each PR should address a single bug fix, feature, or documentation update.
2. **Add tests**: Any new feature or bug fix should be accompanied by comprehensive tests.
3. **Preserve documentation**: When updating code, make sure corresponding documentation (`README.md`, `README_ZH.md`, `doc.go`) is updated and kept in sync.
4. **Code style**: Follow standard Go formatting (`gofmt`, `goimports`).

## Reporting Issues

If you find a bug or have a suggestion:
1. Search existing issues to ensure it hasn't already been reported.
2. Open a new issue with a clear description, reproduction steps, and Go version information.
