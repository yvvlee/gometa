# gometa

English | [简体中文](README_ZH.md)

`gometa` adds strongly typed metadata to Go types, fields, methods, parameters, and results using ordinary Go code.

It does not parse string tags or generate code. It also avoids reflection at runtime. A companion static analyzer verifies that metadata still matches the declarations it describes.

## Features

- Ordinary Go structs as strongly typed annotations.
- Metadata for types, fields, methods, parameters, and results.
- Annotation target validation.
- Zero-based parameter and result identity.
- Single and repeated annotation lookup.
- No runtime reflection or code generation.
- Static validation of declaration bindings.

## Install

```bash
go get github.com/yvvlee/gometa
```

`gometa` requires Go 1.22 or later.

## Complete example

The following example declares and reads every metadata level.

### Define annotations

An annotation is an ordinary Go type that implements `gometa.Annotation`. Its `Targets` method declares where the annotation may be attached.

```go
type Service struct {
	Name string
}

func (Service) Targets() gometa.Target {
	return gometa.TargetType
}

type Inject struct {
	Name string
}

func (Inject) Targets() gometa.Target {
	return gometa.TargetField
}

type GET struct {
	Path string
}

func (GET) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Permission struct {
	Name string
}

func (Permission) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Path struct {
	Name string
}

func (Path) Targets() gometa.Target {
	return gometa.TargetParameter
}

type Header struct {
	Name string
}

func (Header) Targets() gometa.Target {
	return gometa.TargetParameter
}

type Body struct{}

func (Body) Targets() gometa.Target {
	return gometa.TargetResult
}
```

An annotation may support more than one target:

```go
type Deprecated struct {
	Message string
}

func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

Available targets are `TargetType`, `TargetField`, `TargetMethod`, `TargetParameter`, and `TargetResult`. `TargetAll` combines all of them.

Target values also provide helper methods:

```go
targets := gometa.TargetField | gometa.TargetParameter

targets.Supports(gometa.TargetField) // true
targets.Valid()                      // true
targets.String()                     // "field|parameter"
```

### Declare metadata

```go
type User struct {
	ID   int64
	Name string
}

type UserService struct {
	BaseURL string
}

func (*UserService) GetUser(id int64, authorization string) (User, error) {
	// ...
}

var userServiceMetadata = gometa.TypeOf[UserService](
	// Type annotation.
	Service{Name: "users"},

	// Field metadata and annotation.
	gometa.Field(
		"BaseURL",
		Inject{Name: "USER_SERVICE_BASE_URL"},
	),

	// Method, parameter, and result metadata.
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		Permission{Name: "user.read"},
		Permission{Name: "audit.read"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.NamedParam(1, "authorization", Header{Name: "Authorization"}),
		gometa.NamedResult(0, "user", Body{}),
		gometa.NamedResult(1, "err"),
	),
)

func (UserService) Metadata() *gometa.TypeMetadata {
	return userServiceMetadata
}

// Compile-time interface check.
var _ gometa.MetadataProvider = UserService{}
```

Use `Param` and `Result` when a descriptive name is not needed:

```go
gometa.Param(0, Path{Name: "id"})
gometa.Result(0, Body{})
```

`NamedParam` and `NamedResult` store an optional name for documentation and diagnostics. The zero-based index remains the identity of a parameter or result.

Keeping the built metadata in a package variable avoids rebuilding it on every `Metadata` call.

### Read metadata

```go
provider := any(UserService{}).(gometa.MetadataProvider)
metadata := provider.Metadata()

// Type annotation.
service, found := gometa.FindAnnotation[Service](metadata.Annotations)

// Field metadata and annotation.
field, found := metadata.Field("BaseURL")
inject, found := gometa.FindAnnotation[Inject](field.Annotations)

// Method metadata and its first matching annotation.
method, found := metadata.Method("GetUser")
get, found := gometa.FindAnnotation[GET](method.Annotations)

// All repeated annotations of one type.
permissions := gometa.AllAnnotations[Permission](method.Annotations)

// Input parameter metadata and annotation.
parameter, found := method.Parameter(0)
path, found := gometa.FindAnnotation[Path](parameter.Annotations)

// Output result metadata and annotation.
result, found := method.Result(0)
body, found := gometa.FindAnnotation[Body](result.Annotations)
```

`Field`, `Method`, `Parameter`, and `Result` return the requested metadata and a boolean indicating whether it exists. `FindAnnotation` returns the first annotation with the requested Go type. `AllAnnotations` returns every matching annotation in declaration order.

## Invalid metadata fails immediately

The builders panic with `*gometa.DefinitionError` when a static metadata declaration is invalid. Invalid cases include:

- An annotation attached to an unsupported target.
- An annotation that returns an empty or unknown target value.
- A negative parameter or result index.
- Duplicate fields, methods, parameters, or results.
- Nil or unsupported metadata parts.

For example, this fails because `GET` only targets methods:

```go
gometa.Field("BaseURL", GET{Path: "/invalid"})
```

Builders fail fast because malformed static metadata is a programming error.

## Static declaration validation

Runtime builders can validate the metadata tree, but they deliberately do not use reflection to inspect Go declarations. The analyzer performs that separate check.

Install it:

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

Run it through `go vet`:

```bash
go vet -vettool="$(command -v gometa)" ./...
```

The analyzer reports:

- Unknown field names passed to `Field`.
- Unknown method names passed to `Method`.
- Parameter and result indexes outside the method signature.
- Duplicate inline declarations.
- A `Metadata` receiver that does not match its inline `TypeOf` target.

## Runnable example

A complete HTTP-style example is available in [`examples/http`](examples/http):

```bash
go run ./examples/http
```

