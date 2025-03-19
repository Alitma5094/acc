package acc

import (
	"errors"
	"strconv"
)

type Parser struct {
	tokens []token
	index  int
	Tree   nodeProgram
}

type nodeProgram struct {
	function nodeFunction
}
type nodeFunction struct {
	identifier nodeIdentifier
	statement  nodeStatement
}
type nodeStatement struct {
	expression nodeExpression
}
type nodeExpression struct {
	val nodeInt
}
type nodeIdentifier struct {
	val string
}
type nodeInt struct {
	val int
}

func NewParser(tokens []token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() error {
	n, err := p.parseProgram()
	if err != nil {
		return err
	}
	if !p.isAtEnd() {
		return errors.New("invalid chars outside of function")
	}
	p.Tree = n
	return nil
}

func (p *Parser) isAtEnd() bool {
	if p.index >= len(p.tokens) {
		// Is at end
		return true
	}
	return false
}
func (p *Parser) expect(expected tokenType) (bool, token) {
	if p.isAtEnd() {
		return false, token{}
	}
	next_token := p.tokens[p.index]
	p.index++

	if next_token.tokenType == expected {
		return true, next_token
	}
	return false, next_token

}

func (p *Parser) parseProgram() (nodeProgram, error) {
	function, err := p.parseFunction()
	if err != nil {
		return nodeProgram{}, err
	}

	return nodeProgram{function: function}, nil
}

func (p *Parser) parseFunction() (nodeFunction, error) {

	if exists, _ := p.expect(tokenInt); !exists {
		return nodeFunction{}, errors.New("missing int")
	}

	iden, err := p.parseIdentifier()
	if err != nil {
		return nodeFunction{}, err
	}

	if exists, _ := p.expect(tokenOpenParen); !exists {
		return nodeFunction{}, errors.New("missing opening parenthisis")
	}
	if exists, _ := p.expect(tokenVoid); !exists {
		return nodeFunction{}, errors.New("missing void")
	}
	if exists, _ := p.expect(tokenCloseParen); !exists {
		return nodeFunction{}, errors.New("missing closing parenthisis")
	}
	if exists, _ := p.expect(tokenOpenBrace); !exists {
		return nodeFunction{}, errors.New("missing opening brace")
	}

	stmt, err := p.parseStatement()
	if err != nil {
		return nodeFunction{}, err
	}

	if exists, _ := p.expect(tokenCloseBrace); !exists {
		return nodeFunction{}, errors.New("missing closing brace")
	}

	return nodeFunction{
		identifier: iden,
		statement:  stmt}, nil
}

func (p *Parser) parseStatement() (nodeStatement, error) {
	if exists, _ := p.expect(tokenReturn); !exists {
		return nodeStatement{}, errors.New("missing return")
	}

	return_val, err := p.parseExpression()
	if err != nil {
		return nodeStatement{}, err
	}

	if exists, _ := p.expect(tokenSemicolon); !exists {
		return nodeStatement{}, errors.New("missing return")
	}

	return nodeStatement{
		expression: return_val,
	}, nil
}

func (p *Parser) parseExpression() (nodeExpression, error) {
	expr, err := p.parseInt()
	if err != nil {
		return nodeExpression{}, err
	}

	return nodeExpression{val: expr}, nil
}

func (p *Parser) parseIdentifier() (nodeIdentifier, error) {
	exists, tok := p.expect(tokenIdentfier)
	if !exists {
		return nodeIdentifier{}, errors.New("missing identifier")
	}

	return nodeIdentifier{val: tok.literal}, nil
}

func (p *Parser) parseInt() (nodeInt, error) {
	exists, tok := p.expect(tokenConstant)
	if !exists {
		return nodeInt{}, errors.New("missing int")
	}

	pint, _ := strconv.ParseInt(tok.literal, 10, 0)

	return nodeInt{val: int(pint)}, nil
}
