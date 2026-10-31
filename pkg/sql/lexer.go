package sql

import (
	"fmt"
	"strings"
	"unicode"
)

var keywords = map[string]bool{
	"SELECT":  true,
	"FROM":    true,
	"WHERE":   true,
	"AND":     true,
	"OR":      true,
	"BETWEEN": true,
	"LIMIT":   true,
	"IN":      true,
	"NOT":     true,
	"AS":      true,
	"ORDER":   true,
	"BY":      true,
	"ASC":     true,
	"DESC":    true,
}

// Lexer tokenizes SQL queries.
type Lexer struct {
	input []rune
	pos   int
}

// NewLexer creates a new Lexer instance.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Value: "", Pos: l.pos}
	}

	start := l.pos
	ch := l.input[l.pos]

	switch ch {
	case '*':
		l.pos++
		return Token{Type: TokenStar, Value: "*", Pos: start}
	case ',':
		l.pos++
		return Token{Type: TokenComma, Value: ",", Pos: start}
	case '(':
		l.pos++
		return Token{Type: TokenLParen, Value: "(", Pos: start}
	case ')':
		l.pos++
		return Token{Type: TokenRParen, Value: ")", Pos: start}
	case '=':
		l.pos++
		return Token{Type: TokenEq, Value: "=", Pos: start}
	case '!':
		if l.peek() == '=' {
			l.pos += 2
			return Token{Type: TokenNeq, Value: "!=", Pos: start}
		}
		l.pos++
		return Token{Type: TokenError, Value: "unexpected '!'", Pos: start}
	case '<':
		if l.peek() == '=' {
			l.pos += 2
			return Token{Type: TokenLte, Value: "<=", Pos: start}
		}
		if l.peek() == '>' {
			l.pos += 2
			return Token{Type: TokenNeq, Value: "<>", Pos: start}
		}
		l.pos++
		return Token{Type: TokenLt, Value: "<", Pos: start}
	case '>':
		if l.peek() == '=' {
			l.pos += 2
			return Token{Type: TokenGte, Value: ">=", Pos: start}
		}
		l.pos++
		return Token{Type: TokenGt, Value: ">", Pos: start}
	case '\'', '"':
		return l.scanString(ch, start)
	case '`': // GoogleSQL backtick quoted identifier
		return l.scanBacktickIdent(start)
	}

	if ch == '-' || ch == '+' || unicode.IsDigit(ch) {
		// Could be a signed number or an operator
		if (ch == '-' || ch == '+') && !unicode.IsDigit(l.peek()) && l.peek() != '.' {
			l.pos++
			return Token{Type: TokenError, Value: fmt.Sprintf("unexpected %c", ch), Pos: start}
		}
		return l.scanNumber(start)
	}

	if unicode.IsLetter(ch) || ch == '_' {
		return l.scanIdentOrKeyword(start)
	}

	l.pos++
	return Token{Type: TokenError, Value: fmt.Sprintf("unrecognized character %c", ch), Pos: start}
}

func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) scanString(quote rune, start int) Token {
	l.pos++ // skip opening quote
	valStart := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	val := string(l.input[valStart:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == quote {
		l.pos++ // skip closing quote
	}
	return Token{Type: TokenString, Value: val, Pos: start}
}

func (l *Lexer) scanBacktickIdent(start int) Token {
	l.pos++ // skip opening backtick
	valStart := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '`' {
		l.pos++
	}
	val := string(l.input[valStart:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == '`' {
		l.pos++ // skip closing backtick
	}
	return Token{Type: TokenIdent, Value: val, Pos: start}
}

func (l *Lexer) scanNumber(start int) Token {
	if l.input[l.pos] == '-' || l.input[l.pos] == '+' {
		l.pos++
	}
	hasDot := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsDigit(ch) {
			l.pos++
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.pos++
		} else {
			break
		}
	}
	return Token{Type: TokenNumber, Value: string(l.input[start:l.pos]), Pos: start}
}

func (l *Lexer) scanIdentOrKeyword(start int) Token {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' {
			l.pos++
		} else {
			break
		}
	}
	val := string(l.input[start:l.pos])
	upper := strings.ToUpper(val)
	if keywords[upper] {
		return Token{Type: TokenKeyword, Value: upper, Pos: start}
	}
	return Token{Type: TokenIdent, Value: val, Pos: start}
}

