package ir

import (
	"acc/internal/parser"
)

type TacNode interface {
	Accept(visitor TacVisitor) any
}
type TacVisitor interface {
	VisitProgram(node *Program) any
	VisitFunction(node *Function) any
	VisitReturnInstr(node *ReturnInstr) any
	VisitUnaryInstr(node *UnaryInstr) any
	VisitBinaryInstr(node *BinaryInstr) any
	VisitConstant(node *Constant) any
	VisitVariable(node *Variable) any
}

type Instruction interface {
	TacNode
	instr()
}

type Value interface {
	TacNode
	val()
}

type Program struct {
	Function Function
}

func (p *Program) Accept(visitor TacVisitor) any {
	return visitor.VisitProgram(p)
}

type Function struct {
	Identifier string
	Body       []Instruction
}

func (p *Function) Accept(visitor TacVisitor) any {
	return visitor.VisitFunction(p)
}

type ReturnInstr struct {
	Value Value
}

func (i *ReturnInstr) instr() {}

func (p *ReturnInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitReturnInstr(p)
}

type UnaryInstr struct {
	Operator parser.UnopType
	Src      Value
	Dst      Value
}

func (i *UnaryInstr) instr() {}

func (p *UnaryInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitUnaryInstr(p)
}

type BinaryInstr struct {
	Operator parser.BinopType
	Src1     Value
	Src2     Value
	Dst      Value
}

func (i *BinaryInstr) instr() {}
func (p *BinaryInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitBinaryInstr(p)
}

type Constant struct {
	Value int
}

func (v *Constant) val() {}

func (p *Constant) Accept(visitor TacVisitor) any {
	return visitor.VisitConstant(p)
}

type Variable struct {
	Identifier string
}

func (v *Variable) val() {}

func (p *Variable) Accept(visitor TacVisitor) any {
	return visitor.VisitVariable(p)
}
