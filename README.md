# gometa

English | [简体中文](README_ZH.md)

`gometa` adds strongly typed metadata to Go types, fields, methods, parameters, and results using ordinary Go code.

It does not parse string tags or generate code. It also avoids reflection at runtime. A companion static analyzer verifies that metadata still matches the declarations it describes.

## Define annotations

An annotation is an ordinary Go type that implements `gometa.Annotation`. Its `Targets` method declares where the annotation may be attached.

```go
type GET struct {
	Path string
}

func (GET) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Path struct {
	Name string
}

func (Path) Targets() gometa.Target {
	return gometa.TargetParameter
}
```

Targets can be combined when an annotation is valid in more than one place:

```go
func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

## Declare metadata

```go
type UserService struct{}

func (*UserService) GetUser(id int64) (User, error) {
	// ...
}

var userServiceMetadata = gometa.TypeOf[UserService](
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.Result(0, Body{}),
	),
)

func (UserService) Metadata() *gometa.TypeMetadata {
	return userServiceMetadata
}
```

Pass type annotations directly to `TypeOf`. Use `Field` for fields and `Method` for methods. Parameters and results are identified by their zero-based signature index.

The builders immediately reject invalid annotation targets, negative indexes, and duplicate declarations.

## Read metadata

```go
provider := any(UserService{}).(gometa.MetadataProvider)
metadata := provider.Metadata()

method, ok := metadata.Method("GetUser")
get, ok := gometa.FindAnnotation[GET](method.Annotations)
```

## Static validation

Install the analyzer:

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

Run it through `go vet`:

```bash
go vet -vettool="$(command -v gometa)" ./...
```

The analyzer reports unknown fields and methods, out-of-range parameters and results, and mismatches between a `Metadata` receiver and its `TypeOf` target.

## Requirements

`gometa` requires Go 1.22 or later.

