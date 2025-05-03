package semanticanalysis

import (
	"acc/internal/common/errors"
	"acc/internal/parser"
	"fmt"
)

type SemanticAnalyzer struct {
	variables      map[string]Variable
	TempVarCounter int
	program        parser.Program
}

type Variable struct {
	NewName          string
	FromCurrentBlock bool
}

func (a *SemanticAnalyzer) copyVars() map[string]Variable {
	newVar := make(map[string]Variable, len(a.variables))
	for k, v := range a.variables {
		newVar[k] = Variable{NewName: v.NewName, FromCurrentBlock: false}
	}
	return newVar
}

func NewSemanticAnalyzer(program parser.Program) SemanticAnalyzer {
	return SemanticAnalyzer{program: program, variables: make(map[string]Variable)}
}

func (a *SemanticAnalyzer) makeTemporaryVar(prefix string) string {
	a.TempVarCounter++
	return fmt.Sprintf("%d.%s", a.TempVarCounter, prefix)
}

func (a *SemanticAnalyzer) ResolveVariables() error {
	return a.resolveBlock(&a.program.Function.Body)
}

func (a *SemanticAnalyzer) resolveDeclaration(declaration *parser.Declaration) error {
	variable, ok := a.variables[declaration.Name.Value]
	if ok && variable.FromCurrentBlock {
		return errors.NewAnalysisError("duplicate variable declaration", declaration.Loc)
	}

	a.variables[declaration.Name.Value] = Variable{NewName: a.makeTemporaryVar(declaration.Name.Value), FromCurrentBlock: true}
	declaration.Name.Value = a.variables[declaration.Name.Value].NewName
	if declaration.Init != nil {
		err := a.resolveExpression(&declaration.Init)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *SemanticAnalyzer) resolveBlock(block *parser.Block) error {
	for _, item := range block.Body {
		switch item := item.(type) {
		case *parser.DeclarationBlock:
			err := a.resolveDeclaration(&item.Declaration)
			if err != nil {
				return err
			}
		case *parser.StmtBlock:
			err := a.resolveStatement(&item.Statement)
			if err != nil {
				return err
			}
		default:
			panic("invalid block item type")
		}
	}
	return nil
}

func (a *SemanticAnalyzer) resolveExpression(expression *parser.Expression) error {
	switch item := (*expression).(type) {
	case *parser.AssignmentExp:
		_, ok := item.Left.(*parser.FactorExp).Factor.(*parser.IdentifierFactor)
		if !ok {
			return errors.NewAnalysisError("invalid lvalue", item.Loc)
		}
		err := a.resolveExpression(&item.Left)
		if err != nil {
			return err
		}

		err = a.resolveExpression(&item.Right)
		if err != nil {
			return err
		}
		return nil
	case *parser.BinaryExp:
		err := a.resolveExpression(&item.Left)
		if err != nil {
			return err
		}

		err = a.resolveExpression(&item.Right)
		if err != nil {
			return err
		}
		return nil
	case *parser.FactorExp:
		return a.resolveFactor(&item.Factor)
	case *parser.ConditionalExp:
		err := a.resolveExpression(&item.Condition)
		if err != nil {
			return err
		}

		err = a.resolveExpression(&item.Expression1)
		if err != nil {
			return err
		}

		err = a.resolveExpression(&item.Expression2)
		if err != nil {
			return err
		}
		return nil
	default:
		panic("invalid expression type")

	}
}

func (a *SemanticAnalyzer) resolveStatement(statement *parser.Statement) error {
	switch item := (*statement).(type) {
	case *parser.ReturnStmt:
		return a.resolveExpression(&item.Expression)
	case *parser.ExpressionStmt:
		return a.resolveExpression(&item.Expression)
	case *parser.NullStmt:
		return nil
	case *parser.IfStmt:
		err := a.resolveExpression(&item.Condition)
		if err != nil {
			return err
		}
		err = a.resolveStatement(&item.Then)
		if err != nil {
			return err
		}
		if item.Else != nil {
			err = a.resolveStatement(&item.Else)
			if err != nil {
				return err
			}
		}
		return nil
	case *parser.CompoundStmt:
		oldVars := a.variables
		a.variables = a.copyVars()
		err := a.resolveBlock(&item.Block)
		a.variables = oldVars
		return err
	case *parser.WhileStmt:
		err := a.resolveExpression(&item.Condition)
		if err != nil {
			return err
		}
		err = a.resolveStatement(&item.Body)
		if err != nil {
			return err
		}
		return nil
	case *parser.DoWhileStmt:
		err := a.resolveStatement(&item.Body)
		if err != nil {
			return err
		}
		err = a.resolveExpression(&item.Condition)
		if err != nil {
			return err
		}
		return nil
	case *parser.ForStmt:
		oldVars := a.variables
		a.variables = a.copyVars()

		err := a.resolveForInit(item.Init)
		if err != nil {
			return err
		}

		err = a.resolveOptionalExpression(item.Condition)
		if err != nil {
			return err
		}

		err = a.resolveOptionalExpression(item.Post)
		if err != nil {
			return err
		}

		err = a.resolveStatement(&item.Body)
		if err != nil {
			return err
		}

		a.variables = oldVars
		return nil
	case *parser.BreakStmt:
		return nil
	case *parser.ContinueStmt:
		return nil
	default:
		panic("invalid statement type")

	}
}

func (a *SemanticAnalyzer) resolveForInit(forInit parser.ForInit) error {
	if forInit == nil {
		return nil
	}
	switch i := forInit.(type) {
	case *parser.InitExp:
		return a.resolveOptionalExpression(i.Expression)
	case *parser.InitDecl:
		return a.resolveDeclaration(&i.Declaration)

	default:
		panic("invalid for init type")
	}
}

func (a *SemanticAnalyzer) resolveOptionalExpression(exp parser.Expression) error {
	if exp == nil {
		return nil
	}

	return a.resolveExpression(&exp)
}

func (a *SemanticAnalyzer) resolveFactor(factor *parser.Factor) error {
	switch item := (*factor).(type) {
	case *parser.IntLiteral:
		return nil
	case *parser.UnaryFactor:
		return a.resolveFactor(&item.Value)
	case *parser.NestedExp:
		return a.resolveExpression(&item.Expr)
	case *parser.IdentifierFactor:
		if variable, ok := a.variables[item.Value]; ok {
			item.Value = variable.NewName
		} else {
			return errors.NewAnalysisError("undeclared variable", item.Loc)
		}
		return nil
	default:
		panic("invalid factor type")

	}
}
