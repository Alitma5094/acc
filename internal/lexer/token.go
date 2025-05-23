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

	TokenConditionalOpFront
	TokenConditionalOpEnd

	// Literals
	TokenIdentifier
	TokenConstant

	// Keywords
	TokenInt
	TokenVoid
	TokenReturn
	TokenIf
	TokenElse
	TokenDo
	TokenWhile
	TokenFor
	TokenBreak
	TokenContinue

	// Unary Operators
	TokenBitwiseCompOp
	TokenNegationOp

	// Binary Operators
	TokenDecrementOp
	TokenAdditionOp
	TokenMultiplicationOp
	TokenDivisionOp
	TokenRemainderOp

	// Logical Operators
	TokenNotOp
	TokenAndOp
	TokenOrOp

	// Relational Operators
	TokenEqualOp
	TokenNotEqualOp
	TokenLessThanOp
	TokenGreaterThanOp
	TokenLessOrEqualOp
	TokenGreaterOrEqualOp

	TokenAssignmentOp
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
	"int":      TokenInt,
	"void":     TokenVoid,
	"return":   TokenReturn,
	"if":       TokenIf,
	"else":     TokenElse,
	"do":       TokenDo,
	"while":    TokenWhile,
	"for":      TokenFor,
	"break":    TokenBreak,
	"continue": TokenContinue,
}
