package gometa

// Annotation is ordinary Go data that declares where it may be attached.
// Implementations normally use a small value struct and a value-receiver
// Targets method.
type Annotation interface {
	Targets() Target
}

// TypeMetadata describes a type and its fields and methods.
type TypeMetadata struct {
	Annotations []Annotation
	Fields      map[string]*FieldMetadata
	Methods     map[string]*MethodMetadata
}

// FieldMetadata describes a directly declared struct field.
type FieldMetadata struct {
	Name        string
	Annotations []Annotation
}

// MethodMetadata describes a method, including its input and output values.
type MethodMetadata struct {
	Name        string
	Annotations []Annotation
	Parameters  []ParameterMetadata
	Results     []ResultMetadata
}

// ParameterMetadata describes a method parameter. Index is its zero-based
// position in the method signature. Name is optional descriptive information.
type ParameterMetadata struct {
	Index       int
	Name        string
	Annotations []Annotation
}

// ResultMetadata describes a method result. Index is its zero-based position
// in the method signature. Name is optional descriptive information.
type ResultMetadata struct {
	Index       int
	Name        string
	Annotations []Annotation
}

// Field looks up field metadata by its Go field name.
func (m *TypeMetadata) Field(name string) (*FieldMetadata, bool) {
	if m == nil {
		return nil, false
	}
	field, ok := m.Fields[name]
	return field, ok
}

// Method looks up method metadata by its Go method name.
func (m *TypeMetadata) Method(name string) (*MethodMetadata, bool) {
	if m == nil {
		return nil, false
	}
	method, ok := m.Methods[name]
	return method, ok
}

// Parameter looks up parameter metadata by its signature index.
func (m *MethodMetadata) Parameter(index int) (*ParameterMetadata, bool) {
	if m == nil {
		return nil, false
	}
	for i := range m.Parameters {
		if m.Parameters[i].Index == index {
			return &m.Parameters[i], true
		}
	}
	return nil, false
}

// Result looks up result metadata by its signature index.
func (m *MethodMetadata) Result(index int) (*ResultMetadata, bool) {
	if m == nil {
		return nil, false
	}
	for i := range m.Results {
		if m.Results[i].Index == index {
			return &m.Results[i], true
		}
	}
	return nil, false
}

// FindAnnotation returns the first annotation whose dynamic type is A.
func FindAnnotation[A Annotation](annotations []Annotation) (A, bool) {
	for _, annotation := range annotations {
		if value, ok := annotation.(A); ok {
			return value, true
		}
	}
	var zero A
	return zero, false
}

// AllAnnotations returns all annotations whose dynamic type is A.
func AllAnnotations[A Annotation](annotations []Annotation) []A {
	values := make([]A, 0)
	for _, annotation := range annotations {
		if value, ok := annotation.(A); ok {
			values = append(values, value)
		}
	}
	return values
}
