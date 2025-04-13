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

	thenStart := pass.Fset.Position(ifStmt.Body.List[0].Pos()).Line
	thenEnd := pass.Fset.Position(ifStmt.Body.List[len(ifStmt.Body.List)-1].End()).Line
	thenLines := thenEnd - thenStart + 1
	thenLen := thenLines

	if thenLen == 0 {
		return true
	}

	afterStart := pass.Fset.Position(ifStmt.Body.End()).Line + 1 // increment, because the after block starts next line
	afterEnd := pass.Fset.Position(parentBlock.End()).Line - 1   // decrement to exclude the `}` block termination line
	afterLen := afterEnd - afterStart + 1

	lastStmt := parentBlock.List[len(parentBlock.List)-1]
	returned := isTerminalStmt(lastStmt)
	afterTerminates := isFuncBody || returned
	afterIsEmpty := afterLen == 0

	// When the function doesn't have explicit return statement increment the `afterLen`.
	// Because is we would to invert the condition, a return statement would appear.
	if isFuncBody && !returned {
		afterLen++
	}

	bodyTerminates := blockEndsWithTerminal(ifStmt.Body.List)
	report = thenLen > afterLen && (bodyTerminates || afterIsEmpty) && afterTerminates

	logger.Printf("- THEN block: start = %d, end = %d, len = %d\n", thenStart, thenEnd, thenLines)
	logger.Printf("- AFTER block: start = %d, end = %d, len = %d\n", afterStart, afterEnd, afterLen)
	logger.Printf("- TERMINATES: body = %v, after = %v\n", bodyTerminates, afterTerminates)
	logger.Printf("- parentFunc = %v, isFuncBody = %v, report = %v\n", parentFunc, isFuncBody, report)

	if report {
		pass.Reportf(ifStmt.Pos(), "consider inverting the condition to return early and avoid nesting (then: %d > after: %d)", thenLen, afterLen)
	}

	return true
}

func blockEndsWithTerminal(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	last := stmts[len(stmts)-1]
	return isTerminalStmt(last)
}

func isTerminalStmt(stmt ast.Stmt) bool {
	if stmt == nil {
		return false
	}
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
