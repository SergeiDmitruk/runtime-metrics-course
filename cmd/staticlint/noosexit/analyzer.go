package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `noosexit checks for direct calls to os.Exit in main function of main package

This analyzer reports any direct calls to os.Exit in the main function
of the main package, as they can prevent proper cleanup and testing.`

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  doc,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fnDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}

		if fnDecl.Name.Name != "main" {
			return
		}

		ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "direct call to os.Exit in main function of main package")
			}

			return true
		})
	})

	return nil, nil
}
