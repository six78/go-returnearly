package returnearly

import "go/ast"

type ParentStack struct {
	stack []ast.Node
}

// Push adds a node to the stack
func (ps *ParentStack) Push(node ast.Node) {
	ps.stack = append(ps.stack, node)
}

// Pop removes and returns the last node from the stack
func (ps *ParentStack) Pop() ast.Node {
	if len(ps.stack) == 0 {
		return nil
	}
	node := ps.stack[len(ps.stack)-1]
	ps.stack = ps.stack[:len(ps.stack)-1]
	return node
}

// Peek returns the last node without removing it
func (ps *ParentStack) Peek() ast.Node {
	if len(ps.stack) == 0 {
		return nil
	}
	return ps.stack[len(ps.stack)-1]
}

// ParentBlock returns the nearest enclosing *ast.BlockStmt node.
func (ps *ParentStack) ParentBlock() *ast.BlockStmt {
	if len(ps.stack) == 0 {
		return nil
	}
	// Check the last node in the stack
	if blockStmt, ok := ps.Peek().(*ast.BlockStmt); ok {
		return blockStmt
	}
	return nil
}

// ParentFunc returns the nearest enclosing *ast.FuncDecl node.
func (ps *ParentStack) ParentFunc() *ast.FuncDecl {
	if len(ps.stack) < 2 {
		return nil
	}
	// Check the second-to-last node in the stack
	if funcDecl, ok := ps.stack[len(ps.stack)-2].(*ast.FuncDecl); ok {
		return funcDecl
	}
	return nil
}

func (ps *ParentStack) ParentInfo() (*ast.BlockStmt, *ast.FuncDecl, bool) {
	funcDecl := ps.ParentFunc()
	blockStmt := ps.ParentBlock()
	return blockStmt, funcDecl, funcDecl != nil && blockStmt != nil && blockStmt == funcDecl.Body
}

// Len returns the size of the stack
func (ps *ParentStack) Len() int {
	return len(ps.stack)
}
