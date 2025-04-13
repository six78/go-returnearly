package returnearly

import (
	"flag"
	"fmt"
	"go/ast"
	"io"
	"log"
	"os"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

var (
	Analyzer = &analysis.Analyzer{
		Name: "returnearly",
		Doc:  "suggests inverting conditions to follow the return early pattern",
		Run:  run,
		Flags: func() flag.FlagSet {
			fs := flag.NewFlagSet("returnearly", flag.ExitOnError)
			fs.BoolVar(&verboseOutput, "verbose", false, "enable verbose output")
			return *fs
		}(),
	}

	verboseOutput = false
	logger        = log.New(io.Discard, "", 0)
)

func position(pass *analysis.Pass, node ast.Node) string {
	if node == nil || reflect.ValueOf(node).IsNil() {
		return "nil"
	}
	pos := pass.Fset.Position(node.Pos())
	return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
}

func run(pass *analysis.Pass) (interface{}, error) {
	if verboseOutput {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
		logger.SetOutput(os.Stdout)
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Create a stack to track parent nodes
		parentStack := &ParentStack{}

		// Traverse the node
		ast.Inspect(file, func(n ast.Node) bool {
			// Pop the stack when exiting a node
			if n == nil {
				parentStack.Pop()
				return true
			}

			// Analyze the node
			result := analyzeNode(pass, file, n, parentStack)

			// Push the current node to the stack
			parentStack.Push(n)

			return result
		})
	}

	return nil, nil
}

func analyzeNode(pass *analysis.Pass, file *ast.File, n ast.Node, stack *ParentStack) bool {
	ifStmt, report := n.(*ast.IfStmt)
	if !report || ifStmt.Body == nil || ifStmt.Else != nil {
		return true // we only care about if-without-else
	}

	logger.Println("------------")
	logger.Printf("found condition: %s\n", position(pass, ifStmt))

	// Get the direct parent of the `ifStmt` from the stack
	parentBlock, parentFunc, isFuncBody := stack.ParentInfo()
	if parentBlock == nil {
		logger.Println("no enclosing block found")
		return true
	}

	//isFuncBody := stack.IsFuncBody()
	//if !isFuncBody {
	//	logger.Println("Block is not part of a function body")
	//	return true
	//}

	//parentFunc := stack.ParentFunc()

	//// If no parentBlock block is found, skip the node
	//if parentBlock == nil {
	//	logger.Printf("no parent block found\n")
	//	return true
	//}

	// Find statements that follow the if block in the parent block
	//parentBlock, parentFunc := findEnclosingBlock(pass, file, ifStmt)
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
	//	//logger.Printf("- parent list iteration %d: %s, added = %d, total = %d\n", i, position(pass, stmt), afterLines, afterLen)
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

	logger.Printf("- THEN block: start = %d, end = %d, len = %d\n", thenStart, thenEnd, thenLines)
	logger.Printf("- AFTER block: start = %d, end = %d, len = %d\n", afterStart, afterEnd, afterLen)
	logger.Printf("- TERMINATES: body = %v, after = %v\n", bodyTerminates, afterTerminates)
	logger.Printf("- parentFunc = %v, isFuncBody = %v, report = %v\n", parentFunc, isFuncBody, report)

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
