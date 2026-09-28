// Package analyzer implements gometa's static declaration binding checks.
package analyzer

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const gometaPackagePath = "github.com/yvvlee/gometa"

// Analyzer checks that metadata registered with Register refers to real fields,
// methods, parameters, and results.
var Analyzer = &analysis.Analyzer{
	Name: "gometa",
	Doc:  "check gometa declaration metadata bindings",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, declaration := range file.Decls {
			switch value := declaration.(type) {
			case *ast.FuncDecl:
				inspect(pass, value.Body)
			case *ast.GenDecl:
				inspect(pass, value)
			}
		}
	}
	return nil, nil
}

func inspect(pass *analysis.Pass, node ast.Node) {
	if node == nil {
		return
	}
	ast.Inspect(node, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isGometaCall(pass, call, "Register") {
			return true
		}
		checkRegister(pass, call)
		return true
	})
}

func checkRegister(pass *analysis.Pass, call *ast.CallExpr) {
	identifier := calledIdentifier(call.Fun)
	instance, ok := pass.TypesInfo.Instances[identifier]
	if !ok || instance.TypeArgs.Len() != 1 {
		pass.Reportf(call.Fun.Pos(), "cannot resolve gometa.Register type argument")
		return
	}

	target := namedType(instance.TypeArgs.At(0))
	if target == nil {
		pass.Reportf(call.Fun.Pos(), "gometa.Register target must be a named Go type")
		return
	}

	fields := make(map[string]ast.Expr)
	methods := make(map[string]ast.Expr)
	for _, part := range call.Args {
		builder, ok := part.(*ast.CallExpr)
		if !ok {
			continue
		}
		switch {
		case isGometaCall(pass, builder, "Field"):
			name, ok := stringArgument(pass, builder, 0, "field")
			if !ok {
				continue
			}
			if previous, duplicate := fields[name]; duplicate {
				pass.Reportf(builder.Args[0].Pos(), "duplicate gometa field %q; first declared at %s", name, pass.Fset.Position(previous.Pos()))
				continue
			}
			fields[name] = builder.Args[0]
			if !hasDirectField(target, name) {
				pass.Reportf(builder.Args[0].Pos(), "unknown field %s.%s", typeName(target), name)
			}
		case isGometaCall(pass, builder, "Method"):
			name, ok := stringArgument(pass, builder, 0, "method")
			if !ok {
				continue
			}
			if previous, duplicate := methods[name]; duplicate {
				pass.Reportf(builder.Args[0].Pos(), "duplicate gometa method %q; first declared at %s", name, pass.Fset.Position(previous.Pos()))
				continue
			}
			methods[name] = builder.Args[0]
			signature := methodSignature(pass.Pkg, target, name)
			if signature == nil {
				pass.Reportf(builder.Args[0].Pos(), "unknown method %s.%s", typeName(target), name)
				continue
			}
			checkMethodParts(pass, target, name, signature, builder.Args[1:])
		}
	}
}

func checkMethodParts(pass *analysis.Pass, target *types.Named, methodName string, signature *types.Signature, parts []ast.Expr) {
	parameters := make(map[int]ast.Expr)
	results := make(map[int]ast.Expr)
	for _, part := range parts {
		builder, ok := part.(*ast.CallExpr)
		if !ok {
			continue
		}

		switch {
		case isGometaCall(pass, builder, "Param"), isGometaCall(pass, builder, "NamedParam"):
			index, ok := integerArgument(pass, builder, 0, "parameter")
			if !ok {
				continue
			}
			if previous, duplicate := parameters[index]; duplicate {
				pass.Reportf(builder.Args[0].Pos(), "duplicate parameter index %d; first declared at %s", index, pass.Fset.Position(previous.Pos()))
				continue
			}
			parameters[index] = builder.Args[0]
			if index < 0 || index >= signature.Params().Len() {
				pass.Reportf(
					builder.Args[0].Pos(),
					"parameter index %d out of range for %s.%s (%d parameters)",
					index,
					typeName(target),
					methodName,
					signature.Params().Len(),
				)
			}
		case isGometaCall(pass, builder, "Result"), isGometaCall(pass, builder, "NamedResult"):
			index, ok := integerArgument(pass, builder, 0, "result")
			if !ok {
				continue
			}
			if previous, duplicate := results[index]; duplicate {
				pass.Reportf(builder.Args[0].Pos(), "duplicate result index %d; first declared at %s", index, pass.Fset.Position(previous.Pos()))
				continue
			}
			results[index] = builder.Args[0]
			if index < 0 || index >= signature.Results().Len() {
				pass.Reportf(
					builder.Args[0].Pos(),
					"result index %d out of range for %s.%s (%d results)",
					index,
					typeName(target),
					methodName,
					signature.Results().Len(),
				)
			}
		}
	}
}

func namedType(value types.Type) *types.Named {
	if value == nil {
		return nil
	}
	value = types.Unalias(value)
	if pointer, ok := value.(*types.Pointer); ok {
		value = types.Unalias(pointer.Elem())
	}
	named, _ := value.(*types.Named)
	return named
}

func hasDirectField(named *types.Named, name string) bool {
	structure, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for index := 0; index < structure.NumFields(); index++ {
		if structure.Field(index).Name() == name {
			return true
		}
	}
	return false
}

func methodSignature(pkg *types.Package, named *types.Named, name string) *types.Signature {
	for _, value := range []types.Type{named, types.NewPointer(named)} {
		selection := types.NewMethodSet(value).Lookup(pkg, name)
		if selection == nil {
			continue
		}
		function, ok := selection.Obj().(*types.Func)
		if !ok {
			continue
		}
		signature, _ := function.Type().(*types.Signature)
		if signature != nil {
			return signature
		}
	}
	return nil
}

func stringArgument(pass *analysis.Pass, call *ast.CallExpr, index int, kind string) (string, bool) {
	if len(call.Args) <= index {
		pass.Reportf(call.Pos(), "gometa %s name is missing", kind)
		return "", false
	}
	value := pass.TypesInfo.Types[call.Args[index]].Value
	if value == nil || value.Kind() != constant.String {
		pass.Reportf(call.Args[index].Pos(), "gometa %s name must be a compile-time string constant", kind)
		return "", false
	}
	return constant.StringVal(value), true
}

func integerArgument(pass *analysis.Pass, call *ast.CallExpr, index int, kind string) (int, bool) {
	if len(call.Args) <= index {
		pass.Reportf(call.Pos(), "gometa %s index is missing", kind)
		return 0, false
	}
	value := pass.TypesInfo.Types[call.Args[index]].Value
	if value == nil || value.Kind() != constant.Int {
		pass.Reportf(call.Args[index].Pos(), "gometa %s index must be a compile-time integer constant", kind)
		return 0, false
	}
	integer, exact := constant.Int64Val(value)
	if !exact || int64(int(integer)) != integer {
		pass.Reportf(call.Args[index].Pos(), "gometa %s index does not fit in int", kind)
		return 0, false
	}
	return int(integer), true
}

func isGometaCall(pass *analysis.Pass, call *ast.CallExpr, name string) bool {
	identifier := calledIdentifier(call.Fun)
	if identifier == nil {
		return false
	}
	function, ok := pass.TypesInfo.Uses[identifier].(*types.Func)
	return ok && function.Name() == name && function.Pkg() != nil && function.Pkg().Path() == gometaPackagePath
}

func calledIdentifier(expression ast.Expr) *ast.Ident {
	for {
		switch value := expression.(type) {
		case *ast.IndexExpr:
			expression = value.X
		case *ast.IndexListExpr:
			expression = value.X
		case *ast.ParenExpr:
			expression = value.X
		case *ast.SelectorExpr:
			return value.Sel
		case *ast.Ident:
			return value
		default:
			return nil
		}
	}
}

func typeName(named *types.Named) string {
	if named == nil || named.Obj() == nil {
		return "<unknown>"
	}
	if named.TypeArgs() == nil || named.TypeArgs().Len() == 0 {
		return named.Obj().Name()
	}
	return types.TypeString(named, func(*types.Package) string { return "" })
}
