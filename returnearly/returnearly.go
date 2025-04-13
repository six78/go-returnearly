package returnearly

import (
	"flag"
	"fmt"
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "returnearly",
	Doc:  "suggests inverting conditions to follow the return early pattern",
	Run:  run,
}

func position(pass *analysis.Pass, node ast.Node) string {
	if node == nil || reflect.ValueOf(node).IsNil() {
		return "nil"
	}
	//if pass == nil {
	//	return ""
	//}
	//if pass.Fset == nil {
	//	return ""
	//}
	p := node.Pos()
	pos := pass.Fset.Position(p)
	return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Create a stack to track parent nodes
		parentStack := &ParentStack{}

		ast.Inspect(file, func(n ast.Node) bool {
			// Pop the stack when exiting a node
			if n == nil {
				fmt.Printf("- pop parent\n")
				parentStack.Pop()
				return true
			}

			// Push the current node to the stack

			result := inspectNode(pass, file, n, parentStack)

			fmt.Printf("- push parent: type = %s, value = %v\n", reflect.TypeOf(n), n)
			parentStack.Push(n)

			return result
		})
	}
	return nil, nil
}

func inspectNode(pass *analysis.Pass, file *ast.File, n ast.Node, stack *ParentStack) bool {
	ifStmt, report := n.(*ast.IfStmt)
	if !report || ifStmt.Body == nil || ifStmt.Else != nil {
		return true // we only care about if-without-else
	}

	fmt.Println("------------")
	fmt.Printf("found condition: %s\n", position(pass, ifStmt))

	// Get the direct parent of the `ifStmt` from the stack
	parentBlock, parentFunc, isFuncBody := stack.ParentInfo()
	if parentBlock == nil {
		fmt.Println("no enclosing block found")
		return true
	}

	//isFuncBody := stack.IsFuncBody()
	//if !isFuncBody {
	//	fmt.Println("Block is not part of a function body")
	//	return true
	//}

	//parentFunc := stack.ParentFunc()

	//// If no parentBlock block is found, skip the node
	//if parentBlock == nil {
	//	fmt.Printf("no parent block found\n")
	//	return true
	//}

	// Find statements that follow the if block in the parent block
	//parentBlock, parentFunc := findEnclosingBlock(pass, file, ifStmt)
	fmt.Printf("- parentBlock: %s\n", position(pass, parentBlock))
	//if parentBlock == nil {
	//	return true
	//}

	//thenLen, _ := countRelevantStatements(ifStmt.Body.List, true)
	thenStart := pass.Fset.Position(ifStmt.Body.List[0].Pos()).Line
	thenEnd := pass.Fset.Position(ifStmt.Body.List[len(ifStmt.Body.List)-1].End()).Line
	thenLines := thenEnd - thenStart + 1
	thenLen := thenLines

	if thenLen == 0 {
		return true
	}

	//afterLen := 0
	//seen := false
	//afterTerminates := isFuncBody
	//returned := false

	//var stmtStart, stmtEnd token.Position

	afterStart := pass.Fset.Position(ifStmt.Body.End()).Line + 1 // increment, because the after block starts next line
	afterEnd := pass.Fset.Position(parentBlock.End()).Line - 1   // decrement to exclude the `}` block termination line
	afterLen := afterEnd - afterStart + 1

	// Take the last block as terminal
	lastStatement := parentBlock.List[len(parentBlock.List)-1]
	returned := lastStatement != nil && isTerminalStmt(lastStatement)
	afterTerminates := isFuncBody || returned

	//// Check if there's a terminal statement before the end
	//for i, stmt := range parentBlock.List {
	//	if !seen {
	//		// Mark when we reach the `if` statement
	//		if stmt == ifStmt {
	//			seen = true
	//		}
	//		continue
	//	}
	//	//afterLen, returned = countRelevantStatements([]ast.Stmt{stmt}, true)
	//
	//	//stmtStart, stmtEnd = pass.Fset.Position(stmt.Pos()), pass.Fset.Position(stmt.End())
	//	//afterLines := stmtEnd.Line - stmtStart.Line
	//	//afterLen += afterLines
	//
	//	//fmt.Printf("- parent list iteration %d: %s, added = %d, total = %d\n", i, position(pass, stmt), afterLines, afterLen)
	//
	//	if isTerminalStmt(stmt) {
	//		afterTerminates = true
	//		break
	//	}
	//}

	emptyAfter := afterLen == 0

	if isFuncBody && !returned {
		afterLen++
	}

	bodyTerminates := blockEndsWithTerminal(ifStmt.Body.List)
	report = thenLen > afterLen && (bodyTerminates || emptyAfter) && afterTerminates

	fmt.Printf("- THEN block: start = %d, end = %d, len = %d\n", thenStart, thenEnd, thenLines)
	fmt.Printf("- AFTER block: start = %d, end = %d, len = %d\n", afterStart, afterEnd, afterLen)
	fmt.Printf("- TERMINATES: body = %v, after = %v\n", bodyTerminates, afterTerminates)
	fmt.Printf("- parentFunc: %v\n", parentFunc)
	fmt.Printf("- isFuncBody: %v\n", isFuncBody)
	fmt.Printf("- report: %v\n", report)

	if report {
		pass.Reportf(ifStmt.Pos(), "consider inverting the condition to return early and avoid nesting (then: %d > after: %d)", thenLen, afterLen)
	}

	return true
}

func countRelevantStatements(stmts []ast.Stmt, countReturns bool) (count int, returned bool) {
	returned = false
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.EmptyStmt:
			continue
		case *ast.ReturnStmt, *ast.BranchStmt:
			if countReturns {
				count++
			}
			returned = true
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
