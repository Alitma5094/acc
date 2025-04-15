package acc

import (
	"fmt"
)

type tokenType int

const (
	// Single Charicter Tokens
	tokenOpenParen tokenType = iota
	tokenCloseParen
	tokenOpenBrace
	tokenCloseBrace
	tokenSemicolon

	// Literals
	tokenIdentfier
	tokenConstant

	// Keywords
	tokenInt
	tokenVoid
	tokenReturn

	// Operator
	tokenBitwiseCompOp
	tokenNegationOp
	tokenDecrementOp

	tokenAdditionOp
	tokenMultiplicationOp
	tokenDivisionOp
	tokenRemainderOp
)

var keywords = map[string]tokenType{
	"int":    tokenInt,
	"void":   tokenVoid,
	"return": tokenReturn,
}

type token struct {
	tokenType tokenType
	literal   string
	line      int
}

type Lexer struct {
	Tokens  []token
	source  string
	start   int
	current int
	line    int
}

func NewLexer(source string) *Lexer {
	l := &Lexer{
		source: source,
	}
	return l
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) Lex() error {
	for {
		if l.isAtEnd() {
			return nil
		}

		l.start = l.current
		err := l.scanToken()
		if err != nil {
			return err
		}
	}
}

func (l *Lexer) scanToken() error {
	c := l.advance()
	switch c {
	case '(':
		l.addToken(tokenOpenParen, "(")
	case ')':
		l.addToken(tokenCloseParen, ")")
	case '{':
		l.addToken(tokenOpenBrace, "{")
	case '}':
		l.addToken(tokenCloseBrace, "}")
	case ';':
		l.addToken(tokenSemicolon, ";")
	case '~':
		l.addToken(tokenBitwiseCompOp, "~")
	case '-':
		if l.peek() == '-' {
			l.addToken(tokenDecrementOp, "--")
		} else {
			l.addToken(tokenNegationOp, "-")
		}
	case '+':
		l.addToken(tokenAdditionOp, "+")
	case '*':
		l.addToken(tokenMultiplicationOp, "*")
	case '/':
		l.addToken(tokenDivisionOp, "/")
	case '%':
		l.addToken(tokenRemainderOp, "%")

	// Ignore whitespace
	case ' ', '\t', '\r':
		return nil

	case '\n':
		l.line += 1
	default:
		if isDecimal(c) {
			if err := l.number(); err != nil {
				return err
			}
		} else if isAlpha(c) {
			l.identifier()
		} else {
			return fmt.Errorf("invalid char: %s", string(c))
		}
	}
	return nil
}

func (l *Lexer) number() error {
	for {
		if isDecimal(l.peek()) {
			l.advance()
		} else {
			break
		}
	}

	if !l.isAtEnd() && isAlpha(l.peek()) {
		return fmt.Errorf("invalid number at line %d", l.line)
	}
	l.addToken(tokenConstant, l.source[l.start:l.current])
	return nil
}

func (l *Lexer) identifier() {
	for {
		if isAlphaNumeric(l.peek()) {
			l.advance()
		} else {
			break
		}
	}

	text := l.source[l.start:l.current]
	keyword, ok := keywords[text]
	if !ok {
		l.addToken(tokenIdentfier, l.source[l.start:l.current])
	} else {
		l.addToken(keyword, l.source[l.start:l.current])
	}
}

func (l *Lexer) addToken(tokenType tokenType, literal string) {
	l.Tokens = append(l.Tokens, token{tokenType: tokenType, literal: literal, line: l.line})
}

func (l *Lexer) advance() byte {
	c := l.source[l.current]
	l.current += 1
	return c
}

func (l *Lexer) peek() byte {
	return l.source[l.current]
}

func isDecimal(c byte) bool {
	return '0' <= c && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDecimal(c)
}
