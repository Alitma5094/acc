package acc

import (
	"errors"
	"strconv"
)

type unopType int
type binopType int

const (
	unopBitwiseComp unopType = iota
	unopNegate
)

const (
	binopAdd binopType = iota
	binopSubtract
	binopMultiply
	binopDivide
	binopRemainder
)

type expression interface {
	nodeExpression()
	parse(*Parser) error
}

type factor interface {
	expression
	nodeFactor()
}

type Parser struct {
	tokens []token
	index  int
	Tree   nodeProgram
}

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
	return parser.index >= len(parser.tokens)
}

func (parser *Parser) expect(expected tokenType) (bool, token) {
	if parser.isAtEnd() {
		return false, token{}
	}
	nextToken := parser.tokens[parser.index]
	parser.index++
	if nextToken.tokenType == expected {
		return true, nextToken
	}
	return false, nextToken
}

func (parser *Parser) peek() token {
	if parser.isAtEnd() {
		return token{}
	}
	return parser.tokens[parser.index]
}

type nodeProgram struct {
	function nodeFunction
}

type nodeFunction struct {
	name nodeIdentifier
	body nodeStatement
}

type nodeStatement struct {
	expression expression
}

type nodeBinop struct {
	left   expression
	right  expression
	opType binopType
}

type nodeUnop struct {
	opType unopType
	exp    factor
}

type nodeFactor struct {
	factor factor
}

type nodeInt struct {
	val int
}

type nodeIdentifier struct {
	val string
}

func (*nodeBinop) nodeExpression()  {}
func (*nodeUnop) nodeExpression()   {}
func (*nodeFactor) nodeExpression() {}
func (*nodeInt) nodeExpression()    {}

func (*nodeInt) nodeFactor()    {}
func (*nodeUnop) nodeFactor()   {}
func (*nodeBinop) nodeFactor()  {}
func (*nodeFactor) nodeFactor() {}

func (program *nodeProgram) parse(parser *Parser) error {
	function := nodeFunction{}
	if err := function.parse(parser); err != nil {
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
	if err := identifier.parse(parser); err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenOpenParen); !exists {
		return errors.New("missing (")
	}
	if exists, _ := parser.expect(tokenVoid); !exists {
		return errors.New("missing void")
	}
	if exists, _ := parser.expect(tokenCloseParen); !exists {
		return errors.New("missing )")
	}
	if exists, _ := parser.expect(tokenOpenBrace); !exists {
		return errors.New("missing {")
	}

	statement := nodeStatement{}
	if err := statement.parse(parser); err != nil {
		return err
	}

	if exists, _ := parser.expect(tokenCloseBrace); !exists {
		return errors.New("missing }")
	}

	function.name = identifier
	function.body = statement
	return nil
}

func (statement *nodeStatement) parse(parser *Parser) error {
	if exists, _ := parser.expect(tokenReturn); !exists {
		return errors.New("missing return")
	}

	expr, err := parseExpression(parser, 0)
	if err != nil {
		return err
	}
	statement.expression = expr

	if exists, _ := parser.expect(tokenSemicolon); !exists {
		return errors.New("missing semicolon")
	}
	return nil
}

func parseExpression(parser *Parser, minPrecedence int) (expression, error) {
	left, err := parseFactor(parser)
	if err != nil {
		return nil, err
	}
	leftExpr := &nodeFactor{factor: left}

	for {
		nextToken := parser.peek()
		if binopPrecedence(nextToken) < minPrecedence {
			break
		}
		precedence := binopPrecedence(nextToken)

		bin := &nodeBinop{}
		if err := bin.parse(parser); err != nil {
			return nil, err
		}

		rightExpr, err := parseExpression(parser, precedence+1)
		if err != nil {
			return nil, err
		}

		bin.left = leftExpr
		bin.right = rightExpr
		leftExpr = &nodeFactor{factor: bin}
	}
	return leftExpr, nil
}

func parseFactor(parser *Parser) (factor, error) {
	switch parser.peek().tokenType {
	case tokenConstant:
		intNode := &nodeInt{}
		if err := intNode.parse(parser); err != nil {
			return nil, err
		}
		return intNode, nil

	case tokenNegationOp, tokenBitwiseCompOp:
		unopNode := &nodeUnop{}
		if err := unopNode.parse(parser); err != nil {
			return nil, err
		}
		return unopNode, nil

	case tokenOpenParen:
		parser.expect(tokenOpenParen)
		expr, err := parseExpression(parser, 0)
		if err != nil {
			return nil, err
		}
		if exists, _ := parser.expect(tokenCloseParen); !exists {
			return nil, errors.New("missing )")
		}
		// Must be wrapped in nodeFactor to satisfy `factor` interface
		return &nodeFactor{factor: expr.(factor)}, nil

	default:
		return nil, errors.New("malformed factor")
	}
}

func binopPrecedence(tok token) int {
	switch tok.tokenType {
	case tokenAdditionOp, tokenNegationOp:
		return 45
	case tokenMultiplicationOp, tokenDivisionOp, tokenRemainderOp:
		return 50
	default:
		return -1
	}
}

func (binop *nodeBinop) parse(parser *Parser) error {
	switch parser.peek().tokenType {
	case tokenAdditionOp:
		parser.expect(tokenAdditionOp)
		binop.opType = binopAdd
	case tokenNegationOp:
		parser.expect(tokenNegationOp)
		binop.opType = binopSubtract
	case tokenMultiplicationOp:
		parser.expect(tokenMultiplicationOp)
		binop.opType = binopMultiply
	case tokenDivisionOp:
		parser.expect(tokenDivisionOp)
		binop.opType = binopDivide
	case tokenRemainderOp:
		parser.expect(tokenRemainderOp)
		binop.opType = binopRemainder
	default:
		return errors.New("expected binary operator")
	}
	return nil
}

func (unop *nodeUnop) parse(parser *Parser) error {
	switch parser.peek().tokenType {
	case tokenBitwiseCompOp:
		parser.expect(tokenBitwiseCompOp)
		unop.opType = unopBitwiseComp
	case tokenNegationOp:
		parser.expect(tokenNegationOp)
		unop.opType = unopNegate
	default:
		return errors.New("expected unary operator")
	}

	exp, err := parseFactor(parser)
	if err != nil {
		return err
	}
	unop.exp = exp
	return nil
}

func (factor *nodeFactor) parse(parser *Parser) error {
	innerFactor, err := parseFactor(parser)
	if err != nil {
		return err
	}
	factor.factor = innerFactor
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
		return errors.New("missing int constant")
	}
	pint, err := strconv.ParseInt(tok.literal, 10, 0)
	if err != nil {
		return err
	}
	nInt.val = int(pint)
	return nil
}
