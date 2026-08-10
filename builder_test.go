package gometa_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yvvlee/gometa"
)

type Deprecated struct{}

func (Deprecated) Targets() gometa.Target { return gometa.TargetType | gometa.TargetMethod }

type Column struct{ Name string }

func (Column) Targets() gometa.Target { return gometa.TargetField }

type GET struct{ Path string }

func (GET) Targets() gometa.Target { return gometa.TargetMethod }

type Path struct{ Name string }

func (Path) Targets() gometa.Target { return gometa.TargetParameter }

type Body struct{}

func (Body) Targets() gometa.Target { return gometa.TargetResult }

type Service struct {
	ID int64
}

func (Service) Get(int64) (string, error) { return "", nil }

var serviceMetadata = gometa.TypeOf[Service](
	Deprecated{},
	gometa.Field("ID", Column{Name: "id"}),
	gometa.Method(
		"Get",
		GET{Path: "/users/:id"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.Result(0, Body{}),
	),
)

func (Service) Metadata() *gometa.TypeMetadata { return serviceMetadata }

var _ gometa.MetadataProvider = Service{}

func TestBuildAndLookup(t *testing.T) {
	metadata := Service{}.Metadata()
	if _, ok := gometa.FindAnnotation[Deprecated](metadata.Annotations); !ok {
		t.Fatal("type annotation not found")
	}

	field, ok := metadata.Field("ID")
	if !ok {
		t.Fatal("field metadata not found")
	}
	column, ok := gometa.FindAnnotation[Column](field.Annotations)
	if !ok || column.Name != "id" {
		t.Fatalf("unexpected field annotation: %#v, %v", column, ok)
	}

	method, ok := metadata.Method("Get")
	if !ok {
		t.Fatal("method metadata not found")
	}
	parameter, ok := method.Parameter(0)
	if !ok || parameter.Name != "id" {
		t.Fatalf("unexpected parameter metadata: %#v, %v", parameter, ok)
	}
	if _, ok := method.Result(0); !ok {
		t.Fatal("result metadata not found")
	}
}

func TestDefinitionFailures(t *testing.T) {
	tests := []struct {
		name  string
		build func()
		want  string
	}{
		{
			name:  "invalid target",
			build: func() { gometa.Field("ID", GET{}) },
			want:  "not field",
		},
		{
			name: "duplicate field",
			build: func() {
				gometa.TypeOf[Service](
					gometa.FieldMetadata{Name: "ID"},
					gometa.FieldMetadata{Name: "ID"},
				)
			},
			want: "duplicate field ID",
		},
		{
			name: "duplicate parameter",
			build: func() {
				gometa.Method("Get", gometa.Param(0), gometa.Param(0))
			},
			want: "duplicate parameter index 0",
		},
		{
			name:  "negative result",
			build: func() { gometa.Result(-1) },
			want:  "must not be negative",
		},
		{
			name: "invalid direct method metadata",
			build: func() {
				gometa.TypeOf[Service](gometa.MethodMetadata{
					Name: "Get",
					Parameters: []gometa.ParameterMetadata{
						{Index: 0},
						{Index: 0},
					},
				})
			},
			want: "duplicate parameter index 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				value := recover()
				if value == nil {
					t.Fatal("expected panic")
				}
				if message := fmt.Sprint(value); !strings.Contains(message, test.want) {
					t.Fatalf("panic %q does not contain %q", message, test.want)
				}
			}()
			test.build()
		})
	}
}
