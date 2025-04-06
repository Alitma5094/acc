package acc

import (
	"errors"
	"fmt"
)

type AsmParser struct {
	ast  tacProgram
	Tree asmProgram
}

func NewAsmParser(tree tacProgram) *AsmParser {
	return &AsmParser{ast: tree}
}

func (p *AsmParser) Emit() string {
	return p.Tree.emitASM()
}

func (p *AsmParser) Parse() error {
	// Conert TAC to ASM
	prog := asmProgram{}
	err := prog.parse(p.ast)
	if err != nil {
		return err
	}
	p.Tree = prog

	// Replace psudoregisters
	stackIndex := 4
	stackVars := map[string]int{}
	for i, v := range p.Tree.function.instructions {
		switch inst := v.(type) {
		case *asmMov:
			if j, ok := inst.src.(*asmPseudo); ok {
				if val, ok := stackVars[j.identifier]; ok {
					inst.src = &asmStack{val: val}
				} else {
					stackVars[j.identifier] = stackIndex
					inst.src = &asmStack{val: stackIndex}
					stackIndex += 4
				}
				p.Tree.function.instructions[i] = inst
			}
			if j, ok := inst.dst.(*asmPseudo); ok {
				if val, ok := stackVars[j.identifier]; ok {
					inst.dst = &asmStack{val: val}
				} else {
					stackVars[j.identifier] = stackIndex
					inst.dst = &asmStack{val: stackIndex}
					stackIndex += 4
				}
				p.Tree.function.instructions[i] = inst
			}
			_, srcIsStack := inst.src.(*asmStack)
			_, dstIsStack := inst.dst.(*asmStack)
			if srcIsStack && dstIsStack {
				p.Tree.function.instructions[i] = &asmMov{src: inst.src, dst: &asmReg{reg: regR10}}
				p.Tree.function.instructions = append(p.Tree.function.instructions[:i+1], append([]asmInstruction{&asmMov{src: &asmReg{reg: regR10}, dst: inst.dst}}, p.Tree.function.instructions[i+1:]...)...)
			}

		case *asmUnary:
			if j, ok := inst.operand.(*asmPseudo); ok {
				if val, ok := stackVars[j.identifier]; ok {
					p.Tree.function.instructions[i].(*asmUnary).operand = &asmStack{val: val}
				} else {

					p.Tree.function.instructions[i].(*asmUnary).operand = &asmStack{val: stackIndex}
					stackIndex += 4
				}
			}

		}
	}

	// Insert allocate stack instruction
	p.Tree.function.instructions = append([]asmInstruction{&asmAllocateStack{val: stackIndex}}, p.Tree.function.instructions...)

	return nil

}

type asmProgram struct {
	function asmFunction
}

func (p *asmProgram) parse(n tacProgram) error {
	funct := asmFunction{}
	err := funct.parse(*n.function_definition)
	if err != nil {
		return err
	}
	p.function = funct
	return nil
}
func (p *asmProgram) emitASM() string {
	return p.function.emitASM()
}

type asmFunction struct {
	name         string
	instructions []asmInstruction
}

func (f *asmFunction) parse(n tacFunction) error {
	f.name = n.identifier
	var instructions []asmInstruction
	for _, i := range n.body {
		switch instr := i.(type) {
		case *tacReturn:
			src, err := convertOperand(instr.val)
			if err != nil {
				return err
			}
			instructions = append(instructions, &asmMov{src: src, dst: &asmReg{reg: regAX}})
			instructions = append(instructions, &asmRet{})

		case *tacUnary:
			src, err := convertOperand(instr.src)
			if err != nil {
				return err
			}
			dst, err := convertOperand(instr.dst)
			if err != nil {
				return err
			}
			op, err := convertOp(instr.operator)
			if err != nil {
				return err
			}

			instructions = append(instructions, &asmMov{src: src, dst: dst})
			instructions = append(instructions, &asmUnary{operator: op, operand: dst})
		default:
			return fmt.Errorf("invalid instruction type: %T", i)
		}
	}
	f.instructions = instructions
	return nil

}

func (f *asmFunction) emitASM() string {
	i := ""
	for _, in := range f.instructions {
		i += in.emitASM()
	}
	return fmt.Sprintf("\t.global _%s\n_%s:\n\tpushq\t%%rbp\n\tmovq\t%%rsp, %%rbp\n%s", f.name, f.name, i)
}

func convertOperand(n tacVal) (asmOperand, error) {
	switch op := n.(type) {
	case *tacConstant:
		o := &asmImn{val: op.val}
		return o, nil

	case *tacVar:
		o := &asmPseudo{identifier: op.identifier}
		return o, nil
	default:
		return nil, fmt.Errorf("invalid operand type: %T", n)
	}
}

type asmInstruction interface {
	instr()
	emitASM() string
}

type asmMov struct {
	src asmOperand
	dst asmOperand
}

func (i *asmMov) instr() {}
func (m *asmMov) emitASM() string {
	return fmt.Sprintf("\tmovl\t%s, %s\n", m.src.emitASM(), m.dst.emitASM())
}

type asmUnary struct {
	operator asmUnaryOp
	operand  asmOperand
}

func (i *asmUnary) instr() {}
func (r *asmUnary) emitASM() string {
	return fmt.Sprintf("\t%s\t%s\n", r.operator.emitASM(), r.operand.emitASM())
}

type asmAllocateStack struct {
	val int
}

func (i *asmAllocateStack) instr() {}
func (r *asmAllocateStack) emitASM() string {
	return fmt.Sprintf("\tsubq\t$%d, %%rsp\n", r.val)
}

type asmRet struct {
}

func (i *asmRet) instr() {}
func (r *asmRet) emitASM() string {
	return "\tmovq\t%rbp, %rsp\n\tpopq\t%rbp\n\tret\n"
}

type asmOperand interface {
	op()
	emitASM() string
}

type asmImn struct {
	val int
}

func (o *asmImn) op() {}
func (r *asmImn) emitASM() string {
	return fmt.Sprintf("$%d", r.val)
}

type asmReg struct {
	reg asmRegister
}

func (o *asmReg) op() {}
func (r *asmReg) emitASM() string {
	switch r.reg {
	case regAX:
		return "%eax"
	case regR10:
		return "%r10d"
	default:
		panic(fmt.Sprintf("invalid register type: %d", r.reg))
	}
}

type asmPseudo struct {
	identifier string
}

func (o *asmPseudo) op() {}
func (r *asmPseudo) emitASM() string {
	panic("pseudo registers not allowed in final asm")
}

type asmStack struct {
	val int
}

func (o *asmStack) op() {}
func (o *asmStack) emitASM() string {
	return fmt.Sprintf("-%d(%%rbp)", o.val)
}

type asmRegister int

const (
	regAX asmRegister = iota
	regR10
)

type asmUnaryOp int

const (
	opNeg asmUnaryOp = iota
	opNot
)

func convertOp(n unopType) (asmUnaryOp, error) {
	switch n {
	case unopBitwiseComp:
		return opNot, nil
	case unopNegate:
		return opNeg, nil
	default:
		return -1, errors.New("invalid operation type")
	}
}
func (o asmUnaryOp) emitASM() string {
	switch o {
	case opNeg:
		return "negl"
	case opNot:
		return "notl"
	default:
		panic(fmt.Sprintf("invalid unary operator type: %d", 0))
	}
}
