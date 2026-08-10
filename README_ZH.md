# gometa

[English](README.md) | 简体中文

`gometa` 用普通 Go 代码为类型、字段、方法、参数和返回值声明强类型元数据。

它不解析字符串标签，不生成代码，运行时也不使用反射。静态检查器负责校验元数据和 Go 声明是否一致。

## 定义注解

注解是实现 `gometa.Annotation` 的普通 Go 类型。`Targets` 声明该注解可以放在哪里。

```go
type GET struct {
	Path string
}

func (GET) Targets() gometa.Target {
	return gometa.TargetMethod
}

type Path struct {
	Name string
}

func (Path) Targets() gometa.Target {
	return gometa.TargetParameter
}
```

一个注解可以组合多个合法目标：

```go
func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

## 声明元数据

```go
type UserService struct{}

func (*UserService) GetUser(id int64) (User, error) {
	// ...
}

var userServiceMetadata = gometa.TypeOf[UserService](
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.Result(0, Body{}),
	),
)

func (UserService) Metadata() *gometa.TypeMetadata {
	return userServiceMetadata
}
```

类型级注解可以直接传给 `TypeOf`。字段使用 `Field`。方法使用 `Method`。参数和返回值以从零开始的下标作为身份。

构造器会立即拒绝非法注解目标、负数下标和重复声明。

## 读取元数据

```go
provider := any(UserService{}).(gometa.MetadataProvider)
metadata := provider.Metadata()

method, ok := metadata.Method("GetUser")
get, ok := gometa.FindAnnotation[GET](method.Annotations)
```

## 静态检查

安装检查器：

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

通过 `go vet` 运行：

```bash
go vet -vettool="$(command -v gometa)" ./...
```

检查器会报告不存在的字段和方法、越界的参数和返回值，以及 `Metadata` 接收者与 `TypeOf` 类型不一致的问题。

## 环境要求

`gometa` 需要 Go 1.22 或更高版本。

