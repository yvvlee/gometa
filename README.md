# gometa

[![Go Version](https://img.shields.io/github/go-mod/go-version/yvvlee/gometa)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/yvvlee/gometa.svg)](https://pkg.go.dev/github.com/yvvlee/gometa)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yvvlee/gometa)](https://goreportcard.com/report/github.com/yvvlee/gometa)

English | [简体中文](README_ZH.md)

`gometa` provides strongly typed metadata and annotations for Go types, fields, methods, parameters, and results using **ordinary Go code**.

By defining annotations as plain Go structs, `gometa` preserves full compile-time type safety, IDE autocompletion, and refactoring support. It enforces safety through **boot-time reflection validation** and **compile-time/CI static analysis**, while guaranteeing **zero reflection overhead** during request handling.

---

## Table of Contents

- [Why gometa?](#why-gometa)
- [Key Features](#key-features)
- [Feature Comparison](#feature-comparison)
- [Installation](#installation)
- [Core Concepts & Architecture](#core-concepts--architecture)
- [End-to-End Walkthrough](#end-to-end-walkthrough)
  - [1. Define Strong Annotations](#1-define-strong-annotations)
  - [2. Declare and Register Metadata](#2-declare-and-register-metadata)
  - [3. Read and Query Metadata](#3-read-and-query-metadata)
  - [4. Deterministic Traversal](#4-deterministic-traversal)
  - [5. Global Discovery & Auto-Wiring](#5-global-discovery--auto-wiring)
- [Static Analyzer & CI Integration](#static-analyzer--ci-integration)
  - [Installation & Usage](#installation--usage)
  - [GitHub Actions Integration](#github-actions-integration)
  - [Rules Checked by Analyzer](#rules-checked-by-analyzer)
- [Error Handling & Fail-Fast Design](#error-handling--fail-fast-design)
- [API Cheatsheet](#api-cheatsheet)
- [Best Practices & Performance](#best-practices--performance)
- [Common Use Cases](#common-use-cases)
- [Contributing](#contributing)
- [License](#license)

---

## Why gometa?

Attaching metadata to code declarations in Go has historically presented difficult trade-offs:

1. **Limitations of Struct Tags**:
   - Untyped string concatenation with no compiler validation for types or typos.
   - Limited strictly to struct fields; cannot annotate types, methods, parameters, or return values.
   - No IDE autocompletion, navigation, or safe rename refactoring.
   - Repeated reflection parsing overhead on request execution paths.
2. **Friction of Code Generators & Magic Comments**:
   - Dependent on fragile comment directives (such as `// +k8s:...` or `// @Router`).
   - Introduces extra tooling, code generation steps, and CI build friction.
   - Generated code can easily drift out of sync with manual changes.

**How `gometa` Solves This**:
- **100% Ordinary Go Code**: Annotations are plain Go structs. Benefit from native Go type checking, refactoring, and linting.
- **Full Declaration Coverage**: Annotate types, fields, methods, parameters (by index/name), and results (by index/name).
- **Dual-Phase Safety**:
  - *Boot-time Validation*: `Register[T]` validates bindings via reflection during application initialization. Typos or out-of-bounds indexes trigger immediate, clear panics.
  - *Compile-Time / CI Analysis*: The companion `cmd/gometa` analyzer checks bindings during `go vet` before code runs.
- **Zero Request-Time Reflection**: Reflection is confined to boot-time registration. Runtime queries use direct pointer, slice, and map lookups.

---

## Key Features

- 🛡️ **Strongly Typed Annotations**: Ordinary Go structs implement annotations with compile-time type safety.
- 🎯 **Comprehensive Targets**: Decorate types (`TargetType`), fields (`TargetField`), methods (`TargetMethod`), parameters (`TargetParameter`), and results (`TargetResult`).
- 🔍 **Dual-Layer Validation**:
  - **Boot-time**: `Register[T]` verifies fields, methods, parameter, and result boundaries using reflection with fail-fast semantics.
  - **Build/CI time**: Official `cmd/gometa` tool integrates with `go vet` to catch invalid declarations ahead of runtime.
- ⚡ **Zero Reflection in Hot Paths**: Consumers read plain Go structs and generic helper functions with no reflection penalty.
- 🔁 **Deterministic Iteration**: Ordered iteration via `RangeFields`, `RangeMethods`, `Walk`, and `WalkAnnotations` using stable alphabetical or index ordering.
- 🌐 **Global Discovery**: Introspect all registered components across packages via `Registrations()` for automatic routing or DI container setup.
- 📦 **Minimal Dependencies**: Core library depends only on the Go standard library (Go 1.22+). Static analyzer uses `golang.org/x/tools`.

---

## Feature Comparison

| Feature | gometa | Struct Tags | Code Generation (`go:generate`) | Runtime Reflection |
| :--- | :---: | :---: | :---: | :---: |
| **Type Safety** | ✅ Compile-time types | ❌ Untyped string | ⚠️ Generator dependent | ❌ Runtime assertions |
| **Annotate Types / Methods / Params** | ✅ Full support | ❌ Struct fields only | ⚠️ Comment syntax only | ⚠️ Limited capability |
| **IDE Autocomplete & Refactoring** | ✅ Native Go support | ❌ No autocomplete / breaks | ⚠️ Requires custom plugins | ❌ String-based |
| **Runtime Performance** | ✅ **Zero reflection** | ❌ Repeated string parsing | ✅ Native code | ❌ High reflection overhead |
| **Build Pipeline Complexity** | ✅ **No code generation** | ✅ Native | ❌ Extra build step required | ✅ Native |
| **Early Error Detection** | ✅ Static check + Boot | ❌ Silent failure at runtime | ⚠️ Build-time failures | ❌ Runtime only |

---

## Installation

### Core Library

```bash
go get github.com/yvvlee/gometa
```

> **Requirement**: Go 1.22 or higher (utilizes generics and `reflect.TypeFor`).

### Static Analyzer (Recommended)

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

---

## Core Concepts & Architecture

`gometa` operates across three distinct phases: **Declaration & Registration**, **Dual Validation**, and **Zero-Overhead Consumption**.

```
┌────────────────────────────────────────────────────────┐
│ 1. Declare Annotations (Annotation) & Metadata         │
│    gometa.Register[T](parts...)                        │
└──────────────────────────┬─────────────────────────────┘
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
┌─────────────────────────┐ ┌───────────────────────────┐
│ 2a. Static Analysis     │ │ 2b. Boot-Time Validation  │
│ (go vet with gometa)    │ │ (Fail-fast DefinitionErr) │
└─────────────────────────┘ └─────────────┬─────────────┘
                                          ▼
┌────────────────────────────────────────────────────────┐
│ 3. Runtime Consumption (Zero Reflection Overhead)      │
│ - MetadataOf[T]() / Registrations()                    │
│ - FindAnnotation[A]() / AllAnnotations[A]()            │
│ - RangeFields / RangeMethods / Walk / WalkAnnotations  │
└────────────────────────────────────────────────────────┘
```

### Targets

Every annotation implements the `gometa.Annotation` interface, returning a bitmask of valid targets:

```go
type Annotation interface {
    Targets() Target
}
```

Standard target bitmasks:
- `TargetType`: Attach to types.
- `TargetField`: Attach to struct fields.
- `TargetMethod`: Attach to methods.
- `TargetParameter`: Attach to method parameters.
- `TargetResult`: Attach to method return results.
- `TargetAll`: Combination of all targets.

---

## End-to-End Walkthrough

Here is a full example demonstrating annotation definition, metadata registration, and query operations. A complete runnable project is available in [examples/http](examples/http).

### 1. Define Strong Annotations

Annotations are normal Go types implementing `Targets() gometa.Target`:

```go
package main

import "github.com/yvvlee/gometa"

// Service targets types.
type Service struct {
	Name string
}
func (Service) Targets() gometa.Target { return gometa.TargetType }

// Inject targets struct fields.
type Inject struct {
	Name string
}
func (Inject) Targets() gometa.Target { return gometa.TargetField }

// GET targets methods.
type GET struct {
	Path string
}
func (GET) Targets() gometa.Target { return gometa.TargetMethod }

// Permission targets methods (supports repeated annotations).
type Permission struct {
	Name string
}
func (Permission) Targets() gometa.Target { return gometa.TargetMethod }

// Path and Header target method parameters.
type Path struct {
	Name string
}
func (Path) Targets() gometa.Target { return gometa.TargetParameter }

type Header struct {
	Name string
}
func (Header) Targets() gometa.Target { return gometa.TargetParameter }

// Body targets method return results.
type Body struct{}
func (Body) Targets() gometa.Target { return gometa.TargetResult }

// Deprecated targets both types and methods via bitwise OR.
type Deprecated struct {
	Message string
}
func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

Target helpers:

```go
targets := gometa.TargetField | gometa.TargetParameter

targets.Supports(gometa.TargetField) // true
targets.Valid()                      // true
targets.String()                     // "field|parameter"
```

---

### 2. Declare and Register Metadata

Register metadata with `gometa.Register[T]`. Invalid field names, missing methods, or out-of-bounds parameter indexes will fail immediately with a panic during application boot:

```go
type User struct {
	ID   int64
	Name string
}

type UserService struct {
	BaseURL string
}

func (*UserService) GetUser(id int64, authorization string) (User, error) {
	return User{ID: id, Name: authorization}, nil
}

// Register inside an init() function during package initialization.
func init() {
	gometa.Register[UserService](
		// 1. Type annotations
		Service{Name: "users"},
		Deprecated{Message: "Use UserServiceV2 for new integrations"},

		// 2. Field metadata and annotations
		gometa.Field(
			"BaseURL",
			Inject{Name: "USER_SERVICE_BASE_URL"},
		),

		// 3. Method, parameter, and result metadata
		gometa.Method(
			"GetUser",
			GET{Path: "/users/:id"},
			Permission{Name: "user.read"},
			Permission{Name: "audit.read"}, // Repeated annotations

			// Zero-based parameter indexes with optional descriptive names
			gometa.NamedParam(0, "id", Path{Name: "id"}),
			gometa.NamedParam(1, "authorization", Header{Name: "Authorization"}),

			// Zero-based result indexes
			gometa.NamedResult(0, "user", Body{}),
			gometa.NamedResult(1, "err"),
		),
	)
}
```

> **Note**: For index-only parameters and results without descriptive names, use `Param` and `Result`:
> ```go
> gometa.Param(0, Path{Name: "id"})
> gometa.Result(0, Body{})
> ```

---

### 3. Read and Query Metadata

Look up metadata using strongly typed generic functions. **Zero reflection is involved**:

```go
// 1. Retrieve registered metadata for a type
metadata, found := gometa.MetadataOf[UserService]()
if !found {
	panic("UserService metadata is not registered")
}

// 2. Query type annotation
service, found := gometa.FindAnnotation[Service](metadata.Annotations)
if found {
	println("Service Name:", service.Name)
}

// 3. Query field metadata and annotations
if field, found := metadata.Field("BaseURL"); found {
	if inject, found := gometa.FindAnnotation[Inject](field.Annotations); found {
		println("Inject Env:", inject.Name)
	}
}

// 4. Query method metadata and annotations
if method, found := metadata.Method("GetUser"); found {
	if get, found := gometa.FindAnnotation[GET](method.Annotations); found {
		println("Route:", get.Path)
	}

	// Retrieve all repeated annotations of a given type
	permissions := gometa.AllAnnotations[Permission](method.Annotations)
	for _, p := range permissions {
		println("Permission:", p.Name)
	}

	// 5. Query parameter annotations by index
	if param, found := method.Parameter(0); found {
		if path, found := gometa.FindAnnotation[Path](param.Annotations); found {
			println("Param 0 path param:", path.Name)
		}
	}

	// 6. Query result annotations by index
	if result, found := method.Result(0); found {
		if _, hasBody := gometa.FindAnnotation[Body](result.Annotations); hasBody {
			println("Result 0 has body annotation")
		}
	}
}
```

---

### 4. Deterministic Traversal

Frameworks typically discover declared structures dynamically. `Range` and `Walk` functions traverse metadata in deterministic order:

#### Iteration via `Range...`

- Fields and methods are traversed in ascending alphabetical order by name.
- Parameters and results are traversed in ascending numerical order by index.
- Returning `false` from a visitor aborts iteration immediately.

```go
metadata.RangeFields(func(field *gometa.FieldMetadata) bool {
	println("Field:", field.Name)
	return true
})

metadata.RangeMethods(func(method *gometa.MethodMetadata) bool {
	println("Method:", method.Name)

	method.RangeParameters(func(param *gometa.ParameterMetadata) bool {
		println("  Parameter:", param.Index, param.Name)
		return true
	})

	method.RangeResults(func(result *gometa.ResultMetadata) bool {
		println("  Result:", result.Index, result.Name)
		return true
	})
	return true
})
```

#### Tree Traversal via `Walk` and `WalkAnnotations`

`Walk` visits the whole declaration hierarchy in a deterministic depth-first order:

```go
gometa.Walk(metadata, func(decl gometa.Declaration) bool {
	switch decl.Target {
	case gometa.TargetType:
		println("Type node, annotations:", len(decl.Annotations()))
	case gometa.TargetField:
		println("Field:", decl.Field.Name)
	case gometa.TargetMethod:
		println("Method:", decl.Method.Name)
	case gometa.TargetParameter:
		println("Parameter:", decl.Method.Name, decl.Parameter.Index)
	case gometa.TargetResult:
		println("Result:", decl.Method.Name, decl.Result.Index)
	}
	return true
})
```

To visit only annotations across the entire tree, use `WalkAnnotations`:

```go
gometa.WalkAnnotations(metadata, func(decl gometa.Declaration, ann gometa.Annotation) bool {
	switch a := ann.(type) {
	case GET:
		println("Register route:", decl.Method.Name, a.Path)
	case Path:
		println("Bind path param:", decl.Method.Name, decl.Parameter.Index, a.Name)
	}
	return true
})
```

---

### 5. Global Discovery & Auto-Wiring

Frameworks can discover all registered services across the application for automated route registration or DI container wiring:

```go
registrations := gometa.Registrations()

for _, reg := range registrations {
	println("Registered type:", reg.Type.String())

	// Check if this type has a Service annotation
	if service, found := gometa.FindAnnotation[Service](reg.Metadata.Annotations); found {
		println("-> Discovered service:", service.Name)
	}
}
```

`Registrations()` returns all registered entries sorted alphabetically by type name, and is thread-safe.

---

## Static Analyzer & CI Integration

The companion static analyzer detects binding errors ahead of time during `go vet`.

### Installation & Usage

```bash
# Install the standalone checker
go install github.com/yvvlee/gometa/cmd/gometa@latest

# Run static checks on your repository
go vet -vettool="$(command -v gometa)" ./...
```

### GitHub Actions Integration

Add the analyzer to your CI pipeline (e.g. `.github/workflows/ci.yml`):

```yaml
name: CI

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Install gometa analyzer
        run: go install github.com/yvvlee/gometa/cmd/gometa@latest
      - name: Run gometa vet
        run: go vet -vettool=$(which gometa) ./...
```

### Rules Checked by Analyzer

1. ❌ **Non-Existent Fields**: `Field("Missing")` references a struct field that does not exist.
2. ❌ **Non-Existent Methods**: `Method("Missing")` references a method not found on the type or its pointer receiver.
3. ❌ **Parameter Index Out of Range**: `Param(2)` exceeds the number of method input parameters.
4. ❌ **Result Index Out of Range**: `Result(1)` exceeds the number of method return values.
5. ❌ **Duplicate Parameter/Result Index**: Defining multiple parameter or result blocks with identical indexes in one method.
6. ❌ **Duplicate Field/Method Declarations**: Specifying duplicate field or method names in a single `Register` call.
7. ❌ **Invalid Target Type**: Passing an unnamed type or primitive (e.g., `Register[any]`).

---

## Error Handling & Fail-Fast Design

Invalid static metadata is treated as a programming bug rather than a recoverable business error. `gometa` follows a strict **fail-fast** principle:

- Structural errors panic during initialization with a `*gometa.DefinitionError`.
- Duplicate type registrations immediately panic to prevent ambiguous state.
- Example panic messages:
  ```text
  panic: gometa: field BaseURL: GET annotation targets method, not field
  panic: gometa: method GetUser: parameter index 5 out of range for UserService.GetUser (2 parameters)
  panic: gometa: type: metadata for UserService is already registered
  ```

---

## API Cheatsheet

| Category | Identifier | Description |
| :--- | :--- | :--- |
| **Targets** | `Target` | Target bitmask type (`uint32`) |
| | `TargetType`, `TargetField`, `TargetMethod` | Base declaration targets |
| | `TargetParameter`, `TargetResult`, `TargetAll` | Parameters, results, and combined mask |
| | `(t Target) Supports(Target) bool` | Tests if target bitmask contains another target |
| | `(t Target) Valid() bool` | Checks if target mask is non-empty and valid |
| | `(t Target) String() string` | Returns formatted string (e.g. `"field\|parameter"`) |
| **Interface** | `Annotation` | Core interface required for annotations: `Targets() Target` |
| **Registry** | `Register[T](parts ...any)` | Validates via reflection and registers metadata in global registry (call in `init()`) |
| | `MetadataOf[T]() (*TypeMetadata, bool)` | Retrieves registered metadata for type `T` |
| | `Registrations() []Registration` | Returns all registrations sorted by type name |
| **Builders** | `Field(name, ...Annotation) FieldMetadata` | Constructs field metadata |
| | `Method(name, ...any) MethodMetadata` | Constructs method metadata (with annotations, params, results) |
| | `Param(index, ...Annotation) ParameterMetadata` | Constructs unnamed parameter metadata |
| | `NamedParam(index, name, ...Annotation)` | Constructs named parameter metadata |
| | `Result(index, ...Annotation) ResultMetadata` | Constructs unnamed result metadata |
| | `NamedResult(index, name, ...Annotation)` | Constructs named result metadata |
| **Queries** | `FindAnnotation[A]([]Annotation) (A, bool)` | Generic lookup for the first matching annotation |
| | `AllAnnotations[A]([]Annotation) []A` | Generic lookup for all matching annotations |
| **Traversal** | `(m *TypeMetadata) RangeFields(func)` | Iterates fields in ascending name order |
| | `(m *TypeMetadata) RangeMethods(func)` | Iterates methods in ascending name order |
| | `(m *MethodMetadata) RangeParameters(func)` | Iterates parameters in ascending index order |
| | `(m *MethodMetadata) RangeResults(func)` | Iterates results in ascending index order |
| | `Walk(metadata, func(Declaration) bool)` | Depth-first traversal over entire metadata hierarchy |
| | `WalkAnnotations(metadata, func)` | Traverses all annotations in declaration order |

---

## Best Practices & Performance

1. **Registration Pattern & Decoupling**:
   - Strongly recommend registering metadata inside `func init() { gometa.Register[T](...) }`, which aligns with standard Go driver/component registration patterns.
   - `Register[T]` is purely a registration function and returns no value. Direct variable binding at the declaration site is discouraged to avoid tight coupling.
   - All consumers (whether within the declaring package or external frameworks) retrieve metadata through the standard entrypoints [`gometa.MetadataOf[T]()`](file:///Users/mshadow/go/src/github.com/yvvlee/gometa/registry.go#L32) or [`gometa.Registrations()`](file:///Users/mshadow/go/src/github.com/yvvlee/gometa/registry.go#L37).
2. **Pointer and Value Receivers**:
   `gometa.Register[T]` automatically checks the method set of both the value type and its pointer (`*T`). Methods with pointer receivers are fully supported whether `T` or `*T` is registered.
3. **Concurrency and Runtime Overhead**:
   - The global registry is guarded by a `sync.RWMutex`.
   - Once initialized, metadata structures (`TypeMetadata`, `MethodMetadata`, etc.) are immutable. Runtime lookups are lockless, non-allocating, and reflection-free.

---

## Common Use Cases

- 🚀 **Declarative Web & RPC Routing**: Automatically configure route tables, middleware bindings, and OpenAPI/Swagger documentation using `@GET`, `@Path`, and `@Header` annotations.
- 🔑 **Declarative Authorization & Auditing**: Attach `@Permission("...")` or `@Audited` to service methods for centralized middleware enforcement.
- 💉 **Dependency Injection**: Decorate struct fields with `@Inject` or `@Value("${ENV}")` for container resolution.
- 📊 **Observability & Tracing**: Instrument specific methods with trace attributes or metrics tagging metadata.

---

## Contributing

Contributions, bug reports, and feature proposals are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on guidelines and our pull request process.

Local development and testing:

```bash
# Run unit tests
go test -v -race ./...

# Run static checks
go vet ./...
```

---

## License

This project is licensed under the [MIT License](LICENSE).
