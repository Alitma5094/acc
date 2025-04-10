package acc

import (
	"errors"
	"fmt"
)

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

type asmInstruction interface {
	instr()
	emitASM() string
}
type asmOperand interface {
	op()
	emitASM() string
}

type AsmParser struct {
	ast  tacProgram
	Tree asmProgram
}

type asmProgram struct {
	function asmFunction
}

type asmFunction struct {
	name         string
	instructions []asmInstruction
}

type asmMov struct {
	src asmOperand
	dst asmOperand
}

type asmUnary struct {
	operator asmUnaryOp
	operand  asmOperand
}
type asmRet struct {
}

type asmAllocateStack struct {
	val int
}

type asmImn struct {
	val int
}

type asmReg struct {
	reg asmRegister
}

type asmPseudo struct {
	identifier string
}

type asmStack struct {
	val int
}

func (i *asmMov) instr()           {}
func (i *asmAllocateStack) instr() {}
func (i *asmUnary) instr()         {}
func (i *asmRet) instr()           {}

func (o *asmImn) op()    {}
func (o *asmReg) op()    {}
func (o *asmPseudo) op() {}
func (o *asmStack) op()  {}

func NewAsmParser(tree tacProgram) *AsmParser {
	return &AsmParser{ast: tree}
}

func (parser *AsmParser) Emit() string {
	return parser.Tree.emitASM()
}

func (parser *AsmParser) Parse() error {
	// Conert TAC to ASM
	prog := asmProgram{}
	err := prog.parse(parser.ast)
	if err != nil {
		return err
	}
	parser.Tree = prog

	// Replace psudoregisters
	stackIndex := 4
	stackVars := map[string]int{}
	for i, v := range parser.Tree.function.instructions {
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
				parser.Tree.function.instructions[i] = inst
			}
			if j, ok := inst.dst.(*asmPseudo); ok {
				if val, ok := stackVars[j.identifier]; ok {
					inst.dst = &asmStack{val: val}
				} else {
					stackVars[j.identifier] = stackIndex
					inst.dst = &asmStack{val: stackIndex}
					stackIndex += 4
				}
				parser.Tree.function.instructions[i] = inst
			}
			_, srcIsStack := inst.src.(*asmStack)
			_, dstIsStack := inst.dst.(*asmStack)
			if srcIsStack && dstIsStack {
				parser.Tree.function.instructions[i] = &asmMov{src: inst.src, dst: &asmReg{reg: regR10}}
				parser.Tree.function.instructions = append(parser.Tree.function.instructions[:i+1], append([]asmInstruction{&asmMov{src: &asmReg{reg: regR10}, dst: inst.dst}}, parser.Tree.function.instructions[i+1:]...)...)
			}

		case *asmUnary:
			if j, ok := inst.operand.(*asmPseudo); ok {
				if val, ok := stackVars[j.identifier]; ok {
					parser.Tree.function.instructions[i].(*asmUnary).operand = &asmStack{val: val}
				} else {

					parser.Tree.function.instructions[i].(*asmUnary).operand = &asmStack{val: stackIndex}
					stackIndex += 4
				}
			}

		}
	}

	// Insert allocate stack instruction
	parser.Tree.function.instructions = append([]asmInstruction{&asmAllocateStack{val: stackIndex}}, parser.Tree.function.instructions...)

	return nil

}
func (program *asmProgram) parse(n tacProgram) error {
	funct := asmFunction{}
	err := funct.parse(*n.function_definition)
	if err != nil {
		return err
	}
	program.function = funct
	return nil
}
func (function *asmFunction) parse(n tacFunction) error {
	function.name = n.identifier

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
	function.instructions = instructions
	return nil

}
func (program *asmProgram) emitASM() string {
	return program.function.emitASM()
}

func (function *asmFunction) emitASM() string {
	i := ""
	for _, in := range function.instructions {
		i += in.emitASM()
	}
	return fmt.Sprintf("\t.global _%s\n_%s:\n\tpushq\t%%rbp\n\tmovq\t%%rsp, %%rbp\n%s", function.name, function.name, i)
}

func (move *asmMov) emitASM() string {
	return fmt.Sprintf("\tmovl\t%s, %s\n", move.src.emitASM(), move.dst.emitASM())
}

func (r *asmUnary) emitASM() string {
	return fmt.Sprintf("\t%s\t%s\n", r.operator.emitASM(), r.operand.emitASM())
}

func (r *asmAllocateStack) emitASM() string {
	return fmt.Sprintf("\tsubq\t$%d, %%rsp\n", r.val)
}

func (r *asmRet) emitASM() string {
	return "\tmovq\t%rbp, %rsp\n\tpopq\t%rbp\n\tret\n"
}

func (r *asmImn) emitASM() string {
	return fmt.Sprintf("$%d", r.val)
}

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

func (r *asmPseudo) emitASM() string {
	panic("pseudo registers not allowed in final asm")
}

func (o *asmStack) emitASM() string {
	return fmt.Sprintf("-%d(%%rbp)", o.val)
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
