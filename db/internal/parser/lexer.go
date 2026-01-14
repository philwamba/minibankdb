package parser

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenIdentifier
	TokenKeyword
	TokenString
	TokenNumber
	TokenSymbol
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input string
	pos   int
	len   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		len:   len(input),
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	if l.pos >= l.len {
		return Token{Type: TokenEOF}
	}

	ch := l.input[l.pos]

	if isAlpha(ch) {
		start := l.pos
		for l.pos < l.len && (isAlphaNum(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}
		val := l.input[start:l.pos]
		if isKeyword(val) {
			return Token{Type: TokenKeyword, Value: strings.ToUpper(val)}
		}
		return Token{Type: TokenIdentifier, Value: val}
	}

	if isDigit(ch) {
		start := l.pos
		for l.pos < l.len && (isDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
			l.pos++
		}
		return Token{Type: TokenNumber, Value: l.input[start:l.pos]}
	}

	if ch == '\'' {
		l.pos++
		start := l.pos
		for l.pos < l.len && l.input[l.pos] != '\'' {
			l.pos++
		}
		val := l.input[start:l.pos]
		if l.pos < l.len {
			l.pos++
		}
		return Token{Type: TokenString, Value: val}
	}

	l.pos++
	switch ch {
	case '=', '*', '(', ')', ',', ';':
		return Token{Type: TokenSymbol, Value: string(ch)}
	case '<', '>', '!':
		if l.pos < l.len && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenSymbol, Value: string(ch) + "="}
		}
		return Token{Type: TokenSymbol, Value: string(ch)}
	}

	return Token{Type: TokenError, Value: string(ch)}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < l.len && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlphaNum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func isKeyword(val string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true,
		"INTO": true, "VALUES": true, "UPDATE": true, "SET": true,
		"DELETE": true, "CREATE": true, "TABLE": true, "INDEX": true,
		"ON": true, "JOIN": true, "AND": true, "OR": true,
		"INT": true, "STRING": true, "DECIMAL": true, "BOOL": true, "TIMESTAMP": true,
		"PRIMARY": true, "KEY": true, "UNIQUE": true,
	}
	return keywords[strings.ToUpper(val)]
}
