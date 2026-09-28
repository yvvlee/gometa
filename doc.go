// Package gometa provides strongly typed declaration metadata and annotations
// for Go types, fields, methods, parameters, and results using ordinary Go code.
//
// Traditional metadata in Go often relies on untyped struct tags or complex
// code generators. gometa replaces those mechanisms with plain Go structs
// implementing the [Annotation] interface, preserving compile-time type safety,
// IDE autocompletion, and safe rename refactoring.
//
// # Architecture and Safety
//
// gometa uses a two-phase verification model:
//
//  1. Boot-time Validation: [Register] validates fields, methods, parameter
//     indexes, and result indexes via reflection during package initialization.
//     Invalid declarations fail immediately with a [*DefinitionError].
//  2. Static Analysis: The companion cmd/gometa tool integrates with go vet
//     to perform the exact same checks during build and CI pipelines before code runs.
//
// Once registered, request-time consumers read metadata through direct pointer,
// map, and slice lookups with zero reflection overhead.
//
// # Defining Annotations
//
// An annotation is any Go type that implements [Annotation] by defining the
// [Annotation.Targets] method:
//
//	type GET struct {
//		Path string
//	}
//
//	func (GET) Targets() gometa.Target {
//		return gometa.TargetMethod
//	}
//
// An annotation may also support multiple targets using bitwise OR (e.g. TargetType | TargetMethod).
//
// # Registering Metadata
//
// Register metadata for a type inside a package's init function:
//
//	func init() {
//		gometa.Register[UserService](
//			Service{Name: "users"},
//			gometa.Field("BaseURL", Inject{Name: "USER_SERVICE_URL"}),
//			gometa.Method("GetUser",
//				GET{Path: "/users/:id"},
//				gometa.NamedParam(0, "id", Path{Name: "id"}),
//				gometa.NamedResult(0, "user", Body{}),
//				gometa.NamedResult(1, "err"),
//			),
//		)
//	}
//
// # Querying Metadata
//
// Retrieve registered metadata without reflection overhead:
//
//	metadata, ok := gometa.MetadataOf[UserService]()
//	if !ok {
//		// not registered
//	}
//
//	// Find first matching annotation:
//	service, found := gometa.FindAnnotation[Service](metadata.Annotations)
//
//	// Query all matching annotations of a type:
//	permissions := gometa.AllAnnotations[Permission](method.Annotations)
//
// # Traversal and Discovery
//
// Frameworks can iterate through metadata in a deterministic order using
// [TypeMetadata.RangeFields], [TypeMetadata.RangeMethods],
// [MethodMetadata.RangeParameters], and [MethodMetadata.RangeResults], or
// perform full hierarchy traversals via [Walk] and [WalkAnnotations].
//
// The global process registry can be inspected via [Registrations].
package gometa
