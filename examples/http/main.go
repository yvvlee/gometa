// Command http demonstrates gometa's complete metadata declaration and lookup
// flow with a small HTTP service.
package main

import (
	"fmt"

	"github.com/yvvlee/gometa"
)

// Service is a type annotation.
type Service struct {
	Name string
}

func (Service) Targets() gometa.Target { return gometa.TargetType }

// Deprecated can annotate both types and methods.
type Deprecated struct {
	Message string
}

func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}

// Inject is a field annotation.
type Inject struct {
	Name string
}

func (Inject) Targets() gometa.Target { return gometa.TargetField }

// GET and Permission are method annotations.
type GET struct {
	Path string
}

func (GET) Targets() gometa.Target { return gometa.TargetMethod }

type Permission struct {
	Name string
}

func (Permission) Targets() gometa.Target { return gometa.TargetMethod }

// Path and Header are parameter annotations.
type Path struct {
	Name string
}

func (Path) Targets() gometa.Target { return gometa.TargetParameter }

type Header struct {
	Name string
}

func (Header) Targets() gometa.Target { return gometa.TargetParameter }

// Body and Failure are result annotations.
type Body struct{}

func (Body) Targets() gometa.Target { return gometa.TargetResult }

type Failure struct{}

func (Failure) Targets() gometa.Target { return gometa.TargetResult }

type User struct {
	ID   int64
	Name string
}

type UserService struct {
	BaseURL string
}

func (*UserService) GetUser(id int64, authorization string) (User, error) {
	return User{ID: id, Name: authorization}, nil
}

var userServiceMetadata = gometa.TypeOf[UserService](
	Service{Name: "users"},
	Deprecated{Message: "use UserServiceV2 for new integrations"},
	gometa.Field(
		"BaseURL",
		Inject{Name: "USER_SERVICE_BASE_URL"},
	),
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		Permission{Name: "user.read"},
		Permission{Name: "audit.read"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.NamedParam(1, "authorization", Header{Name: "Authorization"}),
		gometa.NamedResult(0, "user", Body{}),
		gometa.NamedResult(1, "err", Failure{}),
	),
)

func (UserService) Metadata() *gometa.TypeMetadata {
	return userServiceMetadata
}

var _ gometa.MetadataProvider = UserService{}

func main() {
	provider := any(UserService{}).(gometa.MetadataProvider)
	metadata := provider.Metadata()

	service, _ := gometa.FindAnnotation[Service](metadata.Annotations)
	deprecated, _ := gometa.FindAnnotation[Deprecated](metadata.Annotations)

	field, _ := metadata.Field("BaseURL")
	inject, _ := gometa.FindAnnotation[Inject](field.Annotations)

	method, _ := metadata.Method("GetUser")
	get, _ := gometa.FindAnnotation[GET](method.Annotations)
	permissions := gometa.AllAnnotations[Permission](method.Annotations)

	idParameter, _ := method.Parameter(0)
	path, _ := gometa.FindAnnotation[Path](idParameter.Annotations)

	authorizationParameter, _ := method.Parameter(1)
	header, _ := gometa.FindAnnotation[Header](authorizationParameter.Annotations)

	userResult, _ := method.Result(0)
	_, hasBody := gometa.FindAnnotation[Body](userResult.Annotations)

	errorResult, _ := method.Result(1)
	_, hasFailure := gometa.FindAnnotation[Failure](errorResult.Annotations)

	fmt.Printf("service: %s\n", service.Name)
	fmt.Printf("deprecated: %s\n", deprecated.Message)
	fmt.Printf("field injection: %s\n", inject.Name)
	fmt.Printf("route: GET %s\n", get.Path)
	fmt.Printf("permissions: %s, %s\n", permissions[0].Name, permissions[1].Name)
	fmt.Printf("parameter %d (%s): path %s\n", idParameter.Index, idParameter.Name, path.Name)
	fmt.Printf("parameter %d (%s): header %s\n", authorizationParameter.Index, authorizationParameter.Name, header.Name)
	fmt.Printf("result %d (%s): body=%t\n", userResult.Index, userResult.Name, hasBody)
	fmt.Printf("result %d (%s): failure=%t\n", errorResult.Index, errorResult.Name, hasFailure)
}
