package gometa

import "strings"

// Target identifies the declarations to which an annotation may be attached.
// Multiple targets can be combined with the bitwise OR operator.
type Target uint32

const (
	TargetType Target = 1 << iota
	TargetField
	TargetMethod
	TargetParameter
	TargetResult
)

// TargetAll contains every target supported by gometa.
const TargetAll = TargetType | TargetField | TargetMethod | TargetParameter | TargetResult

// Supports reports whether t contains target.
func (t Target) Supports(target Target) bool {
	return t&target != 0
}

// Valid reports whether t is a non-empty combination of known targets.
func (t Target) Valid() bool {
	return t != 0 && t&^TargetAll == 0
}

func (t Target) String() string {
	if t == 0 {
		return "none"
	}

	names := make([]string, 0, 5)
	for _, item := range []struct {
		target Target
		name   string
	}{
		{TargetType, "type"},
		{TargetField, "field"},
		{TargetMethod, "method"},
		{TargetParameter, "parameter"},
		{TargetResult, "result"},
	} {
		if t.Supports(item.target) {
			names = append(names, item.name)
		}
	}

	unknown := t &^ TargetAll
	if unknown != 0 {
		names = append(names, "unknown")
	}
	return strings.Join(names, "|")
}
