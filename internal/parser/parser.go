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

	body := []BlockItem{}

	for p.peek().Type != lexer.TokenCloseBrace {
		item, err := p.parseBlockItem()
		if err != nil {
			return nil, err
		}
		body = append(body, item)
	}

	if exists, _ := p.expect(lexer.TokenCloseBrace); !exists {
		return nil, errors.New("missing }")
	}

	return &Function{
		Name: ident,
		Body: body,
	}, nil
}

func (p *Parser) parseBlockItem() (BlockItem, error) {
	if p.peek().Type == lexer.TokenInt {
		// Declaration
		p.expect(lexer.TokenInt)

		ident, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		var expression Expression
		if p.peek().Type == lexer.TokenAssignmentOp {
			p.expect(lexer.TokenAssignmentOp)
			expression, err = p.parseExpression(0)
			if err != nil {
				return nil, err
			}
		}

		if exists, _ := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.New("missing semicolon")
		}

		return &DeclarationBlock{Declaration: Declaration{Name: ident, Init: expression}}, nil

	} else {
		// Statement
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return &StmtBlock{Statement: stmt}, nil
	}
}

func (p *Parser) parseStatement() (Statement, error) {
	if p.peek().Type == lexer.TokenSemicolon {
		p.expect(lexer.TokenSemicolon)
		return &NullStmt{}, nil
	} else if p.peek().Type == lexer.TokenReturn {
		p.expect(lexer.TokenReturn)

		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, _ := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.New("missing semicolon")
		}
		return &ReturnStmt{Expression: expr}, nil
	} else {
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, _ := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.New("missing semicolon")
		}

		return &ExpressionStmt{Expression: expr}, nil
	}
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

		if nextToken.Type == lexer.TokenAssignmentOp {
			p.expect(lexer.TokenAssignmentOp)
			rightExpr, err := p.parseExpression(precedence)
			if err != nil {
				return nil, err
			}

			leftExpr = &AssignmentExp{Left: leftExpr, Right: rightExpr}

		} else {
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

	case lexer.TokenNegationOp, lexer.TokenBitwiseCompOp, lexer.TokenNotOp:
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
		ident, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		return &ident, nil
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
	case lexer.TokenNotOp:
		p.expect(lexer.TokenNotOp)
		opType = UnopNot
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
	case lexer.TokenAndOp:
		p.expect(lexer.TokenAndOp)
		return BinopAnd, nil
	case lexer.TokenOrOp:
		p.expect(lexer.TokenOrOp)
		return BinopOr, nil
	case lexer.TokenEqualOp:
		p.expect(lexer.TokenEqualOp)
		return BinopEqual, nil
	case lexer.TokenNotEqualOp:
		p.expect(lexer.TokenNotEqualOp)
		return BinopNotEqual, nil
	case lexer.TokenLessThanOp:
		p.expect(lexer.TokenLessThanOp)
		return BinopLessThan, nil
	case lexer.TokenLessOrEqualOp:
		p.expect(lexer.TokenLessOrEqualOp)
		return BinopLessOrEqual, nil
	case lexer.TokenGreaterThanOp:
		p.expect(lexer.TokenGreaterThanOp)
		return BinopGreaterThan, nil
	case lexer.TokenGreaterOrEqualOp:
		p.expect(lexer.TokenGreaterOrEqualOp)
		return BinopGreaterOrEqual, nil

	default:
		return -1, errors.New("expected binary operator")
	}
}

func (p *Parser) parseIdentifier() (IdentifierFactor, error) {
	exists, tok := p.expect(lexer.TokenIdentifier)
	if !exists {
		return IdentifierFactor{}, errors.New("missing identifier")
	}
	return IdentifierFactor{Value: tok.Literal}, nil
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
	case lexer.TokenMultiplicationOp, lexer.TokenDivisionOp, lexer.TokenRemainderOp:
		return 50
	case lexer.TokenAdditionOp, lexer.TokenNegationOp:
		return 45
	case lexer.TokenLessThanOp, lexer.TokenLessOrEqualOp, lexer.TokenGreaterThanOp, lexer.TokenGreaterOrEqualOp:
		return 35
	case lexer.TokenEqualOp, lexer.TokenNotEqualOp:
		return 30
	case lexer.TokenAndOp:
		return 10
	case lexer.TokenOrOp:
		return 5
	case lexer.TokenAssignmentOp:
		return 1
	default:
		return -1
	}
}
