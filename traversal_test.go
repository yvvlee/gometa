package gometa_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/yvvlee/gometa"
)

type TraversalService struct {
	Alpha string
	Zeta  string
}

func (TraversalService) AlphaMethod(first, second int64) (string, error) {
	return "", nil
}

func (TraversalService) ZetaMethod() {}

func init() {
	gometa.Register[TraversalService](
		Deprecated{},
		gometa.Field("Zeta", Column{Name: "zeta"}),
		gometa.Field("Alpha", Column{Name: "alpha"}),
		gometa.Method("ZetaMethod", GET{}),
		gometa.Method(
			"AlphaMethod",
			GET{},
			gometa.NamedParam(1, "second", Path{Name: "second"}),
			gometa.NamedParam(0, "first", Path{Name: "first"}),
			gometa.NamedResult(1, "err"),
			gometa.NamedResult(0, "value", Body{}),
		),
	)
}

func traversalMetadata() *gometa.TypeMetadata {
	meta, ok := gometa.MetadataOf[TraversalService]()
	if !ok {
		panic("TraversalService metadata not registered")
	}
	return meta
}

func TestRangeMethodsUseStableOrder(t *testing.T) {
	metadata := traversalMetadata()

	var fields []string
	metadata.RangeFields(func(field *gometa.FieldMetadata) bool {
		fields = append(fields, field.Name)
		return true
	})
	if want := []string{"Alpha", "Zeta"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %v, want %v", fields, want)
	}

	var methods []string
	metadata.RangeMethods(func(method *gometa.MethodMetadata) bool {
		methods = append(methods, method.Name)
		return true
	})
	if want := []string{"AlphaMethod", "ZetaMethod"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("methods = %v, want %v", methods, want)
	}

	method, _ := metadata.Method("AlphaMethod")
	var parameters []int
	method.RangeParameters(func(parameter *gometa.ParameterMetadata) bool {
		parameters = append(parameters, parameter.Index)
		return true
	})
	if want := []int{0, 1}; !reflect.DeepEqual(parameters, want) {
		t.Fatalf("parameters = %v, want %v", parameters, want)
	}

	var results []int
	method.RangeResults(func(result *gometa.ResultMetadata) bool {
		results = append(results, result.Index)
		return true
	})
	if want := []int{0, 1}; !reflect.DeepEqual(results, want) {
		t.Fatalf("results = %v, want %v", results, want)
	}
}

func TestWalkUsesStableTreeOrder(t *testing.T) {
	metadata := traversalMetadata()
	var visited []string

	gometa.Walk(metadata, func(declaration gometa.Declaration) bool {
		switch declaration.Target {
		case gometa.TargetType:
			visited = append(visited, "type")
		case gometa.TargetField:
			visited = append(visited, "field:"+declaration.Name())
		case gometa.TargetMethod:
			visited = append(visited, "method:"+declaration.Name())
		case gometa.TargetParameter:
			index, _ := declaration.Index()
			visited = append(visited, "parameter:"+declaration.Method.Name+":"+strconv.Itoa(index))
		case gometa.TargetResult:
			index, _ := declaration.Index()
			visited = append(visited, "result:"+declaration.Method.Name+":"+strconv.Itoa(index))
		}
		return true
	})

	want := []string{
		"type",
		"field:Alpha",
		"field:Zeta",
		"method:AlphaMethod",
		"parameter:AlphaMethod:0",
		"parameter:AlphaMethod:1",
		"result:AlphaMethod:0",
		"result:AlphaMethod:1",
		"method:ZetaMethod",
	}
	if !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want %v", visited, want)
	}
}

func TestWalkCanStopEarly(t *testing.T) {
	metadata := traversalMetadata()
	visits := 0
	gometa.Walk(metadata, func(gometa.Declaration) bool {
		visits++
		return visits < 3
	})
	if visits != 3 {
		t.Fatalf("visits = %d, want 3", visits)
	}
}

func TestWalkAnnotations(t *testing.T) {
	metadata := traversalMetadata()
	var targets []gometa.Target
	gometa.WalkAnnotations(metadata, func(declaration gometa.Declaration, _ gometa.Annotation) bool {
		targets = append(targets, declaration.Target)
		return true
	})

	want := []gometa.Target{
		gometa.TargetType,
		gometa.TargetField,
		gometa.TargetField,
		gometa.TargetMethod,
		gometa.TargetParameter,
		gometa.TargetParameter,
		gometa.TargetResult,
		gometa.TargetMethod,
	}
	if !reflect.DeepEqual(targets, want) {
		t.Fatalf("targets = %v, want %v", targets, want)
	}
}
