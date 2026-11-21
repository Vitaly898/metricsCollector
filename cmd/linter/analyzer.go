package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// Analyzer reports all builtin panic calls. It also reports log.Fatal/Fatalf/Fatalln
// and os.Exit calls outside the main function of package main.
var Analyzer = &analysis.Analyzer{
	Name: "linter",
	Doc:  "forbid builtin panic and restrict log.Fatal/Fatalf/Fatalln and os.Exit to main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Walk(visitor{pass: pass}, file)
	}
	return nil, nil
}

type visitor struct {
	pass   *analysis.Pass
	inMain bool
}

// A value receiver keeps the enclosing function's context for sibling nodes.
func (v visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		v.inMain = v.pass.Pkg.Name() == "main" && node.Name.Name == "main" && node.Recv == nil
	case *ast.FuncLit:
		v.inMain = false
	case *ast.CallExpr:
		v.checkCall(node)
	}
	return v
}

func (v visitor) checkCall(call *ast.CallExpr) {
	switch obj := typeutil.Callee(v.pass.TypesInfo, call).(type) {
	case *types.Builtin:
		if obj.Name() == "panic" {
			v.pass.Reportf(call.Pos(), "use of builtin panic is forbidden")
		}
	case *types.Func:
		if v.inMain || obj.Pkg() == nil || obj.Signature().Recv() != nil {
			return
		}
		var forbidden bool
		switch obj.Pkg().Path() {
		case "log":
			forbidden = obj.Name() == "Fatal" || obj.Name() == "Fatalf" || obj.Name() == "Fatalln"
		case "os":
			forbidden = obj.Name() == "Exit"
		}
		if forbidden {
			v.pass.Reportf(call.Pos(), "%s.%s is only allowed in main.main", obj.Pkg().Name(), obj.Name())
		}
	}
}
