package sql

// TokenType represents lexical token types.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdent
	TokenNumber
	TokenString
	TokenKeyword

	// Operators and punctuation
	TokenEq    // =
	TokenNeq   // != or <>
	TokenLt    // <
	TokenLte   // <=
	TokenGt    // >
	TokenGte   // >=
	TokenComma // ,
	TokenLParen
	TokenRParen
	TokenStar // *
)

// Token represents a single lexical token.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

