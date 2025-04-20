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
	VisitCopyInstr(node *CopyInstr) any
	VisitJumpInstr(node *JumpInstr) any
	VisitJumpIfZeroInstr(node *JumpIfZeroInstr) any
	VisitJumpIfNotZeroInstr(node *JumpIfNotZeroInstr) any
	VisitLabelInstr(node *LabelInstr) any
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

type CopyInstr struct {
	Src Value
	Dst Value
}

func (i *CopyInstr) instr() {}

func (p *CopyInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitCopyInstr(p)
}

type JumpInstr struct {
	Identifier string
}

func (i *JumpInstr) instr() {}
func (p *JumpInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitJumpInstr(p)
}

type JumpIfZeroInstr struct {
	Condition Value
	Target    string
}

func (i *JumpIfZeroInstr) instr() {}
func (p *JumpIfZeroInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitJumpIfZeroInstr(p)
}

type JumpIfNotZeroInstr struct {
	Condition Value
	Target    string
}

func (i *JumpIfNotZeroInstr) instr() {}
func (p *JumpIfNotZeroInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitJumpIfNotZeroInstr(p)
}

type LabelInstr struct {
	Identifier string
}

func (i *LabelInstr) instr() {}
func (p *LabelInstr) Accept(visitor TacVisitor) any {
	return visitor.VisitLabelInstr(p)
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
