package acc

import (
	"errors"
	"strconv"
)

type unopType int

const (
	unopBitwiseComp unopType = iota
	unopNegate
)

type nodeExpression interface {
	nodeExpr()
	parse(p *Parser) error
}

type Parser struct {
	tokens []token
	index  int
	Tree   nodeProgram
}

type nodeProgram struct {
	function nodeFunction
}

type nodeFunction struct {
	name nodeIdentifier
	body nodeStatement
}

type nodeStatement struct {
	expression nodeExpression
}

type nodeUnop struct {
	opType unopType
	exp    nodeExpression
}

type nodeIdentifier struct {
	val string
}

type nodeInt struct {
	val int
}

func (i *nodeInt) nodeExpr()  {}
func (o *nodeUnop) nodeExpr() {}

func NewParser(tokens []token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (parser *Parser) Parse() error {
	program := nodeProgram{}
	err := program.parse(parser)
	if err != nil {
		return err
	}
	if !parser.isAtEnd() {
		return errors.New("invalid chars outside of function")
	}
	parser.Tree = program
	return nil
}

func (parser *Parser) isAtEnd() bool {
	if parser.index >= len(parser.tokens) {
		// Is at end
		return true
	}
	return false
}
func (parser *Parser) expect(expected tokenType) (bool, token) {
	if parser.isAtEnd() {
		return false, token{}
	}
	next_token := parser.tokens[parser.index]
	parser.index++

	if next_token.tokenType == expected {
		return true, next_token
	}
	return false, next_token

}

func (program *nodeProgram) parse(parser *Parser) error {
	function := nodeFunction{}
	err := function.parse(parser)
	if err != nil {
		return err
	}

	program.function = function

	return nil
}

func (function *nodeFunction) parse(parser *Parser) error {
	if exists, _ := parser.expect(tokenInt); !exists {
		return errors.New("missing int")
	}

	identifier := nodeIdentifier{}
	err := identifier.parse(parser)
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

	statement := nodeStatement{}
	err = statement.parse(parser)
	if err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenCloseBrace); !exists {
		return errors.New("missing closing brace")
	}

	function.name = identifier
	function.body = statement

	return nil
}

func (statement *nodeStatement) parse(parser *Parser) error {
	if exists, _ := parser.expect(tokenReturn); !exists {
		return errors.New("missing return")
	}

	return_val, err := parseExpression(parser)
	if err != nil {
		return err
	}

	statement.expression = return_val

	if exists, _ := parser.expect(tokenSemicolon); !exists {
		return errors.New("missing return")
	}

	return nil
}

func parseExpression(parser *Parser) (nodeExpression, error) {
	// Check for opening parenthesis
	exists, tok := parser.expect(tokenOpenParen)
	if exists {
		expression, err := parseExpression(parser)
		if err != nil {
			return nil, err
		}
		if exists, _ := parser.expect(tokenCloseParen); !exists {
			return nil, errors.New("missing closing parenthesis")
		}
		return expression, nil
	}

	// Check for unary operators
	if tok.tokenType == tokenBitwiseCompOp || tok.tokenType == tokenNegationOp {
		parser.index-- // Move back to reread the operator
		unop := nodeUnop{}
		err := unop.parse(parser)
		if err != nil {
			return nil, err
		}
		return &unop, nil
	}

	parser.index--

	// Must be a constant
	constant := &nodeInt{}
	err := constant.parse(parser)
	if err != nil {
		return nil, err
	}
	return constant, nil
}

func (unop *nodeUnop) parse(parser *Parser) error {
	exists, tok := parser.expect(tokenBitwiseCompOp)
	if exists {
		unop.opType = unopBitwiseComp
	} else if tok.tokenType == tokenNegationOp {
		unop.opType = unopNegate
	} else {
		return errors.New("expected unary operation")
	}

	expression, err := parseExpression(parser)
	if err != nil {
		return err
	}

	unop.exp = expression
	return nil
}

func (identifier *nodeIdentifier) parse(parser *Parser) error {
	exists, tok := parser.expect(tokenIdentfier)
	if !exists {
		return errors.New("missing identifier")
	}

	identifier.val = tok.literal

	return nil
}

func (nInt *nodeInt) parse(parser *Parser) error {
	exists, tok := parser.expect(tokenConstant)
	if !exists {
		return errors.New("missing int")
	}

	pint, _ := strconv.ParseInt(tok.literal, 10, 0)

	nInt.val = int(pint)

	return nil
}
