package acc

import (
	"fmt"
)

type tacIntruction interface {
	tacInstruction()
}

type tacVal interface {
	tacVal()
}

type TacParser struct {
	ast          nodeProgram
	Tree         tacProgram
	tempCounter  int
	instructions []tacIntruction
}

type tacProgram struct {
	function_definition *tacFunction
}

type tacFunction struct {
	identifier string
	body       []tacIntruction
}

type tacReturn struct {
	val tacVal
}

type tacUnary struct {
	operator unopType
	src      tacVal
	dst      tacVal
}

type tacBinary struct {
	operator binopType
	src1     tacVal
	src2     tacVal
	dst      tacVal
}

type tacConstant struct {
	val int
}

type tacVar struct {
	identifier string
}

func (i *tacReturn) tacInstruction() {}
func (i *tacUnary) tacInstruction()  {}
func (i *tacBinary) tacInstruction() {}
func (v *tacConstant) tacVal()       {}
func (v *tacVar) tacVal()            {}

func NewTacParser(tree nodeProgram) *TacParser {
	return &TacParser{ast: tree}
}

func (p *TacParser) makeTemporary() string {
	p.tempCounter++
	return fmt.Sprintf("t%d", p.tempCounter)
}

func (p *TacParser) Parse() error {
	n := tacProgram{}
	if err := n.parse(p.ast, p); err != nil {
		return err
	}
	n.function_definition.body = p.instructions
	p.Tree = n
	return nil
}

func (p *tacProgram) parse(n nodeProgram, parser *TacParser) error {
	f := &tacFunction{}
	if err := f.parse(n.function, parser); err != nil {
		return err
	}
	p.function_definition = f
	return nil
}

func (f *tacFunction) parse(n nodeFunction, parser *TacParser) error {
	f.identifier = n.name.val

	exp, err := parseTACExpression(n.body.expression, parser)
	if err != nil {
		return err
	}

	// Final return instruction
	ret := &tacReturn{val: exp}
	parser.instructions = append(parser.instructions, ret)

	return nil
}

func parseTACExpression(exp expression, parser *TacParser) (tacVal, error) {
	switch e := exp.(type) {
	case *nodeFactor:
		return parseTACExpression(e.factor, parser)

	case *nodeUnop:
		u := &tacUnary{}
		if err := u.parse(*e, parser); err != nil {
			return nil, err
		}
		parser.instructions = append(parser.instructions, u)
		return u.dst, nil

	case *nodeBinop:
		b := &tacBinary{}
		if err := b.parse(*e, parser); err != nil {
			return nil, err
		}
		parser.instructions = append(parser.instructions, b)
		return b.dst, nil

	case *nodeInt:
		return &tacConstant{val: e.val}, nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", exp)
	}
}

func (p *tacUnary) parse(n nodeUnop, parser *TacParser) error {
	p.operator = n.opType

	// Parse the operand
	src, err := parseTACExpression(n.exp, parser)
	if err != nil {
		return err
	}

	p.src = src
	p.dst = &tacVar{identifier: parser.makeTemporary()}
	return nil
}

func (b *tacBinary) parse(n nodeBinop, parser *TacParser) error {
	// Parse left operand
	left, err := parseTACExpression(n.left, parser)
	if err != nil {
		return err
	}

	// Parse right operand
	right, err := parseTACExpression(n.right, parser)
	if err != nil {
		return err
	}

	b.operator = n.opType
	b.src1 = left
	b.src2 = right
	b.dst = &tacVar{identifier: parser.makeTemporary()}

	return nil
}
