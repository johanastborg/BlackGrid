package sql

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser parses GoogleSQL queries into an AST.
type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

// NewParser creates a new Parser for the given SQL string.
func NewParser(input string) *Parser {
	l := NewLexer(input)
	p := &Parser{lexer: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
}

// Parse parses a SQL statement (currently SELECT).
func (p *Parser) Parse() (Statement, error) {
	if p.cur.Type == TokenKeyword && p.cur.Value == "SELECT" {
		return p.parseSelect()
	}
	return nil, fmt.Errorf("expected SELECT at position %d, got %q", p.cur.Pos, p.cur.Value)
}

func (p *Parser) parseSelect() (*SelectStmt, error) {
	p.nextToken() // consume SELECT

	stmt := &SelectStmt{}

	// Parse fields
	for {
		if p.cur.Type == TokenStar {
			stmt.Fields = append(stmt.Fields, "*")
			p.nextToken()
		} else if p.cur.Type == TokenIdent {
			stmt.Fields = append(stmt.Fields, p.cur.Value)
			p.nextToken()
		} else {
			return nil, fmt.Errorf("unexpected token in field list: %q", p.cur.Value)
		}

		if p.cur.Type == TokenComma {
			p.nextToken()
			continue
		}
		break
	}

	// Expect FROM
	if p.cur.Type != TokenKeyword || p.cur.Value != "FROM" {
		return nil, fmt.Errorf("expected FROM, got %q", p.cur.Value)
	}
	p.nextToken()

	// Table name
	if p.cur.Type != TokenIdent {
		return nil, fmt.Errorf("expected table name, got %q", p.cur.Value)
	}
	stmt.Table = p.cur.Value
	p.nextToken()

	// Optional WHERE
	if p.cur.Type == TokenKeyword && p.cur.Value == "WHERE" {
		p.nextToken()
		whereExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = whereExpr
	}

	// Optional LIMIT
	if p.cur.Type == TokenKeyword && p.cur.Value == "LIMIT" {
		p.nextToken()
		if p.cur.Type != TokenNumber {
			return nil, fmt.Errorf("expected number after LIMIT, got %q", p.cur.Value)
		}
		limitVal, err := strconv.Atoi(p.cur.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT number %q", p.cur.Value)
		}
		stmt.Limit = limitVal
		p.nextToken()
	}

	return stmt, nil
}

// Expression parsing: precedence hierarchy
// parseExpr -> parseOr -> parseAnd -> parseComparison -> parsePrimary

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Expr, error) {
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == TokenKeyword && p.cur.Value == "OR" {
		p.nextToken()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExpr{Left: expr, Op: "OR", Right: right}
	}
	return expr, nil
}

func (p *Parser) parseAnd() (Expr, error) {
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == TokenKeyword && p.cur.Value == "AND" {
		p.nextToken()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExpr{Left: expr, Op: "AND", Right: right}
	}
	return expr, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Check for BETWEEN
	if p.cur.Type == TokenKeyword && (p.cur.Value == "BETWEEN" || p.cur.Value == "NOT") {
		isNot := false
		if p.cur.Value == "NOT" {
			isNot = true
			p.nextToken()
			if p.cur.Type != TokenKeyword || p.cur.Value != "BETWEEN" {
				return nil, fmt.Errorf("expected BETWEEN after NOT, got %q", p.cur.Value)
			}
		}
		p.nextToken() // consume BETWEEN

		low, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		if p.cur.Type != TokenKeyword || p.cur.Value != "AND" {
			return nil, fmt.Errorf("expected AND in BETWEEN clause, got %q", p.cur.Value)
		}
		p.nextToken() // consume AND

		high, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		return &BetweenExpr{
			Expr: left,
			Low:  low,
			High: high,
			Not:  isNot,
		}, nil
	}

	// Check for comparison operators: =, !=, <>, <, <=, >, >=
	switch p.cur.Type {
	case TokenEq, TokenNeq, TokenLt, TokenLte, TokenGt, TokenGte:
		op := p.cur.Value
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Op: op, Right: right}, nil
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	switch p.cur.Type {
	case TokenLParen:
		p.nextToken()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TokenRParen {
			return nil, fmt.Errorf("expected ')', got %q", p.cur.Value)
		}
		p.nextToken()
		return &ParenExpr{Expr: expr}, nil

	case TokenIdent:
		name := p.cur.Value
		p.nextToken()
		return &ColRef{Name: name}, nil

	case TokenNumber:
		val, err := strconv.ParseFloat(p.cur.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", p.cur.Value)
		}
		raw := p.cur.Value
		p.nextToken()
		return &NumberLit{Value: val, Raw: raw}, nil

	case TokenString:
		val := p.cur.Value
		p.nextToken()
		return &StringLit{Value: val}, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression: %q at pos %d", p.cur.Value, p.cur.Pos)
	}
}

// ParseQuery parses a GoogleSQL query string into a SelectStmt.
func ParseQuery(sqlQuery string) (*SelectStmt, error) {
	p := NewParser(strings.TrimSpace(sqlQuery))
	stmt, err := p.Parse()
	if err != nil {
		return nil, err
	}
	sel, ok := stmt.(*SelectStmt)
	if !ok {
		return nil, fmt.Errorf("expected SELECT query")
	}
	return sel, nil
}

