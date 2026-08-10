package gometa

import (
	"fmt"
	"sort"
)

// DefinitionError reports invalid statically declared metadata. Builders panic
// with this error because a malformed metadata declaration is a programming
// error and must not reach runtime consumers.
type DefinitionError struct {
	Path    string
	Message string
}

func (e *DefinitionError) Error() string {
	if e.Path == "" {
		return "gometa: " + e.Message
	}
	return "gometa: " + e.Path + ": " + e.Message
}

// TypeOf builds metadata for T. Its arguments may be type annotations,
// FieldMetadata values returned by Field, and MethodMetadata values returned by
// Method. T is intentionally not inspected at runtime; use the gometa analyzer
// to validate the binding to the actual Go type.
func TypeOf[T any](parts ...any) *TypeMetadata {
	metadata := &TypeMetadata{
		Fields:  make(map[string]*FieldMetadata),
		Methods: make(map[string]*MethodMetadata),
	}

	for _, part := range parts {
		switch value := part.(type) {
		case Annotation:
			metadata.Annotations = append(metadata.Annotations, checkedAnnotations("type", TargetType, value)...)
		case FieldMetadata:
			addField(metadata, value)
		case *FieldMetadata:
			if value == nil {
				definitionPanic("type", "nil field metadata")
			}
			addField(metadata, *value)
		case MethodMetadata:
			addMethod(metadata, value)
		case *MethodMetadata:
			if value == nil {
				definitionPanic("type", "nil method metadata")
			}
			addMethod(metadata, *value)
		case nil:
			definitionPanic("type", "nil metadata part")
		default:
			definitionPanic("type", fmt.Sprintf("unsupported metadata part %T", part))
		}
	}
	return metadata
}

// Field creates metadata for a directly declared struct field.
func Field(name string, annotations ...Annotation) FieldMetadata {
	if name == "" {
		definitionPanic("field", "name must not be empty")
	}
	return FieldMetadata{
		Name:        name,
		Annotations: checkedAnnotations("field "+name, TargetField, annotations...),
	}
}

// Method creates method metadata. Its parts may be method annotations,
// ParameterMetadata values returned by Param or NamedParam, and ResultMetadata
// values returned by Result or NamedResult.
func Method(name string, parts ...any) MethodMetadata {
	if name == "" {
		definitionPanic("method", "name must not be empty")
	}

	metadata := MethodMetadata{Name: name}
	parameterIndexes := make(map[int]struct{})
	resultIndexes := make(map[int]struct{})
	for _, part := range parts {
		switch value := part.(type) {
		case Annotation:
			metadata.Annotations = append(metadata.Annotations, checkedAnnotations("method "+name, TargetMethod, value)...)
		case ParameterMetadata:
			if _, exists := parameterIndexes[value.Index]; exists {
				definitionPanic("method "+name, fmt.Sprintf("duplicate parameter index %d", value.Index))
			}
			parameterIndexes[value.Index] = struct{}{}
			metadata.Parameters = append(metadata.Parameters, value)
		case *ParameterMetadata:
			if value == nil {
				definitionPanic("method "+name, "nil parameter metadata")
			}
			if _, exists := parameterIndexes[value.Index]; exists {
				definitionPanic("method "+name, fmt.Sprintf("duplicate parameter index %d", value.Index))
			}
			parameterIndexes[value.Index] = struct{}{}
			metadata.Parameters = append(metadata.Parameters, *value)
		case ResultMetadata:
			if _, exists := resultIndexes[value.Index]; exists {
				definitionPanic("method "+name, fmt.Sprintf("duplicate result index %d", value.Index))
			}
			resultIndexes[value.Index] = struct{}{}
			metadata.Results = append(metadata.Results, value)
		case *ResultMetadata:
			if value == nil {
				definitionPanic("method "+name, "nil result metadata")
			}
			if _, exists := resultIndexes[value.Index]; exists {
				definitionPanic("method "+name, fmt.Sprintf("duplicate result index %d", value.Index))
			}
			resultIndexes[value.Index] = struct{}{}
			metadata.Results = append(metadata.Results, *value)
		case nil:
			definitionPanic("method "+name, "nil metadata part")
		default:
			definitionPanic("method "+name, fmt.Sprintf("unsupported metadata part %T", part))
		}
	}
	return metadata
}

// Param creates metadata for a method parameter at index.
func Param(index int, annotations ...Annotation) ParameterMetadata {
	return NamedParam(index, "", annotations...)
}

// NamedParam creates parameter metadata with optional descriptive name.
func NamedParam(index int, name string, annotations ...Annotation) ParameterMetadata {
	if index < 0 {
		definitionPanic("parameter", fmt.Sprintf("index %d must not be negative", index))
	}
	return ParameterMetadata{
		Index:       index,
		Name:        name,
		Annotations: checkedAnnotations(fmt.Sprintf("parameter %d", index), TargetParameter, annotations...),
	}
}

// Result creates metadata for a method result at index.
func Result(index int, annotations ...Annotation) ResultMetadata {
	return NamedResult(index, "", annotations...)
}

// NamedResult creates result metadata with optional descriptive name.
func NamedResult(index int, name string, annotations ...Annotation) ResultMetadata {
	if index < 0 {
		definitionPanic("result", fmt.Sprintf("index %d must not be negative", index))
	}
	return ResultMetadata{
		Index:       index,
		Name:        name,
		Annotations: checkedAnnotations(fmt.Sprintf("result %d", index), TargetResult, annotations...),
	}
}

func addField(metadata *TypeMetadata, field FieldMetadata) {
	if field.Name == "" {
		definitionPanic("field", "name must not be empty")
	}
	if _, exists := metadata.Fields[field.Name]; exists {
		definitionPanic("type", "duplicate field "+field.Name)
	}
	copy := field
	copy.Annotations = checkedAnnotations("field "+field.Name, TargetField, field.Annotations...)
	metadata.Fields[field.Name] = &copy
}

func addMethod(metadata *TypeMetadata, method MethodMetadata) {
	if method.Name == "" {
		definitionPanic("method", "name must not be empty")
	}
	if _, exists := metadata.Methods[method.Name]; exists {
		definitionPanic("type", "duplicate method "+method.Name)
	}
	copy := method
	copy.Annotations = checkedAnnotations("method "+method.Name, TargetMethod, method.Annotations...)
	copy.Parameters = normalizeParameters(method.Name, method.Parameters)
	copy.Results = normalizeResults(method.Name, method.Results)
	metadata.Methods[method.Name] = &copy
}

func normalizeParameters(methodName string, parameters []ParameterMetadata) []ParameterMetadata {
	normalized := make([]ParameterMetadata, len(parameters))
	indexes := make(map[int]struct{}, len(parameters))
	for position, parameter := range parameters {
		if parameter.Index < 0 {
			definitionPanic("method "+methodName, fmt.Sprintf("parameter index %d must not be negative", parameter.Index))
		}
		if _, exists := indexes[parameter.Index]; exists {
			definitionPanic("method "+methodName, fmt.Sprintf("duplicate parameter index %d", parameter.Index))
		}
		indexes[parameter.Index] = struct{}{}
		normalized[position] = parameter
		normalized[position].Annotations = checkedAnnotations(
			fmt.Sprintf("method %s parameter %d", methodName, parameter.Index),
			TargetParameter,
			parameter.Annotations...,
		)
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Index < normalized[j].Index })
	return normalized
}

func normalizeResults(methodName string, results []ResultMetadata) []ResultMetadata {
	normalized := make([]ResultMetadata, len(results))
	indexes := make(map[int]struct{}, len(results))
	for position, result := range results {
		if result.Index < 0 {
			definitionPanic("method "+methodName, fmt.Sprintf("result index %d must not be negative", result.Index))
		}
		if _, exists := indexes[result.Index]; exists {
			definitionPanic("method "+methodName, fmt.Sprintf("duplicate result index %d", result.Index))
		}
		indexes[result.Index] = struct{}{}
		normalized[position] = result
		normalized[position].Annotations = checkedAnnotations(
			fmt.Sprintf("method %s result %d", methodName, result.Index),
			TargetResult,
			result.Annotations...,
		)
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Index < normalized[j].Index })
	return normalized
}

func checkedAnnotations(path string, target Target, annotations ...Annotation) []Annotation {
	checked := make([]Annotation, len(annotations))
	for index, annotation := range annotations {
		if annotation == nil {
			definitionPanic(path, fmt.Sprintf("annotation %d is nil", index))
		}
		targets := annotation.Targets()
		if !targets.Valid() {
			definitionPanic(path, fmt.Sprintf("annotation %T has invalid targets %s", annotation, targets))
		}
		if !targets.Supports(target) {
			definitionPanic(path, fmt.Sprintf("annotation %T targets %s, not %s", annotation, targets, target))
		}
		checked[index] = annotation
	}
	return checked
}

func definitionPanic(path, message string) {
	panic(&DefinitionError{Path: path, Message: message})
}
