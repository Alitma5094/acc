package lexer

import "acc/internal/common/errors"

type TokenType int

const (
	// Single Character Tokens
	TokenOpenParen TokenType = iota
	TokenCloseParen
	TokenOpenBrace
	TokenCloseBrace
	TokenSemicolon

	// Literals
	TokenIdentifier
	TokenConstant

	// Keywords
	TokenInt
	TokenVoid
	TokenReturn

	// Operators
	TokenBitwiseCompOp
	TokenNegationOp
	TokenDecrementOp
	TokenAdditionOp
	TokenMultiplicationOp
	TokenDivisionOp
	TokenRemainderOp
)

type Token struct {
	Type    TokenType
	Literal string
	Loc     errors.Location
}

func NewToken(tokenType TokenType, literal string, loc errors.Location) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Loc:     loc,
	}
}

var Keywords = map[string]TokenType{
	"int":    TokenInt,
	"void":   TokenVoid,
	"return": TokenReturn,
}
