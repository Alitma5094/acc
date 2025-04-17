package lexer

import (
	"acc/internal/common/errors"
)

type Lexer struct {
	source  string
	tokens  []Token
	start   int
	current int
	line    int
	column  int
	file    string
}

func NewLexer(source string) *Lexer {
	return &Lexer{
		source: source,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) SetFile(file string) {
	l.file = file
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for !l.isAtEnd() {
		l.start = l.current
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}

	return l.tokens, nil
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) scanToken() error {
	c := l.advance()

	switch c {
	case '(':
		l.addToken(TokenOpenParen, "(")
	case ')':
		l.addToken(TokenCloseParen, ")")
	case '{':
		l.addToken(TokenOpenBrace, "{")
	case '}':
		l.addToken(TokenCloseBrace, "}")
	case ';':
		l.addToken(TokenSemicolon, ";")
	case '~':
		l.addToken(TokenBitwiseCompOp, "~")
	case '-':
		if l.match('-') {
			l.addToken(TokenDecrementOp, "--")
		} else {
			l.addToken(TokenNegationOp, "-")
		}
	case '+':
		l.addToken(TokenAdditionOp, "+")
	case '*':
		l.addToken(TokenMultiplicationOp, "*")
	case '/':
		l.addToken(TokenDivisionOp, "/")
	case '%':
		l.addToken(TokenRemainderOp, "%")

	// Ignore whitespace
	case ' ', '\t', '\r':
		// Do nothing
	case '\n':
		l.line++
		l.column = 0 // Reset column at new line

	default:
		if isDigit(c) {
			return l.number()
		} else if isAlpha(c) {
			l.identifier()
		} else {
			return errors.NewLexError(
				"Unexpected character: "+string(c),
				l.currentLocation(),
			)
		}
	}

	return nil
}

func (l *Lexer) number() error {
	startLoc := l.currentLocation()

	for !l.isAtEnd() && isDigit(l.peek()) {
		l.advance()
	}

	// Check for invalid identifiers immediately after number
	if !l.isAtEnd() && isAlpha(l.peek()) {
		return errors.NewLexError("Invalid number", startLoc)
	}

	l.addToken(TokenConstant, l.source[l.start:l.current])
	return nil
}

func (l *Lexer) identifier() {
	for !l.isAtEnd() && isAlphaNumeric(l.peek()) {
		l.advance()
	}

	text := l.source[l.start:l.current]

	// Check if the identifier is a keyword
	if tokenType, isKeyword := Keywords[text]; isKeyword {
		l.addToken(tokenType, "")
	} else {
		l.addToken(TokenIdentifier, text)
	}
}

func (l *Lexer) addToken(tokenType TokenType, literal string) {
	lexeme := l.source[l.start:l.current]
	l.tokens = append(l.tokens, NewToken(
		tokenType,
		literal,
		errors.NewLocation(l.line, l.column-len(lexeme), l.file),
	))
}

// func (l *Lexer) addTokenWithLiteral(tokenType TokenType, literal string) {
// 	lexeme := l.source[l.start:l.current]
// 	l.tokens = append(l.tokens, NewToken(
// 		tokenType,
// 		literal,
// 		errors.NewLocation(l.line, l.column-len(lexeme), l.file),
// 	))
// }

func (l *Lexer) advance() byte {
	c := l.source[l.current]
	l.current++
	l.column++
	return c
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.current++
	l.column++
	return true
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

func (l *Lexer) currentLocation() errors.Location {
	return errors.NewLocation(l.line, l.column, l.file)
}

// Utility functions
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}
