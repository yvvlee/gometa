# gometa

[![Go Version](https://img.shields.io/github/go-mod/go-version/yvvlee/gometa)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/yvvlee/gometa.svg)](https://pkg.go.dev/github.com/yvvlee/gometa)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yvvlee/gometa)](https://goreportcard.com/report/github.com/yvvlee/gometa)

[English](README.md) | 简体中文

`gometa` 是一个用**普通 Go 代码**为类型、字段、方法、参数和返回值声明强类型元数据与注解的开源库。

通过将注解定义为普通的 Go 结构体，`gometa` 在完全保留类型安全、IDE 自动补全和重命名重构能力的同时，提供了**启动期反射校验**与 **构建/CI 静态分析**双重安全防线，且在**请求运行期实现零反射开销**。

---

## 目录

- [为什么需要 gometa？](#为什么需要-gometa)
- [核心特性](#核心特性)
- [方案对比](#方案对比)
- [安装指南](#安装指南)
- [核心概念与架构](#核心概念与架构)
- [完整使用示例](#完整使用示例)
  - [1. 定义强类型注解](#1-定义强类型注解)
  - [2. 声明并注册元数据](#2-声明并注册元数据)
  - [3. 读取与查询元数据](#3-读取与查询元数据)
  - [4. 确定性迭代遍历](#4-确定性迭代遍历)
  - [5. 全局类型发现与自动装配](#5-全局类型发现与自动装配)
- [静态检查工具 (Static Analyzer)](#静态检查工具-static-analyzer)
  - [安装与运行](#安装与运行)
  - [CI 集成示例 (GitHub Actions)](#ci-集成示例-github-actions)
  - [静态检查器覆盖的规则](#静态检查器覆盖的规则)
- [错误处理与 Fail-Fast 设计](#错误处理与-fail-fast-设计)
- [API 速查表](#api-速查表)
- [最佳实践与性能考虑](#最佳实践与性能考虑)
- [典型应用场景](#典型应用场景)
- [参与贡献](#参与贡献)
- [开源协议](#开源协议)

---

## 为什么需要 gometa？

在 Go 语言中为代码添加元数据长期存在两难困境：

1. **结构体标签（Struct Tags）的局限**：
   - 弱类型字符串拼接，无法利用编译器进行类型与拼写检查；
   - 仅能作用于结构体字段，无法修饰类型本身、方法、入参或返回值；
   - IDE 无法提供自动补全与安全的重命名重构支持；
   - 运行时依赖重复反射解析，性能开销不可忽视。
2. **代码生成与注释指令（Code Generation & Magic Comments）的繁琐**：
   - 依赖特殊的注释语法（如 `// +k8s:...`、`// @Router`）；
   - 增加额外的构建流程与工具链依赖；
   - 生成的代码容易脱节或过时，维护成本高。

**`gometa` 的解决之道**：
- **100% 纯 Go 代码**：注解就是普通的 Go 结构体，全面享受编译期类型检查与 IDE 工具链支持。
- **全声明层级覆盖**：统一支持类型、字段、方法、参数（下标/名称）和返回值（下标/名称）。
- **双重校验防线**：启动期（Boot-time）反射校验 + 编译/CI 期（Compile-time）静态分析，拼写错误和越界直接暴露。
- **运行时零反射开销**：反射仅在进程启动的注册阶段执行一次，运行期查询为直接指针与哈希查找，适合高并发热路径。

---

## 核心特性

- 🛡️ **强类型注解**：普通 Go 结构体即可实现注解，支持类型约束与编译期补全。
- 🎯 **全层级目标支持**：覆盖类型（Type）、字段（Field）、方法（Method）、参数（Parameter）、返回值（Result）。
- 🔍 **双重绑定校验**：
  - **启动期**：`Register[T]` 校验字段、方法、参数及返回值下标，快速失败（Fail-Fast）。
  - **静态期**：配套 `cmd/gometa` 静态分析工具，集成 `go vet` 在编译和 CI 前提前拦截错误。
- ⚡ **零运行时反射开销**：初始化完成后，业务与框架消费元数据仅通过只读结构体和泛型辅助方法查询。
- 🔁 **确定性有序遍历**：内置 `RangeFields`、`RangeMethods`、`Walk`、`WalkAnnotations`，按稳定字典序或下标序迭代。
- 🌐 **全局类型发现**：提供 `Registrations()` 接口，便于框架全量扫描进程中注册的组件并实现自动装配。
- 📦 **轻量无额外侵入**：核心库仅依赖 Go 标准库（Go 1.22+），静态分析器复用官方 `golang.org/x/tools`。

---

## 方案对比

| 特性 | gometa | 结构体标签 (Struct Tag) | 代码生成 (go:generate) | 纯运行时反射 |
| :--- | :---: | :---: | :---: | :---: |
| **强类型检查** | ✅ 编译期强类型 | ❌ 弱类型字符串 | ⚠️ 依赖生成器 | ❌ 运行时断言 |
| **修饰类型/方法/参数/返回值** | ✅ 全面支持 | ❌ 仅支持结构体字段 | ⚠️ 依赖注释语法 | ⚠️ 支持受限 |
| **IDE 补全与安全重构** | ✅ 完美支持 | ❌ 无补全/易断裂 | ⚠️ 依赖外部插件 | ❌ 字符串驱动 |
| **运行时性能 (热路径)** | ✅ **零反射开销** | ❌ 频繁字符串/反射解析 | ✅ 原生代码 | ❌ 性能开销大 |
| **额外构建步骤** | ✅ **无需代码生成** | ✅ 原生支持 | ❌ 需要额外构建步骤 | ✅ 原生支持 |
| **提前捕获拼写/越界错误** | ✅ 静态分析 + 启动期 | ❌ 仅运行时静默失效 | ⚠️ 构建期生成失败 | ❌ 仅运行时报错 |

---

## 安装指南

### 1. 引入核心库

```bash
go get github.com/yvvlee/gometa
```

> **环境要求**：Go 1.22 或更高版本（使用了泛型与 `reflect.TypeFor` 特性）。

### 2. 安装配套静态检查工具（推荐）

```bash
go install github.com/yvvlee/gometa/cmd/gometa@latest
```

---

## 核心概念与架构

`gometa` 的架构分为三个阶段：**元数据声明与注册**、**双重校验**与**元数据消费**。

```
┌────────────────────────────────────────────────────────┐
│ 1. 声明强类型注解 (Annotation) 与元数据 (Register[T]) │
└──────────────────────────┬─────────────────────────────┘
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
┌─────────────────────────┐ ┌───────────────────────────┐
│ 2a. 编译/CI 期静态分析   │ │ 2b. 应用启动期反射校验    │
│ (go vet with gometa)    │ │ (Fail-fast DefinitionErr) │
└─────────────────────────┘ └─────────────┬─────────────┘
                                          ▼
┌────────────────────────────────────────────────────────┐
│ 3. 运行期元数据消费 (零反射)                           │
│ - MetadataOf[T]() / Registrations()                    │
│ - FindAnnotation[A]() / AllAnnotations[A]()            │
│ - RangeFields / RangeMethods / Walk / WalkAnnotations  │
└────────────────────────────────────────────────────────┘
```

### 声明目标 (`Target`)

每个注解必须实现 `gometa.Annotation` 接口，并通过位掩码指定允许生效的目标：

```go
type Annotation interface {
    Targets() Target
}
```

内置的目标掩码包括：
- `TargetType`：作用于类型声明。
- `TargetField`：作用于结构体字段。
- `TargetMethod`：作用于方法声明。
- `TargetParameter`：作用于方法入参。
- `TargetResult`：作用于方法出参/返回值。
- `TargetAll`：所有目标的组合掩码。

---

## 完整使用示例

本示例演示从定义注解、注册元数据到消费元数据的完整流程。完整可运行代码可参考 [examples/http](examples/http)。

### 1. 定义强类型注解

注解是实现了 `Targets() gometa.Target` 的普通 Go 类型：

```go
package main

import "github.com/yvvlee/gometa"

// Service 修饰类型
type Service struct {
	Name string
}
func (Service) Targets() gometa.Target { return gometa.TargetType }

// Inject 修饰字段
type Inject struct {
	Name string
}
func (Inject) Targets() gometa.Target { return gometa.TargetField }

// GET 修饰方法
type GET struct {
	Path string
}
func (GET) Targets() gometa.Target { return gometa.TargetMethod }

// Permission 修饰方法（支持重复注解）
type Permission struct {
	Name string
}
func (Permission) Targets() gometa.Target { return gometa.TargetMethod }

// Path 与 Header 修饰方法入参
type Path struct {
	Name string
}
func (Path) Targets() gometa.Target { return gometa.TargetParameter }

type Header struct {
	Name string
}
func (Header) Targets() gometa.Target { return gometa.TargetParameter }

// Body 修饰方法返回值
type Body struct{}
func (Body) Targets() gometa.Target { return gometa.TargetResult }

// Deprecated 同时修饰类型与方法（支持位组合）
type Deprecated struct {
	Message string
}
func (Deprecated) Targets() gometa.Target {
	return gometa.TargetType | gometa.TargetMethod
}
```

`Target` 还提供了常用辅助方法：

```go
target := gometa.TargetField | gometa.TargetParameter

target.Supports(gometa.TargetField) // true
target.Valid()                      // true
target.String()                     // "field|parameter"
```

---

### 2. 声明并注册元数据

使用 `gometa.Register[T]` 在包初始化时登记元数据。如果声明的字段名或方法名不存在，或参数下标越界，程序将在启动时立即 panic 报错：

```go
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

// 建议使用包级变量注册，在 package 初始化时自动执行校验
var userServiceMetadata = gometa.Register[UserService](
	// 1. 类型级注解
	Service{Name: "users"},
	Deprecated{Message: "推荐使用 UserServiceV2"},

	// 2. 字段级注解与元数据
	gometa.Field(
		"BaseURL",
		Inject{Name: "USER_SERVICE_BASE_URL"},
	),

	// 3. 方法级注解及参数/返回值元数据
	gometa.Method(
		"GetUser",
		GET{Path: "/users/:id"},
		Permission{Name: "user.read"},
		Permission{Name: "audit.read"}, // 支持同类型重复注解

		// 入参元数据：使用下标 (0, 1) 唯一定位，可附加可读性名称
		gometa.NamedParam(0, "id", Path{Name: "id"}),
		gometa.NamedParam(1, "authorization", Header{Name: "Authorization"}),

		// 出参元数据
		gometa.NamedResult(0, "user", Body{}),
		gometa.NamedResult(1, "err"),
	),
)
```

> **提示**：若不需要显式参数名称，可以直接使用匿名下标辅助函数：
> ```go
> gometa.Param(0, Path{Name: "id"})
> gometa.Result(0, Body{})
> ```

---

### 3. 读取与查询元数据

在框架处理请求或初始化路由时，通过强类型泛型接口快速读取元数据，**完全无反射开销**：

```go
// 1. 获取已注册的类型元数据
metadata, found := gometa.MetadataOf[UserService]()
if !found {
	panic("UserService 元数据未注册")
}

// 2. 查找类型注解（泛型查询指定注解类型）
service, found := gometa.FindAnnotation[Service](metadata.Annotations)
if found {
	println("Service Name:", service.Name)
}

// 3. 查找字段元数据与注解
if field, found := metadata.Field("BaseURL"); found {
	if inject, found := gometa.FindAnnotation[Inject](field.Annotations); found {
		println("Inject Env:", inject.Name)
	}
}

// 4. 查找方法元数据与注解
if method, found := metadata.Method("GetUser"); found {
	if get, found := gometa.FindAnnotation[GET](method.Annotations); found {
		println("Route:", get.Path)
	}

	// 查询所有重复出现的注解
	permissions := gometa.AllAnnotations[Permission](method.Annotations)
	for _, p := range permissions {
		println("Permission required:", p.Name)
	}

	// 5. 按下标查询参数注解
	if param, found := method.Parameter(0); found {
		if path, found := gometa.FindAnnotation[Path](param.Annotations); found {
			println("Param 0 path bind:", path.Name)
		}
	}

	// 6. 按下标查询返回值注解
	if result, found := method.Result(0); found {
		if _, hasBody := gometa.FindAnnotation[Body](result.Annotations); hasBody {
			println("Result 0 is response body")
		}
	}
}
```

---

### 4. 确定性迭代遍历

当编写通用框架（例如路由自动扫描、RPC 派发器）时，往往不知道具体字段名和方法名。`gometa` 提供了按稳定顺序遍历的 API：

#### 逐层遍历 (`Range...`)

- 字段与方法按名称升序（字典序）排列；
- 参数与返回值按下标升序排列；
- 回调函数返回 `false` 可随时中断遍历。

```go
metadata.RangeFields(func(field *gometa.FieldMetadata) bool {
	println("Field:", field.Name)
	return true
})

metadata.RangeMethods(func(method *gometa.MethodMetadata) bool {
	println("Method:", method.Name)

	method.RangeParameters(func(param *gometa.ParameterMetadata) bool {
		println("  Param:", param.Index, param.Name)
		return true
	})

	method.RangeResults(func(result *gometa.ResultMetadata) bool {
		println("  Result:", result.Index, result.Name)
		return true
	})
	return true
})
```

#### 树形全遍历 (`Walk` 与 `WalkAnnotations`)

使用 `Walk` 深度优先遍历完整的声明树，节点类型由 `gometa.Declaration` 描述：

```go
gometa.Walk(metadata, func(decl gometa.Declaration) bool {
	switch decl.Target {
	case gometa.TargetType:
		println("Type node, annotations:", len(decl.Annotations()))
	case gometa.TargetField:
		println("Field:", decl.Field.Name)
	case gometa.TargetMethod:
		println("Method:", decl.Method.Name)
	case gometa.TargetParameter:
		println("Parameter:", decl.Method.Name, decl.Parameter.Index)
	case gometa.TargetResult:
		println("Result:", decl.Method.Name, decl.Result.Index)
	}
	return true
})
```

若仅关心所有注解及其所属位置，使用 `WalkAnnotations` 更加简洁：

```go
gometa.WalkAnnotations(metadata, func(decl gometa.Declaration, ann gometa.Annotation) bool {
	switch a := ann.(type) {
	case GET:
		println("Registering route:", decl.Method.Name, a.Path)
	case Path:
		println("Binding path parameter:", decl.Method.Name, decl.Parameter.Index, a.Name)
	}
	return true
})
```

---

### 5. 全局类型发现与自动装配

框架往往需要扫描整个进程中所有注册过的服务以自动构建路由或依赖容器。`gometa` 提供了全局注册项只读视图：

```go
registrations := gometa.Registrations()

for _, reg := range registrations {
	println("Registered Type:", reg.Type.String())

	// 自动探测是否包含 Service 注解并注册
	if service, found := gometa.FindAnnotation[Service](reg.Metadata.Annotations); found {
		println("-> Discovered service component:", service.Name)
	}
}
```

`Registrations()` 返回按类型名称稳定排序的切片，并发安全。

---

## 静态检查工具 (Static Analyzer)

为了在代码编译、打包和 CI 阶段就发现潜在的元数据绑定错误，`gometa` 提供了专属的静态检查器。

### 安装与运行

```bash
# 安装命令行工具
go install github.com/yvvlee/gometa/cmd/gometa@latest

# 使用 go vet 执行静态检查
go vet -vettool="$(command -v gometa)" ./...
```

### CI 集成示例 (GitHub Actions)

在项目 CI 工作流中添加检查步骤（例如 `.github/workflows/ci.yml`）：

```yaml
name: CI

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Install gometa analyzer
        run: go install github.com/yvvlee/gometa/cmd/gometa@latest
      - name: Run gometa vet
        run: go vet -vettool=$(which gometa) ./...
```

### 静态检查器覆盖的规则

静态分析器通过 Go AST 与类型信息解析代码，能够在未启动程序前捕获以下错误：

1. ❌ **不存在的字段名**：`Field("NotExists")` 引用的字段在对应结构体中不存在；
2. ❌ **不存在的方法名**：`Method("NotExists")` 引用的方法在对应类型或指针接收者中不存在；
3. ❌ **参数下标越界**：`Param(2)` 超出了方法签名的入参数量；
4. ❌ **返回值下标越界**：`Result(1)` 超出了方法签名的返回值数量；
5. ❌ **参数/返回值下标重复**：在同一个方法声明中重复定义了相同的参数或返回值下标；
6. ❌ **字段/方法声明重复**：在同一个 `Register` 调用中重复声明了同名字段或方法；
7. ❌ **目标不是具名类型**：`Register[any]` 传入了非具体具名类型。

---

## 错误处理与 Fail-Fast 设计

错误的静态元数据属于代码逻辑错误，而非可恢复的业务异常。因此，`gometa` 遵循 **Fail-Fast（快速失败）** 原则：

- 若在声明构造期违反规则，构造函数会直接抛出 `*gometa.DefinitionError` 并中断初始化；
- 全局注册表确保每个类型只注册一次，重复注册相同类型将立即 panic；
- 典型报错示例：
  ```text
  panic: gometa: field BaseURL: GET annotation targets method, not field
  panic: gometa: method GetUser: parameter index 5 out of range for UserService.GetUser (2 parameters)
  panic: gometa: type: metadata for UserService is already registered
  ```

---

## API 速查表

| 分类 | 函数 / 类型 | 说明 |
| :--- | :--- | :--- |
| **注解目标** | `Target` | 目标位掩码类型（`uint32`） |
| | `TargetType`, `TargetField`, `TargetMethod` | 基础目标常数 |
| | `TargetParameter`, `TargetResult`, `TargetAll` | 基础目标常数及组合全量常数 |
| | `(t Target) Supports(Target) bool` | 判断目标是否包含指定目标 |
| | `(t Target) Valid() bool` | 判断目标掩码是否合法非空 |
| | `(t Target) String() string` | 格式化目标名称（如 `"field\|parameter"`） |
| **接口定义** | `Annotation` | 注解必须实现的接口：`Targets() Target` |
| **注册与获取** | `Register[T](parts ...any) *TypeMetadata` | 校验并注册类型 `T` 的元数据（启动期调用） |
| | `MetadataOf[T]() (*TypeMetadata, bool)` | 查询类型 `T` 已注册的元数据 |
| | `Registrations() []Registration` | 获取全局所有已注册项（按类型名升序排列） |
| **元数据构造** | `Field(name, ...Annotation) FieldMetadata` | 构造字段元数据 |
| | `Method(name, ...any) MethodMetadata` | 构造方法元数据（可包含注解、入参及出参元数据） |
| | `Param(index, ...Annotation) ParameterMetadata` | 构造匿名入参元数据 |
| | `NamedParam(index, name, ...Annotation)` | 构造具名入参元数据 |
| | `Result(index, ...Annotation) ResultMetadata` | 构造匿名出参元数据 |
| | `NamedResult(index, name, ...Annotation)` | 构造具名出参元数据 |
| **注解查询** | `FindAnnotation[A]([]Annotation) (A, bool)` | 泛型查找首个匹配类型的注解 |
| | `AllAnnotations[A]([]Annotation) []A` | 泛型查找所有匹配类型的注解列表 |
| **结构遍历** | `(m *TypeMetadata) RangeFields(func) ` | 按字段名字典序遍历 |
| | `(m *TypeMetadata) RangeMethods(func)` | 按方法名字典序遍历 |
| | `(m *MethodMetadata) RangeParameters(func)` | 按入参数下标升序遍历 |
| | `(m *MethodMetadata) RangeResults(func)` | 按出参数下标升序遍历 |
| | `Walk(metadata, func(Declaration) bool)` | 深度优先遍历完整元数据树 |
| | `WalkAnnotations(metadata, func)` | 遍历完整元数据树中的全部注解项 |

---

## 最佳实践与性能考虑

1. **注册时机**：
   建议将 `gometa.Register[T](...)` 赋值给包级变量（如 `var _ = gometa.Register[...]` 或 `var userMeta = gometa.Register[...]`）。这样可以保证元数据在 `init()` 阶段一次性构建与校验完成，并利用编译器保证顺序。
2. **指针与值接收者方法**：
   `gometa.Register[T]` 在反射校验时会自动检查值类型及其指针类型（`*T`）上的方法集。因此无论目标方法是值接收者还是指针接收者，均可正常校验与绑定。
3. **并发安全与运行时开销**：
   - 全局注册表内部使用 `sync.RWMutex` 保护；
   - 注册完成后的元数据对象（`TypeMetadata`、`MethodMetadata` 等）属于只读结构，并发读取完全无锁、无分配、无反射，极致轻量。

---

## 典型应用场景

- 🚀 **声明式 Web 路由与 API 网关**：通过 `GET`、`POST`、`Path`、`Header` 注解自动生成路由表、参数绑定与 OpenAPI / Swagger 文档。
- 🔑 **声明式权限与审计**：在服务方法上声明 `@Permission("...")` 或 `@Audit`，由统一拦截器/中间件读取并鉴权。
- 💉 **依赖注入与配置注入**：在结构体字段上声明 `@Inject`、`@Value("${ENV}")`，由容器自动组装依赖。
- 📊 **可观测性 (Metrics & Tracing)**：为特定方法打上追踪元数据，自动采集分布式链路 Span 和指标埋点。

---

## 参与贡献

欢迎提出 Issue、建议或提交 Pull Request！详细规范请参见 [CONTRIBUTING.md](CONTRIBUTING.md)。

本地开发与测试：

```bash
# 运行单元测试
go test -v -race ./...

# 运行静态检查
go vet ./...
```

---

## 开源协议

本项目采用 [MIT 许可证](LICENSE) 开源。
