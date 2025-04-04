package acc

import (
	"errors"
	"strconv"
)

type Parser struct {
	tokens []token
	index  int
	Tree   node
}

func NewParser(tokens []token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() error {
	n := &nodeProgram{}
	err := n.parse(p)
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

type node interface {
	parse(parser *Parser) error
}

type nodeProgram struct {
	function *nodeFunction
}

func (f *nodeProgram) parse(parser *Parser) error {
	function := &nodeFunction{}
	err := function.parse(parser)
	if err != nil {
		return err
	}

	f.function = function

	return nil
}

type nodeFunction struct {
	name *nodeIdentifier
	body *nodeStatement
}

func (f *nodeFunction) parse(parser *Parser) error {
	if exists, _ := parser.expect(tokenInt); !exists {
		return errors.New("missing int")
	}

	iden := &nodeIdentifier{}
	err := iden.parse(parser)
	if err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenOpenParen); !exists {
		return errors.New("missing opening parenthisis")
	}
	if exists, _ := parser.expect(tokenVoid); !exists {
		return errors.New("missing void")
	}
	if exists, _ := parser.expect(tokenCloseParen); !exists {
		return errors.New("missing closing parenthisis")
	}
	if exists, _ := parser.expect(tokenOpenBrace); !exists {
		return errors.New("missing opening brace")
	}

	stmt := &nodeStatement{}
	err = stmt.parse(parser)
	if err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenCloseBrace); !exists {
		return errors.New("missing closing brace")
	}

	f.name = iden
	f.body = stmt

	return nil
}

type nodeStatement struct {
	expression *nodeExpression
}

func (s *nodeStatement) parse(parser *Parser) error {
	if exists, _ := parser.expect(tokenReturn); !exists {
		return errors.New("missing return")
	}

	return_val := &nodeExpression{}
	err := return_val.parse(parser)
	if err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenSemicolon); !exists {
		return errors.New("missing return")
	}

	return nil
}

type nodeExpression struct {
	val *nodeInt
}

func (s *nodeExpression) parse(parser *Parser) error {
	expr := &nodeInt{}
	err := expr.parse(parser)
	if err != nil {
		return err
	}

	s.val = expr

	return nil
}

type nodeIdentifier struct {
	val string
}

func (s *nodeIdentifier) parse(parser *Parser) error {
	exists, tok := parser.expect(tokenIdentfier)
	if !exists {
		return errors.New("missing identifier")
	}

	s.val = tok.literal

	return nil
}

type nodeInt struct {
	val int
}

func (s *nodeInt) parse(parser *Parser) error {
	exists, tok := parser.expect(tokenConstant)
	if !exists {
		return errors.New("missing int")
	}

	pint, _ := strconv.ParseInt(tok.literal, 10, 0)

	s.val = int(pint)

	return nil
}
