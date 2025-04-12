package returnearly

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "returnearly",
	Doc:  "suggests inverting conditions to follow the return early pattern",
	Run:  run,
}

func position(pass *analysis.Pass, node ast.Node) string {
	if node == nil {
		return ""
	}
	if pass == nil {
		return ""
	}
	pos := pass.Fset.Position(node.Pos())
	return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			return inspectNode(pass, file, n)
		})
	}
	return nil, nil
}

func inspectNode(pass *analysis.Pass, file *ast.File, n ast.Node) bool {
	ifStmt, report := n.(*ast.IfStmt)
	if !report || ifStmt.Body == nil || ifStmt.Else != nil {
		return true // we only care about if-without-else
	}

	//fmt.Printf("found condition: %s\n", position(pass, ifStmt))

	// Find statements that follow the if block in the parent block
	parentBlock, parentFunc := findEnclosingBlock(pass, file, ifStmt)
	//fmt.Printf("- parentBlock: %s\n", position(pass, parentBlock))
	if parentBlock == nil {
		return true
	}

	thenLen := countRelevantStatements(ifStmt.Body.List)
	if thenLen == 0 {
		return true
	}

	afterLen := 0
	seen := false
	afterTerminates := parentFunc
	for _, stmt := range parentBlock.List {
		if !seen {
			if stmt == ifStmt {
				seen = true
			}
			continue
		}
		afterLen += countRelevantStatements([]ast.Stmt{stmt})
		if isTerminalStmt(stmt) {
			afterTerminates = true
			break
		}
	}

	//if parentFunc {
	//	afterLen++
	//}

	bodyTerminates := blockEndsWithTerminal(ifStmt.Body.List)
	report = thenLen > afterLen && (bodyTerminates || afterLen == 0) && afterTerminates

	//fmt.Printf("- afterLen: %d\n", afterLen)
	//fmt.Printf("- thenLen: %d\n", thenLen)
	//fmt.Printf("- bodyTerminates: %v\n", bodyTerminates)
	//fmt.Printf("- afterTerminates: %v\n", afterTerminates)
	//fmt.Printf("- parentFunc: %v\n", parentFunc)
	//fmt.Printf("- report: %v\n", report)

	if report {
		pass.Reportf(ifStmt.Pos(), "consider inverting the condition to return early and avoid nesting (then: %d > after: %d)", thenLen, afterLen)
	}

	return true
}

func countRelevantStatements(stmts []ast.Stmt) (count int) {
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.EmptyStmt:
			continue
		case *ast.ReturnStmt, *ast.BranchStmt:
			count++
			return
		default:
			count++
		}
	}
	return
}

func blockEndsWithTerminal(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	last := stmts[len(stmts)-1]
	return isTerminalStmt(last)
}

func isTerminalStmt(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	case *ast.ExprStmt:
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			break
		}
		if fun, ok := call.Fun.(*ast.Ident); ok {
			if fun.Name == "panic" {
				return true
			}
			break
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			id, ok := sel.X.(*ast.Ident)
			if ok && id.Name == "os" && sel.Sel.Name == "Exit" {
				return true
			}
			break
		}
	}
	return false
}

func findEnclosingBlock(pass *analysis.Pass, root ast.Node, target ast.Stmt) (*ast.BlockStmt, bool) {
	var enclosing *ast.BlockStmt
	var enclosingFunc bool

	ast.Inspect(root, func(n ast.Node) bool {
		var block *ast.BlockStmt
		funcDecl, isFunc := n.(*ast.FuncDecl)

		if isFunc {
			block = funcDecl.Body
		} else {
			var ok bool
			block, ok = n.(*ast.BlockStmt)
			if !ok {
				return true
			}
		}

		for _, stmt := range block.List {
			if stmt == target {
				enclosing = block
				enclosingFunc = isFunc
				return false
			}
		}

		return true
	})

	return enclosing, enclosingFunc
}
