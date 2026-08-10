# gometa

[English](README.md) | 简体中文

`gometa` 用普通 Go 代码为类型、字段、方法、参数和返回值声明强类型元数据。

它不解析字符串标签，不生成代码，运行时也不使用反射。静态检查器负责校验元数据和 Go 声明是否一致。

## 功能

- 使用普通 Go 结构体定义强类型注解。
- 支持类型、字段、方法、参数和返回值元数据。
- 校验注解可以使用的目标。
- 使用从零开始的下标标识参数和返回值。
- 查询单个或多个同类型注解。
- 运行时不使用反射，也不需要代码生成。
- 静态校验元数据和 Go 声明的绑定关系。

## 安装

```bash
go get github.com/yvvlee/gometa
```

`gometa` 需要 Go 1.22 或更高版本。

## 完整示例

下面的示例包含每一种元数据层级的声明和读取。

### 定义注解

注解是实现 `gometa.Annotation` 的普通 Go 类型。`Targets` 声明该注解可以放在哪里。

```go
type Service struct {
	Name string
}

func (Service) Targets() gometa.Target {
	return gometa.TargetType
}

type Inject struct {
	Name string
}

func (Inject) Targets() gometa.Target {
	return gometa.TargetField
}

type GET struct {
	Path string
}

func (GET) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Permission struct {
	Name string
}

func (Permission) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Path struct {
	Name string
}

func (Path) Targets() gometa.Target {
	return gometa.TargetParameter
}

type Header struct {
	Name string
}

func (Header) Targets() gometa.Target {
	return gometa.TargetParameter
}

type Body struct{}

func (Body) Targets() gometa.Target {
	return gometa.TargetResult
}
```

一个注解可以支持多个目标：

```go
type Deprecated struct {
	Message string
}

func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

可用目标包括 `TargetType`、`TargetField`、`TargetMethod`、`TargetParameter` 和 `TargetResult`。`TargetAll` 是所有目标的组合。

目标值还提供了辅助方法：

```go
targets := gometa.TargetField | gometa.TargetParameter

targets.Supports(gometa.TargetField) // true
targets.Valid()                      // true
targets.String()                     // "field|parameter"
```

### 声明元数据

```go
type User struct {
	ID   int64
	Name string
}

type UserService struct {
	BaseURL string
}

func (*UserService) GetUser(id int64, authorization string) (User, error) {
	// ...
}

var userServiceMetadata = gometa.TypeOf[UserService](
	// 类型注解。
	Service{Name: "users"},

	// 字段元数据和注解。
	gometa.Field(
		"BaseURL",
		Inject{Name: "USER_SERVICE_BASE_URL"},
	),

	// 方法、参数和返回值元数据。
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		Permission{Name: "user.read"},
		Permission{Name: "audit.read"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.NamedParam(1, "authorization", Header{Name: "Authorization"}),
		gometa.NamedResult(0, "user", Body{}),
		gometa.NamedResult(1, "err"),
	),
)

func (UserService) Metadata() *gometa.TypeMetadata {
	return userServiceMetadata
}

// 在编译期检查接口实现。
var _ gometa.MetadataProvider = UserService{}
```

不需要描述性名称时，可以使用 `Param` 和 `Result`：

```go
gometa.Param(0, Path{Name: "id"})
gometa.Result(0, Body{})
```

`NamedParam` 和 `NamedResult` 可以保存用于文档和诊断的名称。参数和返回值仍然以从零开始的下标作为身份。

把构建完成的元数据保存在包变量中，可以避免每次调用 `Metadata` 时重复构建。

### 读取元数据

```go
provider := any(UserService{}).(gometa.MetadataProvider)
metadata := provider.Metadata()

// 类型注解。
service, found := gometa.FindAnnotation[Service](metadata.Annotations)

// 字段元数据和注解。
field, found := metadata.Field("BaseURL")
inject, found := gometa.FindAnnotation[Inject](field.Annotations)

// 方法元数据和第一个匹配的注解。
method, found := metadata.Method("GetUser")
get, found := gometa.FindAnnotation[GET](method.Annotations)

// 获取某种类型的全部重复注解。
permissions := gometa.AllAnnotations[Permission](method.Annotations)

// 入参元数据和注解。
parameter, found := method.Parameter(0)
path, found := gometa.FindAnnotation[Path](parameter.Annotations)

// 出参元数据和注解。
result, found := method.Result(0)
body, found := gometa.FindAnnotation[Body](result.Annotations)
```

`Field`、`Method`、`Parameter` 和 `Result` 返回对应的元数据，以及表示它是否存在的布尔值。`FindAnnotation` 返回指定 Go 类型的第一个注解。`AllAnnotations` 按声明顺序返回全部匹配注解。

## 非法元数据会立即失败

静态元数据声明不合法时，构造器会抛出 `*gometa.DefinitionError`。以下情况都会失败：

- 注解被放在不支持的目标上。
- 注解返回空的目标或未知目标。
- 参数或返回值下标为负数。
- 字段、方法、参数或返回值重复声明。
- 元数据部分为空或类型不受支持。

下面的代码会失败，因为 `GET` 只支持方法：

```go
gometa.Field("BaseURL", GET{Path: "/invalid"})
```

构造器直接失败，因为错误的静态元数据属于程序错误。

## 静态声明校验

运行时构造器可以校验元数据树，但它不会使用反射检查 Go 声明。静态检查器负责完成后一部分校验。

安装检查器：

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

通过 `go vet` 运行：

```bash
go vet -vettool="$(command -v gometa)" ./...
```

检查器会报告：

- `Field` 引用了不存在的字段。
- `Method` 引用了不存在的方法。
- 参数或返回值下标超过方法签名范围。
- 内联声明重复。
- `Metadata` 接收者与内联 `TypeOf` 的目标类型不一致。

## 可运行示例

完整的 HTTP 风格示例位于 [`examples/http`](examples/http)：

```bash
go run ./examples/http
```

