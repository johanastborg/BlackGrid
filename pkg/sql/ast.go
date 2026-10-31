package sql

import (
	"fmt"
	"strings"
)

// Node represents any AST node.
type Node interface {
	String() string
}

// Expr represents any expression node in SQL.
type Expr interface {
	Node
	exprNode()
}

// Statement represents a top-level SQL statement.
type Statement interface {
	Node
	stmtNode()
}

// SelectStmt represents a SELECT query in GoogleSQL.
type SelectStmt struct {
	Fields []string
	Table  string
	Where  Expr
	Limit  int
}

func (s *SelectStmt) stmtNode() {}
func (s *SelectStmt) String() string {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	if len(s.Fields) == 0 {
		sb.WriteString("*")
	} else {
		sb.WriteString(strings.Join(s.Fields, ", "))
	}
	sb.WriteString(" FROM ")
	sb.WriteString(s.Table)
	if s.Where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.Where.String())
	}
	if s.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", s.Limit))
	}
	return sb.String()
}

// ColRef represents a column reference (e.g. `latitude`).
type ColRef struct {
	Name string
}

func (c *ColRef) exprNode()      {}
func (c *ColRef) String() string { return c.Name }

// NumberLit represents a numeric literal (integer or float).
type NumberLit struct {
	Value float64
	Raw   string
}

func (n *NumberLit) exprNode()      {}
func (n *NumberLit) String() string { return n.Raw }

// StringLit represents a string literal.
type StringLit struct {
	Value string
}

func (s *StringLit) exprNode()      {}
func (s *StringLit) String() string { return fmt.Sprintf("'%s'", s.Value) }

// BinaryExpr represents a binary operator expression (e.g. `a >= 10`, `x AND y`).
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op, b.Right.String())
}

// BetweenExpr represents a `col BETWEEN low AND high` expression.
type BetweenExpr struct {
	Expr Expr
	Low  Expr
	High Expr
	Not  bool
}

func (b *BetweenExpr) exprNode() {}
func (b *BetweenExpr) String() string {
	prefix := ""
	if b.Not {
		prefix = "NOT "
	}
	return fmt.Sprintf("(%s %sBETWEEN %s AND %s)", b.Expr.String(), prefix, b.Low.String(), b.High.String())
}

// ParenExpr represents an expression inside parentheses `(expr)`.
type ParenExpr struct {
	Expr Expr
}

func (p *ParenExpr) exprNode()      {}
func (p *ParenExpr) String() string { return fmt.Sprintf("(%s)", p.Expr.String()) }

