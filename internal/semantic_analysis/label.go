package semanticanalysis

import (
	"acc/internal/common/errors"
	"acc/internal/parser"
	"fmt"
)

func (a *SemanticAnalyzer) makeLabel() string {
	a.TempVarCounter++
	return fmt.Sprintf("loop.%d", a.TempVarCounter)
}

func (a *SemanticAnalyzer) LabelLoops() error {
	return a.labelBlock(a.program.Function.Body, "")
}

func (a *SemanticAnalyzer) labelBlock(block parser.Block, currentLabel string) error {
	for _, v := range block.Body {
		if stmt, ok := v.(*parser.StmtBlock); ok {
			err := a.labelStatement(stmt.Statement, currentLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *SemanticAnalyzer) labelStatement(stmt parser.Statement, currentLabel string) error {
	switch stmt := stmt.(type) {
	case *parser.BreakStmt:
		if currentLabel == "" {
			return errors.NewAnalysisError("break statement outside of loop", errors.Location{})
		}
		stmt.Label = currentLabel
		return nil
	case *parser.ContinueStmt:
		if currentLabel == "" {
			return errors.NewAnalysisError("continue statement outside of loop", errors.Location{})
		}
		stmt.Label = currentLabel
		return nil
	case *parser.WhileStmt:
		newLabel := a.makeLabel()
		err := a.labelStatement(stmt.Body, newLabel)
		if err != nil {
			return err
		}
		stmt.Label = newLabel
		return nil
	case *parser.DoWhileStmt:
		newLabel := a.makeLabel()
		err := a.labelStatement(stmt.Body, newLabel)
		if err != nil {
			return err
		}
		stmt.Label = newLabel
		return nil
	case *parser.ForStmt:
		newLabel := a.makeLabel()
		err := a.labelStatement(stmt.Body, newLabel)
		if err != nil {
			return err
		}
		stmt.Label = newLabel
		return nil
	case *parser.IfStmt:
		err := a.labelStatement(stmt.Then, currentLabel)
		if err != nil {
			return err
		}
		if stmt.Else != nil {
			err := a.labelStatement(stmt.Else, currentLabel)
			if err != nil {
				return err
			}
		}
		return nil
	case *parser.CompoundStmt:
		return a.labelBlock(stmt.Block, currentLabel)
	default:
		return nil
	}
}
