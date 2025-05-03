package ir

import (
	"acc/internal/common/errors"
	"acc/internal/parser"
	"fmt"
)

type TACGenerator struct {
	instructions   []Instruction
	tempVarCounter int
	labelCounter   int
}

// Accept starting number from semantic analysis
func NewTACGenerator(startVar int) *TACGenerator {
	return &TACGenerator{tempVarCounter: startVar}
}

func (g *TACGenerator) makeTemporaryVar() string {
	g.tempVarCounter++
	return fmt.Sprintf("tmp.%d", g.tempVarCounter)
}
func (g *TACGenerator) makeLabel(prefix string) string {
	g.labelCounter++
	return fmt.Sprintf("%s.%d", prefix, g.labelCounter)
}

func (g *TACGenerator) Generate(node *parser.Program) (*Program, error) {
	result := node.Accept(g)

	// The result should be a Program
	if program, ok := result.(*Program); ok {
		return program, nil
	}

	return nil, errors.NewIRGenError("top TAC node is not a Program")
}

func (g *TACGenerator) VisitProgram(node *parser.Program) interface{} {
	function := node.Function.Accept(g).(Function)

	function.Body = g.instructions

	return &Program{Function: function}
}

func (g *TACGenerator) VisitFunction(node *parser.Function) interface{} {
	node.Body.Accept(g)

	// Handle situation where function has no return statement; if function has return statement this will do nothing
	g.instructions = append(g.instructions, &ReturnInstr{Value: &Constant{Value: 0}})

	return Function{
		Identifier: node.Name.Value,
	}
}

func (g *TACGenerator) VisitBlock(node *parser.Block) any {
	for _, v := range node.Body {
		v.Accept(g)
	}
	return nil
}

func (g *TACGenerator) VisitReturnStatement(node *parser.ReturnStmt) any {
	resultValue := node.Expression.Accept(g).(Value)

	returnInstr := &ReturnInstr{Value: resultValue}
	g.instructions = append(g.instructions, returnInstr)

	return returnInstr
}

func (g *TACGenerator) VisitDeclaration(node *parser.Declaration) any {
	if node.Init == nil {
		return nil
	}

	initValue := node.Init.Accept(g).(Value)

	variable := &Variable{Identifier: g.makeTemporaryVar()}

	copyInstr := &CopyInstr{Src: initValue, Dst: variable}
	g.instructions = append(g.instructions, copyInstr, &CopyInstr{Src: variable, Dst: &Variable{Identifier: node.Name.Value}})

	return variable

}

func (g *TACGenerator) VisitNullStatement(node *parser.NullStmt) any {
	return nil
}

func (g *TACGenerator) VisitBinaryExp(node *parser.BinaryExp) interface{} {
	if node.Op == parser.BinopAnd {
		leftVal := node.Left.Accept(g).(Value)
		falseLabel := g.makeLabel("and_false")
		endLabel := g.makeLabel("and_end")
		dstVar := &Variable{Identifier: g.makeTemporaryVar()}

		g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: leftVal, Target: falseLabel})

		rightVal := node.Right.Accept(g).(Value)
		g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: rightVal, Target: falseLabel},
			&CopyInstr{Src: &Constant{Value: 1}, Dst: dstVar},
			&JumpInstr{Identifier: endLabel},
			&LabelInstr{Identifier: falseLabel},
			&CopyInstr{Src: &Constant{Value: 0}, Dst: dstVar},
			&LabelInstr{Identifier: endLabel})
		return dstVar

	} else if node.Op == parser.BinopOr {
		leftVal := node.Left.Accept(g).(Value)
		trueLabel := g.makeLabel("or_true")
		endLabel := g.makeLabel("or_end")
		dstVar := &Variable{Identifier: g.makeTemporaryVar()}

		g.instructions = append(g.instructions, &JumpIfNotZeroInstr{Condition: leftVal, Target: trueLabel})

		rightVal := node.Right.Accept(g).(Value)
		g.instructions = append(g.instructions, &JumpIfNotZeroInstr{Condition: rightVal, Target: trueLabel},
			&CopyInstr{Src: &Constant{Value: 0}, Dst: dstVar},
			&JumpInstr{Identifier: endLabel},
			&LabelInstr{Identifier: trueLabel},
			&CopyInstr{Src: &Constant{Value: 1}, Dst: dstVar},
			&LabelInstr{Identifier: endLabel})
		return dstVar
	} else {
		// Visit left and right operands
		leftVal := node.Left.Accept(g).(Value)
		rightVal := node.Right.Accept(g).(Value)

		// Create a destination temporary variable
		destVar := &Variable{Identifier: g.makeTemporaryVar()}

		binInstr := &BinaryInstr{
			Operator: node.Op,
			Src1:     leftVal,
			Src2:     rightVal,
			Dst:      destVar,
		}

		g.instructions = append(g.instructions, binInstr)

		return destVar
	}
}

func (g *TACGenerator) VisitAssignmentExp(node *parser.AssignmentExp) any {
	right := node.Right.Accept(g).(Value)
	left := node.Left.Accept(g).(*Variable)
	g.instructions = append(g.instructions, &CopyInstr{Src: right, Dst: left})
	return left

}

func (g *TACGenerator) VisitIdentifierFactor(node *parser.IdentifierFactor) any {
	return &Variable{Identifier: node.Value}
}

func (g *TACGenerator) VisitUnaryFactor(node *parser.UnaryFactor) interface{} {
	// Visit the operand
	sourceVal := node.Value.Accept(g).(Value)

	// Create a destination temporary variable
	destVar := &Variable{Identifier: g.makeTemporaryVar()}

	// Create unary instruction
	unInstr := &UnaryInstr{
		Operator: node.Op,
		Src:      sourceVal,
		Dst:      destVar,
	}

	g.instructions = append(g.instructions, unInstr)

	return destVar
}

func (g *TACGenerator) VisitIntLiteral(node *parser.IntLiteral) interface{} {
	return &Constant{Value: node.Value}
}

func (g *TACGenerator) VisitConditionalExp(node *parser.ConditionalExp) any {
	condition := node.Condition.Accept(g).(Value)
	e2Label := g.makeLabel("conditional_e2")
	endLabel := g.makeLabel("conditional_end")
	dstVar := &Variable{Identifier: g.makeTemporaryVar()}

	g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: condition, Target: e2Label})

	v1 := node.Expression1.Accept(g).(Value)

	g.instructions = append(g.instructions, &CopyInstr{Src: v1, Dst: dstVar}, &JumpInstr{Identifier: endLabel}, &LabelInstr{Identifier: e2Label})

	v2 := node.Expression2.Accept(g).(Value)

	g.instructions = append(g.instructions, &CopyInstr{Src: v2, Dst: dstVar}, &LabelInstr{Identifier: endLabel})
	return dstVar
}
func (g *TACGenerator) VisitIfStatement(node *parser.IfStmt) any {
	if node.Else == nil {
		condition := node.Condition.Accept(g).(Value)
		endLabel := g.makeLabel("if_end")

		g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: condition, Target: endLabel})

		node.Then.Accept(g)
		g.instructions = append(g.instructions, &LabelInstr{Identifier: endLabel})
		return nil
	} else {
		condition := node.Condition.Accept(g).(Value)
		elseLabel := g.makeLabel("if_else")
		endLabel := g.makeLabel("if_end")

		g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: condition, Target: elseLabel})

		node.Then.Accept(g)
		g.instructions = append(g.instructions, &JumpInstr{Identifier: endLabel}, &LabelInstr{Identifier: elseLabel})

		node.Else.Accept(g)
		g.instructions = append(g.instructions, &LabelInstr{Identifier: endLabel})
		return nil
	}
}

func (g *TACGenerator) VisitBreakStatement(node *parser.BreakStmt) any {
	instr := &JumpInstr{Identifier: fmt.Sprint("break_", node.Label)}
	g.instructions = append(g.instructions, instr)
	return instr
}

func (g *TACGenerator) VisitContinueStatement(node *parser.ContinueStmt) any {
	instr := &JumpInstr{Identifier: fmt.Sprint("continue_", node.Label)}
	g.instructions = append(g.instructions, instr)
	return instr
}

func (g *TACGenerator) VisitDoWhileStatement(node *parser.DoWhileStmt) any {
	conditionVar := &Variable{Identifier: g.makeTemporaryVar()}
	startLabel := fmt.Sprint("start_", node.Label)
	continueLabel := fmt.Sprint("continue_", node.Label)
	breakLabel := fmt.Sprint("break_", node.Label)

	g.instructions = append(g.instructions, &LabelInstr{Identifier: startLabel})
	node.Body.Accept(g)
	g.instructions = append(g.instructions, &LabelInstr{Identifier: continueLabel})
	condition := node.Condition.Accept(g).(Value)
	g.instructions = append(g.instructions, &CopyInstr{Src: condition, Dst: conditionVar}, &JumpIfNotZeroInstr{Condition: conditionVar, Target: startLabel}, &LabelInstr{Identifier: breakLabel})
	return nil
}

func (g *TACGenerator) VisitForStatement(node *parser.ForStmt) any {
	conditionVar := &Variable{Identifier: g.makeTemporaryVar()}
	startLabel := fmt.Sprint("start_", node.Label)
	breakLabel := fmt.Sprint("break_", node.Label)
	continueLabel := fmt.Sprint("continue_", node.Label)

	if node.Init != nil {
		node.Init.Accept(g)
	}
	g.instructions = append(g.instructions, &LabelInstr{Identifier: startLabel})
	if node.Condition != nil {
		condition := node.Condition.Accept(g).(Value)
		g.instructions = append(g.instructions, &CopyInstr{Src: condition, Dst: conditionVar}, &JumpIfZeroInstr{Condition: conditionVar, Target: breakLabel})
	} else {
		g.instructions = append(g.instructions, &JumpIfZeroInstr{Condition: &Constant{Value: 1}, Target: breakLabel})
	}
	node.Body.Accept(g)
	g.instructions = append(g.instructions, &LabelInstr{Identifier: continueLabel})
	if node.Post != nil {
		node.Post.Accept(g)
	}
	g.instructions = append(g.instructions, &JumpInstr{Identifier: startLabel}, &LabelInstr{Identifier: breakLabel})
	return nil
}

func (g *TACGenerator) VisitWhileStatement(node *parser.WhileStmt) any {
	conditionVar := &Variable{Identifier: g.makeTemporaryVar()}
	continueLabel := fmt.Sprint("continue_", node.Label)
	breakLabel := fmt.Sprint("break_", node.Label)

	g.instructions = append(g.instructions, &LabelInstr{Identifier: continueLabel})
	condition := node.Condition.Accept(g).(Value)
	g.instructions = append(g.instructions, &CopyInstr{Src: condition, Dst: conditionVar}, &JumpIfZeroInstr{Condition: conditionVar, Target: breakLabel})
	node.Body.Accept(g)
	g.instructions = append(g.instructions, &JumpInstr{Identifier: continueLabel}, &LabelInstr{Identifier: breakLabel})
	return nil
}
