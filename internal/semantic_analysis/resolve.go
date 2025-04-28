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

func (a *SemanticAnalyzer) ResolveDelclatarions() error {
	for _, item := range a.program.Function.Body {
		switch item := item.(type) {
		case *parser.DeclarationBlock:
			err := a.resolveDelcatation(&item.Declaration)
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

func (a *SemanticAnalyzer) resolveDelcatation(declatation *parser.Declaration) error {
	_, ok := a.variables[declatation.Name.Value]
	if ok {
		return errors.NewAnalysisError("duplicate variable declaration", declatation.Loc)
	}

	a.variables[declatation.Name.Value] = a.makeTemporaryVar(declatation.Name.Value)
	declatation.Name.Value = a.variables[declatation.Name.Value]
	if declatation.Init != nil {
		err := a.resolveExpression(&declatation.Init)
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
			return errors.NewAnalysisError("undelcared variable", item.Loc)
		}
		return nil
	default:
		panic("invalid factor type")

	}
}
