package mainexcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "mainexitcheck",
	Doc:  "check that main function does not exit via os.Exit",
	Run:  run,
}

const (
	checkerr  = "main function should not exit via os.Exit"
	mainIdent = "main"
	osIdent   = "os"
	exitIdent = "Exit"
)

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != mainIdent {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Name.Name != mainIdent {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			fnc, ok := node.(*ast.FuncDecl)
			if !ok || fnc.Name.Name != mainIdent {
				return true
			}
			ast.Inspect(fnc.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok || ident.Name != osIdent {
					return true
				}

				if sel.Sel.Name != exitIdent {
					return true
				}

				pass.Reportf(call.Pos(), checkerr)
				return false
			})
			return false
		})
	}

	return nil, nil
}
