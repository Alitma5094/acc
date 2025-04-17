package ir

import (
	"acc/internal/parser"
	"fmt"
)

type TACGenerator struct {
	instructions []Instruction
	tempCounter  int
}

func NewTACGenerator() *TACGenerator {
	return &TACGenerator{}
}

func (g *TACGenerator) makeTemporary() string {
	g.tempCounter++
	return fmt.Sprintf("t%d", g.tempCounter)
}

func (g *TACGenerator) Generate(node *parser.Program) (*Program, error) {
	result := node.Accept(g)

	// The result should be a Program
	if program, ok := result.(*Program); ok {
		return program, nil
	}

	return nil, fmt.Errorf("failed to generate TAC program")
}

func (g *TACGenerator) VisitProgram(node *parser.Program) interface{} {
	function := node.Function.Accept(g).(Function)

	// Set the final instructions list
	function.Body = g.instructions

	return &Program{Function: function}
}

// VisitFunction implements Visitor interface
func (g *TACGenerator) VisitFunction(node *parser.Function) interface{} {
	// Visit function body (which populates instructions)
	node.Body.Accept(g)

	return Function{
		Identifier: node.Name,
	}
}

// VisitStatement implements Visitor interface
func (g *TACGenerator) VisitStatement(node *parser.Statement) interface{} {
	// Visit expression and get its result value
	resultValue := (*node.Expression).Accept(g).(Value)

	// Create return instruction
	returnInstr := &ReturnInstr{Value: resultValue}
	g.instructions = append(g.instructions, returnInstr)

	return returnInstr
}

// VisitBinaryOp implements Visitor interface
func (g *TACGenerator) VisitBinaryExp(node *parser.BinaryExp) interface{} {
	// Visit left and right operands
	leftVal := node.Left.Accept(g).(Value)
	rightVal := node.Right.Accept(g).(Value)

	// Create a destination temporary variable
	destVar := &Variable{Identifier: g.makeTemporary()}

	// Create binary instruction
	binInstr := &BinaryInstr{
		Operator: node.Op,
		Src1:     leftVal,
		Src2:     rightVal,
		Dst:      destVar,
	}

	g.instructions = append(g.instructions, binInstr)

	return destVar
}

// VisitUnaryOp implements Visitor interface
func (g *TACGenerator) VisitUnaryFactor(node *parser.UnaryFactor) interface{} {
	// Visit the operand
	sourceVal := node.Value.Accept(g).(Value)

	// Create a destination temporary variable
	destVar := &Variable{Identifier: g.makeTemporary()}

	// Create unary instruction
	unInstr := &UnaryInstr{
		Operator: node.Op,
		Src:      sourceVal,
		Dst:      destVar,
	}

	g.instructions = append(g.instructions, unInstr)

	return destVar
}

// VisitInt implements Visitor interface
func (g *TACGenerator) VisitIntLiteral(node *parser.IntLiteral) interface{} {
	return &Constant{Value: node.Value}
}

func (g *TACGenerator) VisitFactorExp(node *parser.FactorExp) interface{} {
	// TODO: Should this do anything?
	return nil
}
