package semanticanalysis

import (
	"acc/internal/common/errors"
	"acc/internal/parser"
	"fmt"
)

type SemanticAnalyzer struct {
	variables      map[string]string
	TempVarCounter int
	program        parser.Program
}

func NewSemanticAnalyzer(program parser.Program) SemanticAnalyzer {
	return SemanticAnalyzer{program: program, variables: make(map[string]string)}
}

func (a *SemanticAnalyzer) makeTemporaryVar(prefix string) string {
	a.TempVarCounter++
	return fmt.Sprintf("%d.%s", a.TempVarCounter, prefix)
}

func (a *SemanticAnalyzer) ResolveDeclarations() error {
	for _, item := range a.program.Function.Body {
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

func (a *SemanticAnalyzer) resolveDeclaration(declaration *parser.Declaration) error {
	_, ok := a.variables[declaration.Name.Value]
	if ok {
		return errors.NewAnalysisError("duplicate variable declaration", declaration.Loc)
	}

	a.variables[declaration.Name.Value] = a.makeTemporaryVar(declaration.Name.Value)
	declaration.Name.Value = a.variables[declaration.Name.Value]
	if declaration.Init != nil {
		err := a.resolveExpression(&declaration.Init)
		if err != nil {
			return err
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
		err := a.resolveFactor(&item.Factor)
		if err != nil {
			return err
		}
		return nil
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
		err := a.resolveExpression(&item.Expression)
		if err != nil {
			return err
		}
		return nil
	case *parser.ExpressionStmt:
		err := a.resolveExpression(&item.Expression)
		if err != nil {
			return err
		}
		return nil
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
	default:
		panic("invalid statement type")

	}
}

func (a *SemanticAnalyzer) resolveFactor(factor *parser.Factor) error {
	switch item := (*factor).(type) {
	case *parser.IntLiteral:
		return nil
	case *parser.UnaryFactor:
		err := a.resolveFactor(&item.Value)
		if err != nil {
			return err
		}
		return nil
	case *parser.NestedExp:
		err := a.resolveExpression(&item.Expr)
		if err != nil {
			return err
		}
		return nil
	case *parser.IdentifierFactor:
		if variable, ok := a.variables[item.Value]; ok {
			item.Value = variable
		} else {
			return errors.NewAnalysisError("undeclared variable", item.Loc)
		}
		return nil
	default:
		panic("invalid factor type")

	}
}
