package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/quickfix/qf1001"
	"honnef.co/go/tools/simple/s1001"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1001"
)

// MainOsExit - Checking if os.Exit is called in main func
var MainOsExit = &analysis.Analyzer{
	Name: "main_os_exit",
	Doc:  "Checking if os.Exit is called in main func",
	Run:  run,
}

// run analyses files to check for os.Exit calls in main function
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Scope() == types.Universe {
		return nil, nil
	}
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if strings.Contains(pass.Fset.Position(file.Package).Filename, "go-build") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				inspectMainFunctionBody(pass, fn)
			}
			return true
		})
	}
	return nil, nil
}

// inspectMainFunctionBody searches for os.Exit calls within the main function body
func inspectMainFunctionBody(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Body == nil {
		return
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok && isOsExitCall(call) {
			pass.Reportf(n.Pos(), "the use of os.Exit is prohibited in the main function")
		}
		return true
	})
}

// isOsExitCall checks if the call expression is os.Exit
func isOsExitCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "os" && sel.Sel.Name == "Exit"
}

// Запуск make staticlint
// Описание анализаторов ./bin/staticlint help
func main() {
	checks := []*analysis.Analyzer{
		timeformat.Analyzer,
		unusedresult.Analyzer,
		waitgroup.Analyzer,
		MainOsExit,
	}
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	checks = append(checks, s1001.Analyzer)
	checks = append(checks, qf1001.Analyzer)
	checks = append(checks, st1001.Analyzer)
	multichecker.Main(checks...)
}
