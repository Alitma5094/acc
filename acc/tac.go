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

type tacConstant struct {
	val int
}

type tacVar struct {
	identifier string
}

func (i *tacReturn) tacInstruction() {}
func (i *tacUnary) tacInstruction()  {}
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

	switch exp := n.body.expression.(type) {
	case *nodeUnop:
		u := &tacUnary{}
		if err := u.parse(*exp, parser); err != nil {
			return err
		}
		parser.instructions = append(parser.instructions, u)

		ret := &tacReturn{val: u.dst}
		parser.instructions = append(parser.instructions, ret)

	case *nodeInt:
		ret := &tacReturn{}
		if err := ret.parse(*exp, parser); err != nil {
			return err
		}
		parser.instructions = append(parser.instructions, ret)

	default:
		return fmt.Errorf("got invalid expression type")
	}

	return nil
}

func (p *tacReturn) parse(n nodeInt, parser *TacParser) error {
	c := &tacConstant{val: n.val}
	p.val = c
	return nil
}

func (p *tacUnary) parse(n nodeUnop, parser *TacParser) error {
	p.operator = n.opType

	switch exp := n.exp.(type) {
	case *nodeInt:
		p.src = &tacConstant{val: exp.val}

	case *nodeUnop:
		nested := &tacUnary{}
		if err := nested.parse(*exp, parser); err != nil {
			return err
		}

		parser.instructions = append(parser.instructions, nested)

		p.src = nested.dst

	default:
		return fmt.Errorf("invalid expression type in unary operator")
	}

	p.dst = &tacVar{identifier: parser.makeTemporary()}
	return nil
}
