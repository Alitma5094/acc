package codegen

import (
	"acc/internal/common/errors"
	"acc/internal/ir"
	"acc/internal/parser"
	"fmt"
)

type AsmGenerator struct {
	Program    *Program
	stackAlloc *stackAllocator
}

func NewASMGenerator() *AsmGenerator {
	return &AsmGenerator{
		stackAlloc: &stackAllocator{
			Variables:    make(map[string]int),
			CurrentIndex: 4,
		},
	}
}

func (g *AsmGenerator) Generate(node *ir.Program) error {
	result := node.Accept(g)

	// The result should be a Program
	program, ok := result.(*Program)
	if ok {
		g.Program = program
		return nil
	}

	return errors.NewCodeGenError("top ASM node os not a Program")
}

func (g *AsmGenerator) VisitProgram(node *ir.Program) any {
	function := node.Function.Accept(g).(Function)
	return &Program{Function: function}
}

func (g *AsmGenerator) VisitFunction(node *ir.Function) any {
	function := Function{Name: node.Identifier}
	var instructions []Instruction

	for _, i := range node.Body {
		switch instr := i.(type) {
		case *ir.ReturnInstr:
			instructions = append(instructions, instr.Accept(g).([]Instruction)...)
		case *ir.UnaryInstr:
			instructions = append(instructions, instr.Accept(g).([]Instruction)...)
		case *ir.BinaryInstr:
			instructions = append(instructions, instr.Accept(g).([]Instruction)...)
		case *ir.CopyInstr:
			instructions = append(instructions, instr.Accept(g).(Instruction))
		case *ir.JumpInstr:
			instructions = append(instructions, instr.Accept(g).(Instruction))
		case *ir.JumpIfZeroInstr:
			instructions = append(instructions, instr.Accept(g).([]Instruction)...)
		case *ir.JumpIfNotZeroInstr:
			instructions = append(instructions, instr.Accept(g).([]Instruction)...)
		case *ir.LabelInstr:
			instructions = append(instructions, instr.Accept(g).(Instruction))
		default:
			panic(fmt.Sprintf("invalid instruction type: %T", i))
		}
	}
	function.Instructions = instructions
	return function
}

func (g *AsmGenerator) VisitReturnInstr(node *ir.ReturnInstr) any {
	src := g.convertOperand(node.Value)
	return []Instruction{&Mov{Src: src, Dst: &Reg{Reg: regAX}}, &Ret{}}
}

func (g *AsmGenerator) VisitUnaryInstr(node *ir.UnaryInstr) interface{} {
	src := g.convertOperand(node.Src)

	dst := g.convertOperand(node.Dst)

	if node.Operator == parser.UnopNot {
		return []Instruction{&Cmp{Operand1: &Imn{0}, Operand2: src}, &Mov{Src: &Imn{0}, Dst: dst}, &SetCC{Condition: CondE, Operand: dst}}
	}

	op := convertUnOp(node.Operator)

	return []Instruction{&Mov{Src: src, Dst: dst}, &Unary{Operator: op, Operand: dst}}
}

func (g *AsmGenerator) VisitBinaryInstr(node *ir.BinaryInstr) any {
	instructions := []Instruction{}

	switch node.Operator {
	case parser.BinopDivide:
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)

		instructions = append(instructions, &Mov{Src: src1, Dst: &Reg{Reg: regAX}})
		instructions = append(instructions, &Cdq{})
		instructions = append(instructions, &Idiv{Operand: src2})
		instructions = append(instructions, &Mov{Src: &Reg{Reg: regAX}, Dst: dst})
	case parser.BinopRemainder:
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)

		instructions = append(instructions, &Mov{Src: src1, Dst: &Reg{Reg: regAX}})
		instructions = append(instructions, &Cdq{})
		instructions = append(instructions, &Idiv{Operand: src2})
		instructions = append(instructions, &Mov{Src: &Reg{Reg: regDX}, Dst: dst})
	case parser.BinopGreaterThan, parser.BinopGreaterOrEqual, parser.BinopLessThan, parser.BinopLessOrEqual, parser.BinopEqual, parser.BinopNotEqual:
		instructions = append(instructions, g.handleRelationalOp(node)...)
	default:
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)
		op := convertBinOp(node.Operator)

		instructions = append(instructions, &Mov{Src: src1, Dst: dst})
		instructions = append(instructions, &Binary{Operator: op, Operand1: src2, Operand2: dst})
	}
	return instructions
}

func (g *AsmGenerator) handleRelationalOp(node *ir.BinaryInstr) []Instruction {
	src1 := g.convertOperand(node.Src1)
	src2 := g.convertOperand(node.Src2)
	dst := g.convertOperand(node.Dst)
	instructions := []Instruction{&Cmp{Operand1: src2, Operand2: src1}, &Mov{Src: &Imn{Val: 0}, Dst: dst}}

	switch node.Operator {
	case parser.BinopLessThan:
		instructions = append(instructions, &SetCC{Condition: CondL, Operand: dst})
	case parser.BinopLessOrEqual:
		instructions = append(instructions, &SetCC{Condition: CondLE, Operand: dst})
	case parser.BinopGreaterThan:
		instructions = append(instructions, &SetCC{Condition: CondG, Operand: dst})
	case parser.BinopGreaterOrEqual:
		instructions = append(instructions, &SetCC{Condition: CondGE, Operand: dst})
	case parser.BinopEqual:
		instructions = append(instructions, &SetCC{Condition: CondE, Operand: dst})
	case parser.BinopNotEqual:
		instructions = append(instructions, &SetCC{Condition: CondNE, Operand: dst})
	default:
		panic("not a relational operation type")
	}
	return instructions
}

func (g *AsmGenerator) VisitCopyInstr(node *ir.CopyInstr) any {
	src := g.convertOperand(node.Src)
	dst := g.convertOperand(node.Dst)
	return &Mov{Src: src, Dst: dst}
}
func (g *AsmGenerator) VisitJumpInstr(node *ir.JumpInstr) any {
	return &Jmp{Identifier: node.Identifier}
}
func (g *AsmGenerator) VisitJumpIfZeroInstr(node *ir.JumpIfZeroInstr) any {
	return []Instruction{&Cmp{Operand1: &Imn{Val: 0}, Operand2: g.convertOperand(node.Condition)}, &JmpCC{Condition: CondE, Identifier: node.Target}}
}
func (g *AsmGenerator) VisitJumpIfNotZeroInstr(node *ir.JumpIfNotZeroInstr) any {
	return []Instruction{&Cmp{Operand1: &Imn{Val: 0}, Operand2: g.convertOperand(node.Condition)}, &JmpCC{Condition: CondNE, Identifier: node.Target}}
}
func (g *AsmGenerator) VisitLabelInstr(node *ir.LabelInstr) any {
	return &Label{Identifier: node.Identifier}
}

func (g *AsmGenerator) VisitConstant(node *ir.Constant) any {
	return &Imn{Val: node.Value}
}

func (g *AsmGenerator) VisitVariable(node *ir.Variable) any {
	return &Pseudo{Identifier: node.Identifier}
}

func (g *AsmGenerator) convertOperand(node ir.Value) Operand {
	switch op := node.(type) {
	case *ir.Constant:
		return op.Accept(g).(*Imn)

	case *ir.Variable:
		return op.Accept(g).(*Pseudo)
	default:
		panic(fmt.Sprintf("invalid operand type: %T", node))
	}
}

func convertUnOp(n parser.UnopType) UnaryOp {
	switch n {
	case parser.UnopBitwiseComp:
		return opNot
	case parser.UnopNegate:
		return opNeg
	default:
		panic("invalid unary operation type")
	}
}

func convertBinOp(n parser.BinopType) BinaryOp {
	switch n {
	case parser.BinopAdd:
		return opAdd
	case parser.BinopSubtract:
		return opSub
	case parser.BinopMultiply:
		return opMult
	default:
		panic("invalid binary operation type")
	}
}
