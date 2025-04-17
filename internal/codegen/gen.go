package codegen

import (
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

	return fmt.Errorf("failed to generate TAC program")
}

func (g *AsmGenerator) VisitProgram(node *ir.Program) interface{} {
	function := node.Function.Accept(g).(Function)
	return &Program{Function: function}
}

func (g *AsmGenerator) VisitFunction(node *ir.Function) interface{} {
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
		default:
			panic(fmt.Sprintf("invalid instruction type: %T", i))
		}
	}
	function.Instructions = instructions
	return function
}

func (g *AsmGenerator) VisitReturnInstr(node *ir.ReturnInstr) interface{} {
	src := g.convertOperand(node.Value)
	return []Instruction{&Mov{Src: src, Dst: &Reg{Reg: regAX}}, &Ret{}}
}

func (g *AsmGenerator) VisitUnaryInstr(node *ir.UnaryInstr) interface{} {
	src := g.convertOperand(node.Src)

	dst := g.convertOperand(node.Dst)

	op := convertUnOp(node.Operator)

	return []Instruction{&Mov{Src: src, Dst: dst}, &Unary{Operator: op, Operand: dst}}
}

func (g *AsmGenerator) VisitBinaryInstr(node *ir.BinaryInstr) interface{} {
	instructions := []Instruction{}

	if node.Operator == parser.BinopDivide {
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)

		instructions = append(instructions, &Mov{Src: src1, Dst: &Reg{Reg: regAX}})
		instructions = append(instructions, &Cdq{})
		instructions = append(instructions, &Idiv{Operand: src2})
		instructions = append(instructions, &Mov{Src: &Reg{Reg: regAX}, Dst: dst})
	} else if node.Operator == parser.BinopRemainder {
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)

		instructions = append(instructions, &Mov{Src: src1, Dst: &Reg{Reg: regAX}})
		instructions = append(instructions, &Cdq{})
		instructions = append(instructions, &Idiv{Operand: src2})
		instructions = append(instructions, &Mov{Src: &Reg{Reg: regDX}, Dst: dst})
	} else {
		src1 := g.convertOperand(node.Src1)
		src2 := g.convertOperand(node.Src2)
		dst := g.convertOperand(node.Dst)
		op := convertBinOp(node.Operator)

		instructions = append(instructions, &Mov{Src: src1, Dst: dst})
		instructions = append(instructions, &Binary{Operator: op, Operand1: src2, Operand2: dst})
	}
	return instructions
}

func (g *AsmGenerator) VisitConstant(node *ir.Constant) interface{} {
	return &Imn{Val: node.Value}
}

func (g *AsmGenerator) VisitVariable(node *ir.Variable) interface{} {
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
