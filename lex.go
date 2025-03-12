package main

import (
	"fmt"
	"log"
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

	// misc
	tokenEOF
)

var keywords = map[string]tokenType{
	"int":    tokenInt,
	"void":   tokenVoid,
	"return": tokenReturn,
}

type token struct {
	tokenType tokenType
	literal   any
	line      int
}

type Lexer struct {
	tokens  []token
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

func (l *Lexer) lex() error {
	for {
		if l.isAtEnd() {
			return nil
		}

		l.start = l.current
		l.scanToken()
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

	// Ignore whitespace
	case ' ', '\t', '\r':
		return nil

	case '\n':
		l.line += 1
	default:
		if isDecimal(c) {
			l.number()
		} else if isAlpha(c) {
			l.identifier()
		} else {
			return fmt.Errorf("invalid char: %s", string(c))
		}
	}
	return nil
}

func (l *Lexer) number() {
	for {
		if isDecimal(l.peek()) {
			l.advance()
		} else {
			break
		}
	}

	if !l.isAtEnd() && isAlpha(l.peek()) {
		log.Fatalf("Invalid number at line %d", l.line)
	}
	l.addToken(tokenConstant, l.source[l.start:l.current])
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
		l.addToken(tokenConstant, l.source[l.start:l.current])
	} else {
		l.addToken(keyword, l.source[l.start:l.current])
	}
}

func (l *Lexer) addToken(tokenType tokenType, literal any) {
	l.tokens = append(l.tokens, token{tokenType: tokenType, literal: literal, line: l.line})
}

func (l *Lexer) advance() byte {
	// TODO: check if is at end like peek()
	c := l.source[l.current]
	l.current += 1
	return c
}

func (l *Lexer) peek() byte {
	// TODO: find some other way to indicate end of file
	//if l.isAtEnd() {
	//  return '\0'
	//}
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
