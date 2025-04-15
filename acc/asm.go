package acc

import (
	"errors"
	"fmt"
)

type asmRegister int

const (
	regAX asmRegister = iota
	regDX
	regR10
	regR11
)

type asmUnaryOp int

const (
	opNeg asmUnaryOp = iota
	opNot
)

type asmBinaryOp int

const (
	opAdd asmBinaryOp = iota
	opSub
	opMult
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

type asmBinary struct {
	operator asmBinaryOp
	operand1 asmOperand
	operand2 asmOperand
}

type asmIdiv struct {
	operand asmOperand
}

type asmCdq struct {
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
func (i *asmBinary) instr()        {}
func (i *asmIdiv) instr()          {}
func (i *asmCdq) instr()           {}
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
	// Covert TAC to ASM
	prog := asmProgram{}
	err := prog.parse(parser.ast)
	if err != nil {
		return err
	}
	parser.Tree = prog

	parser.fixInstructions()

	return nil

}

func (parser *AsmParser) fixInstructions() {
	stackAllocator := &stackAllocator{
		vars:         make(map[string]int),
		currentIndex: 4,
	}

	for i := 0; i < len(parser.Tree.function.instructions); i++ {
		inst := parser.Tree.function.instructions[i]

		switch inst := inst.(type) {
		case *asmMov:
			i += parser.fixMovInstruction(inst, i, stackAllocator)
		case *asmUnary:
			parser.fixUnaryInstruction(inst, i, stackAllocator)
		case *asmBinary:
			i += parser.fixBinaryInstruction(inst, i, stackAllocator)
		case *asmIdiv:
			i += parser.fixIdivInstruction(inst, i, stackAllocator)
		}
	}

	// Insert stack allocation instruction at the beginning
	parser.Tree.function.instructions = append(
		[]asmInstruction{&asmAllocateStack{val: stackAllocator.currentIndex}},
		parser.Tree.function.instructions...,
	)
}

type stackAllocator struct {
	vars         map[string]int
	currentIndex int
}

func (sa *stackAllocator) allocateVar(identifier string) *asmStack {
	if val, exists := sa.vars[identifier]; exists {
		return &asmStack{val: val}
	}

	stack := &asmStack{val: sa.currentIndex}
	sa.vars[identifier] = sa.currentIndex
	sa.currentIndex += 4
	return stack
}

func (parser *AsmParser) fixMovInstruction(inst *asmMov, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	modified := false

	// Handle source operand
	if src, ok := inst.src.(*asmPseudo); ok {
		inst.src = sa.allocateVar(src.identifier)
		modified = true
	}

	// Handle destination operand
	if dst, ok := inst.dst.(*asmPseudo); ok {
		inst.dst = sa.allocateVar(dst.identifier)
		modified = true
	}

	if modified {
		parser.Tree.function.instructions[index] = inst
	}

	// Handle stack-to-stack moves
	if _, srcIsStack := inst.src.(*asmStack); srcIsStack {
		if _, dstIsStack := inst.dst.(*asmStack); dstIsStack {
			// Replace with two instructions using temporary register
			parser.Tree.function.instructions[index] = &asmMov{
				src: inst.src,
				dst: &asmReg{reg: regR10},
			}

			parser.Tree.function.instructions = append(
				parser.Tree.function.instructions[:index+1],
				append(
					[]asmInstruction{
						&asmMov{
							src: &asmReg{reg: regR10},
							dst: inst.dst,
						},
					},
					parser.Tree.function.instructions[index+1:]...,
				)...,
			)
			return 1 // Indicate that extra instruction was added
		}
	}

	return 0
}

func (parser *AsmParser) fixUnaryInstruction(inst *asmUnary, index int, sa *stackAllocator) {
	// Replace pseudoregisters
	if operand, ok := inst.operand.(*asmPseudo); ok {
		inst.operand = sa.allocateVar(operand.identifier)
		parser.Tree.function.instructions[index] = inst
	}
}

func (parser *AsmParser) fixBinaryInstruction(inst *asmBinary, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	if operand1, ok := inst.operand1.(*asmPseudo); ok {
		inst.operand1 = sa.allocateVar(operand1.identifier)
		parser.Tree.function.instructions[index] = inst
	}

	if operand2, ok := inst.operand2.(*asmPseudo); ok {
		inst.operand2 = sa.allocateVar(operand2.identifier)
		parser.Tree.function.instructions[index] = inst
	}

	// Can't have mem address as both src and dst
	if _, dstIsOp := inst.operand2.(*asmStack); dstIsOp {
		if inst.operator == opMult {
			// imul cant have mem address as dst, reguardless of source
			parser.Tree.function.instructions[index] = &asmMov{
				src: inst.operand2,
				dst: &asmReg{reg: regR11},
			}

			parser.Tree.function.instructions = append(
				parser.Tree.function.instructions[:index+1],
				append(
					[]asmInstruction{
						&asmBinary{
							operator: inst.operator,
							operand1: inst.operand1,
							operand2: &asmReg{reg: regR11},
						},
						&asmMov{src: &asmReg{reg: regR11}, dst: inst.operand2},
					},
					parser.Tree.function.instructions[index+1:]...,
				)...,
			)
			return 2
		}

		if _, srcIsOp := inst.operand1.(*asmStack); srcIsOp {
			parser.Tree.function.instructions[index] = &asmMov{
				src: inst.operand1,
				dst: &asmReg{reg: regR10},
			}

			parser.Tree.function.instructions = append(
				parser.Tree.function.instructions[:index+1],
				append(
					[]asmInstruction{
						&asmBinary{
							operator: inst.operator,
							operand1: &asmReg{reg: regR10},
							operand2: inst.operand2,
						},
					},
					parser.Tree.function.instructions[index+1:]...,
				)...,
			)
			return 1
		}
	}

	return 0
}

func (parser *AsmParser) fixIdivInstruction(inst *asmIdiv, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	if operand, ok := inst.operand.(*asmPseudo); ok {
		inst.operand = sa.allocateVar(operand.identifier)
		parser.Tree.function.instructions[index] = inst
		return 0
	}

	// idivl can't operate on constants, copy value into scratch register
	if constant, ok := inst.operand.(*asmImn); ok {
		parser.Tree.function.instructions[index] = &asmMov{
			src: constant,
			dst: &asmReg{reg: regR10},
		}

		parser.Tree.function.instructions = append(
			parser.Tree.function.instructions[:index+1],
			append(
				[]asmInstruction{
					&asmIdiv{
						operand: &asmReg{reg: regR10},
					},
				},
				parser.Tree.function.instructions[index+1:]...,
			)...,
		)
		return 1
	}

	return 0
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
			op, err := convertUnOp(instr.operator)
			if err != nil {
				return err
			}

			instructions = append(instructions, &asmMov{src: src, dst: dst})
			instructions = append(instructions, &asmUnary{operator: op, operand: dst})
		case *tacBinary:
			if instr.operator == binopDivide {
				src1, err := convertOperand(instr.src1)
				if err != nil {
					return err
				}
				src2, err := convertOperand(instr.src2)
				if err != nil {
					return err
				}
				dst, err := convertOperand(instr.dst)
				if err != nil {
					return err
				}

				instructions = append(instructions, &asmMov{src: src1, dst: &asmReg{reg: regAX}})
				instructions = append(instructions, &asmCdq{})
				instructions = append(instructions, &asmIdiv{operand: src2})
				instructions = append(instructions, &asmMov{src: &asmReg{reg: regAX}, dst: dst})
			} else if instr.operator == binopRemainder {
				src1, err := convertOperand(instr.src1)
				if err != nil {
					return err
				}
				src2, err := convertOperand(instr.src2)
				if err != nil {
					return err
				}
				dst, err := convertOperand(instr.dst)
				if err != nil {
					return err
				}

				instructions = append(instructions, &asmMov{src: src1, dst: &asmReg{reg: regAX}})
				instructions = append(instructions, &asmCdq{})
				instructions = append(instructions, &asmIdiv{operand: src2})
				instructions = append(instructions, &asmMov{src: &asmReg{reg: regDX}, dst: dst})
			} else {
				src1, err := convertOperand(instr.src1)
				if err != nil {
					return err
				}
				src2, err := convertOperand(instr.src2)
				if err != nil {
					return err
				}
				dst, err := convertOperand(instr.dst)
				if err != nil {
					return err
				}
				op, err := convertBinOp(instr.operator)
				if err != nil {
					return err
				}

				instructions = append(instructions, &asmMov{src: src1, dst: dst})
				instructions = append(instructions, &asmBinary{operator: op, operand1: src2, operand2: dst})
			}
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
func (r *asmBinary) emitASM() string {
	return fmt.Sprintf("\t%s\t%s, %s\n", r.operator.emitASM(), r.operand1.emitASM(), r.operand2.emitASM())
}
func (r *asmIdiv) emitASM() string {
	return fmt.Sprintf("\tidivl\t%s\n", r.operand.emitASM())
}
func (r *asmCdq) emitASM() string {
	return "\tcdq\n"
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
	case regDX:
		return "%edx"
	case regR10:
		return "%r10d"
	case regR11:
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

func (o asmBinaryOp) emitASM() string {
	switch o {
	case opAdd:
		return "addl"
	case opSub:
		return "subl"
	case opMult:
		return "imull"
	default:
		panic(fmt.Sprintf("invalid binary operator type: %d", 0))
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
func convertUnOp(n unopType) (asmUnaryOp, error) {
	switch n {
	case unopBitwiseComp:
		return opNot, nil
	case unopNegate:
		return opNeg, nil
	default:
		return -1, errors.New("invalid operation type")
	}
}

func convertBinOp(n binopType) (asmBinaryOp, error) {
	switch n {
	case binopAdd:
		return opAdd, nil
	case binopSubtract:
		return opSub, nil
	case binopMultiply:
		return opMult, nil
	default:
		return -1, errors.New("invalid operation type")
	}
}
