package gometa

type TypeMetadata struct{}

func Register[T any](parts ...any) {}
func Field(name string, annotations ...any) any  { return nil }
func Method(name string, parts ...any) any       { return nil }
func Param(index int, annotations ...any) any    { return nil }
func NamedParam(index int, name string, annotations ...any) any {
	return nil
}
func Result(index int, annotations ...any) any { return nil }
func NamedResult(index int, name string, annotations ...any) any {
	return nil
}
