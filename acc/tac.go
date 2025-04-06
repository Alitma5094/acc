// package acc

// import (
// 	"fmt"
// )

// type TacParser struct {
// 	ast         nodeProgram
// 	Tree        tacProgram
// 	tempCounter int
// }

// func NewTacParser(tree nodeProgram) *TacParser {
// 	return &TacParser{ast: tree}
// }

// func (p *TacParser) makeTemporary() string {
// 	p.tempCounter++
// 	return fmt.Sprintf("t%d", p.tempCounter)
// }

// func (p *TacParser) Parse() error {
// 	n := tacProgram{}
// 	if err := n.parse(p.ast, p); err != nil {
// 		return err
// 	}
// 	p.Tree = n
// 	return nil
// }

// type tacProgram struct {
// 	function_definition *tacFunction
// }

// func (p *tacProgram) parse(n nodeProgram, parser *TacParser) error {
// 	f := &tacFunction{}
// 	if err := f.parse(n.function, parser); err != nil {
// 		return err
// 	}
// 	p.function_definition = f
// 	return nil
// }

// type tacFunction struct {
// 	identifier string
// 	body       []tacIntruction
// }

// func (f *tacFunction) parse(n nodeFunction, parser *TacParser) error {
// 	f.identifier = n.name.val
// 	switch exp := n.body.expression.(type) {
// 	case *nodeUnop:
// 		// Handle nested expressions by building up instructions
// 		var instructions []tacIntruction

// 		u := &tacUnary{}
// 		if err := u.parse(*exp, parser); err != nil {
// 			return err
// 		}
// 		instructions = append(instructions, u)

// 		// Add return instruction with the last temporary variable
// 		ret := &tacReturn{val: u.dst}
// 		instructions = append(instructions, ret)

// 		f.body = instructions
// 	case *nodeInt:
// 		i := tacReturn{}
// 		i.parse(*exp, parser)
// 		f.body = append(f.body, &i)
// 	default:
// 		return fmt.Errorf("got invalid expression type")
// 	}

// 	return nil
// }

// type tacIntruction interface {
// 	tacInstruction()
// }

// type tacReturn struct {
// 	val tacVal
// }

// func (i *tacReturn) tacInstruction() {}
// func (p *tacReturn) parse(n nodeInt, parser *TacParser) error {
// 	c := &tacConstant{val: n.val}
// 	p.val = c
// 	return nil
// }

// type tacUnary struct {
// 	operator unopType
// 	src      tacVal
// 	dst      tacVal
// }

// func (i *tacUnary) tacInstruction() {}
// func (p *tacUnary) parse(n nodeUnop, parser *TacParser) error {
// 	p.operator = n.opType

// 	switch exp := n.exp.(type) {
// 	case *nodeInt:
// 		p.src = &tacConstant{val: exp.val}
// 	case *nodeUnop:
// 		tempVap := tacVar{identifier: parser.makeTemporary()}
// 		nestedUnary := tacUnary{
// 			operator: exp.opType,
// 			dst:      &tempVap,
// 		}
// 		nestedUnary.parse(*exp, parser)
// 	default:
// 		return fmt.Errorf("invalid expression type in unary operator")
// 	}

// 	p.dst = &tacVar{identifier: parser.makeTemporary()}
// 	return nil
// }

// type tacVal interface {
// 	tacVal()
// }

// type tacConstant struct {
// 	val int
// }

// func (v *tacConstant) tacVal() {}

// type tacVar struct {
// 	identifier string
// }

// func (v *tacVar) tacVal() {}
package acc

import (
	"fmt"
)

type TacParser struct {
	ast          nodeProgram
	Tree         tacProgram
	tempCounter  int
	instructions []tacIntruction
}

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

type tacProgram struct {
	function_definition *tacFunction
}

func (p *tacProgram) parse(n nodeProgram, parser *TacParser) error {
	f := &tacFunction{}
	if err := f.parse(n.function, parser); err != nil {
		return err
	}
	p.function_definition = f
	return nil
}

type tacFunction struct {
	identifier string
	body       []tacIntruction
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

type tacIntruction interface {
	tacInstruction()
}

type tacReturn struct {
	val tacVal
}

func (i *tacReturn) tacInstruction() {}

func (p *tacReturn) parse(n nodeInt, parser *TacParser) error {
	c := &tacConstant{val: n.val}
	p.val = c
	return nil
}

type tacUnary struct {
	operator unopType
	src      tacVal
	dst      tacVal
}

func (i *tacUnary) tacInstruction() {}

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

type tacVal interface {
	tacVal()
}

type tacConstant struct {
	val int
}

func (v *tacConstant) tacVal() {}

type tacVar struct {
	identifier string
}

func (v *tacVar) tacVal() {}
