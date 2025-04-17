package parser

import (
	"acc/internal/lexer"
	"errors"
	"strconv"
)

type Parser struct {
	tokens []lexer.Token
	index  int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) isAtEnd() bool {
	return p.index >= len(p.tokens)
}

func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return lexer.Token{}
	}
	return p.tokens[p.index]
}

func (p *Parser) expect(expected lexer.TokenType) (bool, lexer.Token) {
	if p.isAtEnd() {
		return false, lexer.Token{}
	}
	nextToken := p.tokens[p.index]
	p.index++
	if nextToken.Type == expected {
		return true, nextToken
	}
	return false, nextToken
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}
	function, err := p.parseFunction()
	if err != nil {
		return nil, err
	}
	program.Function = function

	if !p.isAtEnd() {
		return nil, errors.New("invalid chars outside of function")
	}

	return program, nil
}

func (p *Parser) parseFunction() (*Function, error) {
	if exists, _ := p.expect(lexer.TokenInt); !exists {
		return nil, errors.New("missing int")
	}

	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	if exists, _ := p.expect(lexer.TokenOpenParen); !exists {
		return nil, errors.New("missing (")
	}
	if exists, _ := p.expect(lexer.TokenVoid); !exists {
		return nil, errors.New("missing void")
	}
	if exists, _ := p.expect(lexer.TokenCloseParen); !exists {
		return nil, errors.New("missing )")
	}
	if exists, _ := p.expect(lexer.TokenOpenBrace); !exists {
		return nil, errors.New("missing {")
	}

	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	if exists, _ := p.expect(lexer.TokenCloseBrace); !exists {
		return nil, errors.New("missing }")
	}

	return &Function{
		Name: ident,
		Body: stmt,
	}, nil
}

func (p *Parser) parseStatement() (*Statement, error) {
	if exists, _ := p.expect(lexer.TokenReturn); !exists {
		return nil, errors.New("missing return")
	}

	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if exists, _ := p.expect(lexer.TokenSemicolon); !exists {
		return nil, errors.New("missing semicolon")
	}

	return &Statement{Expression: &expr}, nil
}

func (p *Parser) parseExpression(minPrecedence int) (Expression, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	var leftExpr Expression = &FactorExp{Factor: left}

	for {
		nextToken := p.peek()
		if binopPrecedence(nextToken) < minPrecedence {
			break
		}
		precedence := binopPrecedence(nextToken)

		op, err := p.parseBinaryOp()
		if err != nil {
			return nil, err
		}

		rightExpr, err := p.parseExpression(precedence + 1)
		if err != nil {
			return nil, err
		}

		leftExpr = &BinaryExp{Left: leftExpr, Right: rightExpr, Op: op}
	}
	return leftExpr, nil
}

func (p *Parser) parseFactor() (Factor, error) {
	switch p.peek().Type {
	case lexer.TokenConstant:
		intNode, err := p.parseInt()
		if err != nil {
			return nil, err
		}
		return intNode, nil

	case lexer.TokenNegationOp, lexer.TokenBitwiseCompOp:
		unopNode, err := p.parseUnaryOp()
		if err != nil {
			return nil, err
		}
		return unopNode, nil

	case lexer.TokenOpenParen:
		p.expect(lexer.TokenOpenParen)
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if exists, _ := p.expect(lexer.TokenCloseParen); !exists {
			return nil, errors.New("missing )")
		}
		return &NestedExp{Expr: expr}, nil

	default:
		return nil, errors.New("malformed factor")
	}
}

func (p *Parser) parseUnaryOp() (*UnaryFactor, error) {
	var opType UnopType

	switch p.peek().Type {
	case lexer.TokenBitwiseCompOp:
		p.expect(lexer.TokenBitwiseCompOp)
		opType = UnopBitwiseComp
	case lexer.TokenNegationOp:
		p.expect(lexer.TokenNegationOp)
		opType = UnopNegate
	default:
		return nil, errors.New("expected unary operator")
	}

	exp, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	return &UnaryFactor{Op: opType, Value: exp}, nil
}

func (p *Parser) parseBinaryOp() (BinopType, error) {
	switch p.peek().Type {
	case lexer.TokenAdditionOp:
		p.expect(lexer.TokenAdditionOp)
		return BinopAdd, nil
	case lexer.TokenNegationOp:
		p.expect(lexer.TokenNegationOp)
		return BinopSubtract, nil
	case lexer.TokenMultiplicationOp:
		p.expect(lexer.TokenMultiplicationOp)
		return BinopMultiply, nil
	case lexer.TokenDivisionOp:
		p.expect(lexer.TokenDivisionOp)
		return BinopDivide, nil
	case lexer.TokenRemainderOp:
		p.expect(lexer.TokenRemainderOp)
		return BinopRemainder, nil
	default:
		return 0, errors.New("expected binary operator")
	}
}

func (p *Parser) parseIdentifier() (string, error) {
	exists, tok := p.expect(lexer.TokenIdentifier)
	if !exists {
		return "", errors.New("missing identifier")
	}
	return tok.Literal, nil
}

func (p *Parser) parseInt() (*IntLiteral, error) {
	exists, tok := p.expect(lexer.TokenConstant)
	if !exists {
		return nil, errors.New("missing int constant")
	}
	val, err := strconv.ParseInt(tok.Literal, 10, 0)
	if err != nil {
		return nil, err
	}
	return &IntLiteral{Value: int(val)}, nil
}

func binopPrecedence(tok lexer.Token) int {
	switch tok.Type {
	case lexer.TokenAdditionOp, lexer.TokenNegationOp:
		return 45
	case lexer.TokenMultiplicationOp, lexer.TokenDivisionOp, lexer.TokenRemainderOp:
		return 50
	default:
		return -1
	}
}
