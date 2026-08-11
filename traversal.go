package gometa

import (
	"fmt"
	"sort"
)

// Declaration identifies one node in a metadata tree.
//
// Type always points to the root metadata. Field is set for field nodes.
// Method is set for method nodes and for their parameter and result children.
// Parameter or Result is set for the corresponding child node. Target
// identifies which node is currently being visited.
type Declaration struct {
	Target    Target
	Type      *TypeMetadata
	Field     *FieldMetadata
	Method    *MethodMetadata
	Parameter *ParameterMetadata
	Result    *ResultMetadata
}

// Annotations returns the annotations attached to the current declaration.
func (d Declaration) Annotations() []Annotation {
	switch d.Target {
	case TargetType:
		if d.Type != nil {
			return d.Type.Annotations
		}
	case TargetField:
		if d.Field != nil {
			return d.Field.Annotations
		}
	case TargetMethod:
		if d.Method != nil {
			return d.Method.Annotations
		}
	case TargetParameter:
		if d.Parameter != nil {
			return d.Parameter.Annotations
		}
	case TargetResult:
		if d.Result != nil {
			return d.Result.Annotations
		}
	}
	return nil
}

// Name returns the field, method, parameter, or result name. Type declarations
// have no runtime name because TypeOf deliberately avoids reflection.
func (d Declaration) Name() string {
	switch d.Target {
	case TargetField:
		if d.Field != nil {
			return d.Field.Name
		}
	case TargetMethod:
		if d.Method != nil {
			return d.Method.Name
		}
	case TargetParameter:
		if d.Parameter != nil {
			return d.Parameter.Name
		}
	case TargetResult:
		if d.Result != nil {
			return d.Result.Name
		}
	}
	return ""
}

// Index returns the zero-based index for a parameter or result declaration.
func (d Declaration) Index() (int, bool) {
	switch d.Target {
	case TargetParameter:
		if d.Parameter != nil {
			return d.Parameter.Index, true
		}
	case TargetResult:
		if d.Result != nil {
			return d.Result.Index, true
		}
	}
	return 0, false
}

// RangeFields visits fields in ascending name order. Returning false stops the
// traversal.
func (m *TypeMetadata) RangeFields(visit func(*FieldMetadata) bool) {
	if m == nil {
		return
	}
	if visit == nil {
		definitionPanic("range fields", "visitor must not be nil")
	}

	fields := make([]*FieldMetadata, 0, len(m.Fields))
	for name, field := range m.Fields {
		if field == nil {
			definitionPanic("range fields", fmt.Sprintf("field %q is nil", name))
		}
		if field.Name != name {
			definitionPanic("range fields", fmt.Sprintf("map key %q does not match field name %q", name, field.Name))
		}
		fields = append(fields, field)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })

	for _, field := range fields {
		if !visit(field) {
			return
		}
	}
}

// RangeMethods visits methods in ascending name order. Returning false stops
// the traversal.
func (m *TypeMetadata) RangeMethods(visit func(*MethodMetadata) bool) {
	if m == nil {
		return
	}
	if visit == nil {
		definitionPanic("range methods", "visitor must not be nil")
	}

	methods := make([]*MethodMetadata, 0, len(m.Methods))
	for name, method := range m.Methods {
		if method == nil {
			definitionPanic("range methods", fmt.Sprintf("method %q is nil", name))
		}
		if method.Name != name {
			definitionPanic("range methods", fmt.Sprintf("map key %q does not match method name %q", name, method.Name))
		}
		methods = append(methods, method)
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Name < methods[j].Name })

	for _, method := range methods {
		if !visit(method) {
			return
		}
	}
}

// RangeParameters visits parameters in ascending index order. Returning false
// stops the traversal.
func (m *MethodMetadata) RangeParameters(visit func(*ParameterMetadata) bool) {
	if m == nil {
		return
	}
	if visit == nil {
		definitionPanic("range parameters", "visitor must not be nil")
	}

	parameters := make([]*ParameterMetadata, len(m.Parameters))
	indexes := make(map[int]struct{}, len(m.Parameters))
	for position := range m.Parameters {
		parameter := &m.Parameters[position]
		if parameter.Index < 0 {
			definitionPanic("range parameters", fmt.Sprintf("index %d must not be negative", parameter.Index))
		}
		if _, exists := indexes[parameter.Index]; exists {
			definitionPanic("range parameters", fmt.Sprintf("duplicate index %d", parameter.Index))
		}
		indexes[parameter.Index] = struct{}{}
		parameters[position] = parameter
	}
	sort.Slice(parameters, func(i, j int) bool { return parameters[i].Index < parameters[j].Index })

	for _, parameter := range parameters {
		if !visit(parameter) {
			return
		}
	}
}

// RangeResults visits results in ascending index order. Returning false stops
// the traversal.
func (m *MethodMetadata) RangeResults(visit func(*ResultMetadata) bool) {
	if m == nil {
		return
	}
	if visit == nil {
		definitionPanic("range results", "visitor must not be nil")
	}

	results := make([]*ResultMetadata, len(m.Results))
	indexes := make(map[int]struct{}, len(m.Results))
	for position := range m.Results {
		result := &m.Results[position]
		if result.Index < 0 {
			definitionPanic("range results", fmt.Sprintf("index %d must not be negative", result.Index))
		}
		if _, exists := indexes[result.Index]; exists {
			definitionPanic("range results", fmt.Sprintf("duplicate index %d", result.Index))
		}
		indexes[result.Index] = struct{}{}
		results[position] = result
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Index < results[j].Index })

	for _, result := range results {
		if !visit(result) {
			return
		}
	}
}

// Walk visits the complete metadata tree in deterministic order. It visits the
// type first, then fields by name, then methods by name. Each method is followed
// by its parameters and results in ascending index order. Returning false stops
// the complete traversal.
func Walk(metadata *TypeMetadata, visit func(Declaration) bool) {
	if metadata == nil {
		definitionPanic("walk", "metadata must not be nil")
	}
	if visit == nil {
		definitionPanic("walk", "visitor must not be nil")
	}

	if !visit(Declaration{Target: TargetType, Type: metadata}) {
		return
	}

	stopped := false
	metadata.RangeFields(func(field *FieldMetadata) bool {
		if !visit(Declaration{Target: TargetField, Type: metadata, Field: field}) {
			stopped = true
			return false
		}
		return true
	})
	if stopped {
		return
	}

	metadata.RangeMethods(func(method *MethodMetadata) bool {
		if !visit(Declaration{Target: TargetMethod, Type: metadata, Method: method}) {
			return false
		}

		continueMethod := true
		method.RangeParameters(func(parameter *ParameterMetadata) bool {
			continueMethod = visit(Declaration{
				Target:    TargetParameter,
				Type:      metadata,
				Method:    method,
				Parameter: parameter,
			})
			return continueMethod
		})
		if !continueMethod {
			return false
		}

		method.RangeResults(func(result *ResultMetadata) bool {
			continueMethod = visit(Declaration{
				Target: TargetResult,
				Type:   metadata,
				Method: method,
				Result: result,
			})
			return continueMethod
		})
		return continueMethod
	})
}

// WalkAnnotations visits every annotation in deterministic declaration order.
// Returning false stops the complete traversal.
func WalkAnnotations(metadata *TypeMetadata, visit func(Declaration, Annotation) bool) {
	if visit == nil {
		definitionPanic("walk annotations", "visitor must not be nil")
	}
	Walk(metadata, func(declaration Declaration) bool {
		for _, annotation := range declaration.Annotations() {
			if !visit(declaration, annotation) {
				return false
			}
		}
		return true
	})
}
