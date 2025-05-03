package parser

import (
	"acc/internal/common/errors"
	"acc/internal/lexer"
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
		return nil, errors.NewParseError("invalid chars outside of function", p.tokens[p.index].Loc)
	}

	return program, nil
}

func (p *Parser) parseFunction() (*Function, error) {
	if exists, tok := p.expect(lexer.TokenInt); !exists {
		return nil, errors.NewParseError("missing int", tok.Loc)
	}

	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	if exists, tok := p.expect(lexer.TokenOpenParen); !exists {
		return nil, errors.NewParseError("missing (", tok.Loc)
	}
	if exists, tok := p.expect(lexer.TokenVoid); !exists {
		return nil, errors.NewParseError("missing void", tok.Loc)
	}
	if exists, tok := p.expect(lexer.TokenCloseParen); !exists {
		return nil, errors.NewParseError("missing )", tok.Loc)
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &Function{
		Name: ident,
		Body: body,
	}, nil
}

func (p *Parser) parseBlock() (Block, error) {
	if exists, tok := p.expect(lexer.TokenOpenBrace); !exists {
		return Block{}, errors.NewParseError("missing {", tok.Loc)
	}

	body := []BlockItem{}

	for p.peek().Type != lexer.TokenCloseBrace {
		item, err := p.parseBlockItem()
		if err != nil {
			return Block{}, err
		}
		body = append(body, item)
	}

	if exists, tok := p.expect(lexer.TokenCloseBrace); !exists {
		return Block{}, errors.NewParseError("missing }", tok.Loc)
	}

	return Block{Body: body}, nil
}

func (p *Parser) parseBlockItem() (BlockItem, error) {
	if p.peek().Type == lexer.TokenInt {
		decl, err := p.parseDeclaration()
		if err != nil {
			return nil, err
		}
		return &DeclarationBlock{Declaration: *decl}, err

	} else {
		// Statement
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return &StmtBlock{Statement: stmt}, nil
	}
}

func (p *Parser) parseDeclaration() (*Declaration, error) {
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

	if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
		return nil, errors.NewParseError("missing semicolon", tok.Loc)
	}

	return &Declaration{Name: ident, Init: expression}, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	nextToken := p.peek()
	switch nextToken.Type {
	case lexer.TokenSemicolon:
		p.expect(lexer.TokenSemicolon)
		return &NullStmt{Loc: nextToken.Loc}, nil
	case lexer.TokenReturn:
		p.expect(lexer.TokenReturn)

		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}
		return &ReturnStmt{Loc: nextToken.Loc, Expression: expr}, nil
	case lexer.TokenIf:
		p.expect(lexer.TokenIf)

		if exists, tok := p.expect(lexer.TokenOpenParen); !exists {
			return nil, errors.NewParseError("missing opening parenthesis", tok.Loc)
		}

		condition, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenCloseParen); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}

		then, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		if p.peek().Type == lexer.TokenElse {
			p.expect(lexer.TokenElse)
			elseStmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			return &IfStmt{Loc: nextToken.Loc, Condition: condition, Then: then, Else: elseStmt}, nil
		}
		return &IfStmt{Loc: nextToken.Loc, Condition: condition, Then: then}, nil
	case lexer.TokenOpenBrace:
		block, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &CompoundStmt{Block: block}, nil
	case lexer.TokenBreak:
		p.expect(lexer.TokenBreak)

		if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}
		return &BreakStmt{}, nil
	case lexer.TokenContinue:
		p.expect(lexer.TokenBreak)

		if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}
		return &ContinueStmt{}, nil
	case lexer.TokenWhile:
		p.expect(lexer.TokenWhile)

		if exists, tok := p.expect(lexer.TokenOpenParen); !exists {
			return nil, errors.NewParseError("missing open parenthesis", tok.Loc)
		}

		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenCloseParen); !exists {
			return nil, errors.NewParseError("missing close parenthesis", tok.Loc)
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		return &WhileStmt{Condition: exp, Body: stmt}, nil
	case lexer.TokenDo:
		p.expect(lexer.TokenDo)

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenWhile); !exists {
			return nil, errors.NewParseError("missing while", tok.Loc)
		}
		if exists, tok := p.expect(lexer.TokenOpenParen); !exists {
			return nil, errors.NewParseError("missing open parenthesis", tok.Loc)
		}

		exp, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenCloseParen); !exists {
			return nil, errors.NewParseError("missing close parenthesis", tok.Loc)
		}

		if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}

		return &DoWhileStmt{Body: stmt, Condition: exp}, nil
	case lexer.TokenFor:
		p.expect(lexer.TokenFor)
		if exists, tok := p.expect(lexer.TokenOpenParen); !exists {
			return nil, errors.NewParseError("missing open parenthesis", tok.Loc)
		}
		init, err := p.parseForInit()
		if err != nil {
			return nil, err
		}

		condition, err := p.parseOptionalExpression(lexer.TokenSemicolon)
		if err != nil {
			return nil, err
		}

		post, err := p.parseOptionalExpression(lexer.TokenCloseParen)
		if err != nil {
			return nil, err
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		return &ForStmt{Init: init, Condition: condition, Post: post, Body: stmt}, nil
	default:
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		if exists, tok := p.expect(lexer.TokenSemicolon); !exists {
			return nil, errors.NewParseError("missing semicolon", tok.Loc)
		}

		return &ExpressionStmt{Loc: nextToken.Loc, Expression: expr}, nil
	}
}

func (p *Parser) parseOptionalExpression(terminatingToken lexer.TokenType) (Expression, error) {
	if tok := p.peek(); tok.Type == terminatingToken {
		p.expect(terminatingToken)
		return nil, nil
	}

	exp, err := p.parseExpression(0)
	if exists, _ := p.expect(terminatingToken); !exists {
		return nil, err
	}
	return exp, err
}

func (p *Parser) parseForInit() (ForInit, error) {
	if p.peek().Type == lexer.TokenSemicolon {
		p.expect(lexer.TokenSemicolon)
		return nil, nil
	} else if p.peek().Type == lexer.TokenInt {
		decl, err := p.parseDeclaration()
		if err != nil {
			return nil, err
		}
		return &InitDecl{Declaration: *decl}, nil
	}

	exp, err := p.parseOptionalExpression(lexer.TokenSemicolon)
	if err != nil {
		return nil, err
	}
	return &InitExp{Expression: exp}, err
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

			leftExpr = &AssignmentExp{Loc: nextToken.Loc, Left: leftExpr, Right: rightExpr}

		} else if nextToken.Type == lexer.TokenConditionalOpFront {
			middle, err := p.parseConditionalMiddle()
			if err != nil {
				return nil, err
			}
			rightExpr, err := p.parseExpression(precedence)
			if err != nil {
				return nil, err
			}
			leftExpr = &ConditionalExp{Loc: nextToken.Loc, Condition: leftExpr, Expression1: middle, Expression2: rightExpr}
		} else {
			op, err := p.parseBinaryOp()
			if err != nil {
				return nil, err
			}

			rightExpr, err := p.parseExpression(precedence + 1)
			if err != nil {
				return nil, err
			}

			leftExpr = &BinaryExp{Loc: nextToken.Loc, Left: leftExpr, Right: rightExpr, Op: op}
		}
	}
	return leftExpr, nil
}

func (p *Parser) parseConditionalMiddle() (Expression, error) {
	p.expect(lexer.TokenConditionalOpFront)
	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if exists, tok := p.expect(lexer.TokenConditionalOpEnd); !exists {
		return nil, errors.NewParseError("missing end of conditional", tok.Loc)
	}
	return expr, nil
}

func (p *Parser) parseFactor() (Factor, error) {
	nextTok := p.peek()
	switch nextTok.Type {
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
		if exists, tok := p.expect(lexer.TokenCloseParen); !exists {
			return nil, errors.NewParseError("missing )", tok.Loc)
		}
		return &NestedExp{Loc: nextTok.Loc, Expr: expr}, nil

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

	nextTok := p.peek()
	switch nextTok.Type {
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
		return nil, errors.NewParseError("expected unary operator", nextTok.Loc)
	}

	exp, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	return &UnaryFactor{Loc: nextTok.Loc, Op: opType, Value: exp}, nil
}

func (p *Parser) parseBinaryOp() (BinopType, error) {
	nextTok := p.peek()
	switch nextTok.Type {
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
		return -1, errors.NewParseError("expected binary operator", nextTok.Loc)
	}
}

func (p *Parser) parseIdentifier() (IdentifierFactor, error) {
	exists, tok := p.expect(lexer.TokenIdentifier)
	if !exists {
		return IdentifierFactor{}, errors.NewParseError("missing identifier", tok.Loc)
	}
	return IdentifierFactor{Loc: tok.Loc, Value: tok.Literal}, nil
}

func (p *Parser) parseInt() (*IntLiteral, error) {
	exists, tok := p.expect(lexer.TokenConstant)
	if !exists {
		return nil, errors.NewParseError("missing int constant", tok.Loc)
	}
	val, err := strconv.ParseInt(tok.Literal, 10, 0)
	if err != nil {
		return nil, err
	}
	return &IntLiteral{Loc: tok.Loc, Value: int(val)}, nil
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
	case lexer.TokenConditionalOpFront:
		return 3
	case lexer.TokenAssignmentOp:
		return 1
	default:
		return -1
	}
}
