package parser

import "acc/internal/common/errors"

type BinopType int
type UnopType int

const (
	BinopAdd BinopType = iota
	BinopSubtract
	BinopMultiply
	BinopDivide
	BinopRemainder
	BinopAnd
	BinopOr
	BinopEqual
	BinopNotEqual
	BinopLessThan
	BinopLessOrEqual
	BinopGreaterThan
	BinopGreaterOrEqual
)

const (
	UnopBitwiseComp UnopType = iota
	UnopNegate
	UnopNot
)

type AstVisitor interface {
	VisitProgram(node *Program) any
	VisitFunction(node *Function) any
	VisitReturnStatement(node *ReturnStmt) any
	VisitNullStatement(node *NullStmt) any
	VisitDeclaration(node *Declaration) any
	VisitBinaryExp(node *BinaryExp) any
	VisitAssignmentExp(node *AssignmentExp) any
	VisitUnaryFactor(node *UnaryFactor) any
	VisitIdentifierFactor(node *IdentifierFactor) any
	VisitIntLiteral(node *IntLiteral) any
}

type Node interface {
	// Position() errors.Location
	Accept(visitor AstVisitor) any
}

type BlockItem interface {
	Node
	block()
}

type Expression interface {
	Node
	exp()
}

type Factor interface {
	Node
	factor()
}

type Statement interface {
	Node
	stmt()
}

type Program struct {
	Loc      errors.Location
	Function *Function
}

type Function struct {
	Loc  errors.Location
	Name IdentifierFactor
	Body []BlockItem
}

type StmtBlock struct {
	Loc       errors.Location
	Statement Statement
}
type DeclarationBlock struct {
	Loc         errors.Location
	Declaration Declaration
}

type ReturnStmt struct {
	Loc        errors.Location
	Expression Expression
}

type ExpressionStmt struct {
	Loc        errors.Location
	Expression Expression
}

type NullStmt struct {
	Loc errors.Location
}

type BinaryExp struct {
	Loc   errors.Location
	Left  Expression
	Op    BinopType
	Right Expression
}

type FactorExp struct {
	Loc    errors.Location
	Factor Factor
}

type IntLiteral struct {
	Loc   errors.Location
	Value int
}

type UnaryFactor struct {
	Loc   errors.Location
	Op    UnopType
	Value Factor
}

type NestedExp struct {
	Loc  errors.Location
	Expr Expression
}

type AssignmentExp struct {
	Loc   errors.Location
	Left  Expression
	Right Expression
}

type IdentifierFactor struct {
	Loc   errors.Location
	Value string
}

type Declaration struct {
	Loc  errors.Location
	Name IdentifierFactor
	Init Expression
}

func (p *Program) Accept(visitor AstVisitor) any {
	return visitor.VisitProgram(p)
}

func (f *Function) Accept(visitor AstVisitor) any {
	return visitor.VisitFunction(f)
}

func (s *StmtBlock) Accept(visitor AstVisitor) any {
	return s.Statement.Accept(visitor)
}

func (s *DeclarationBlock) Accept(visitor AstVisitor) any {
	return s.Declaration.Accept(visitor)
}

func (s *ReturnStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitReturnStatement(s)
}
func (s *ExpressionStmt) Accept(visitor AstVisitor) any {
	return s.Expression.Accept(visitor)
}

func (s *NullStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitNullStatement(s)
}

func (b *BinaryExp) Accept(visitor AstVisitor) any {
	return visitor.VisitBinaryExp(b)
}

func (n *FactorExp) Accept(visitor AstVisitor) any {
	return n.Factor.Accept(visitor)
}

func (i *IntLiteral) Accept(visitor AstVisitor) any {
	return visitor.VisitIntLiteral(i)
}

func (u *UnaryFactor) Accept(visitor AstVisitor) any {
	return visitor.VisitUnaryFactor(u)
}

func (u *NestedExp) Accept(visitor AstVisitor) any {
	return u.Expr.Accept(visitor)
}

func (u *AssignmentExp) Accept(visitor AstVisitor) any {
	return visitor.VisitAssignmentExp(u)
}

func (u *IdentifierFactor) Accept(visitor AstVisitor) any {
	return visitor.VisitIdentifierFactor(u)
}

func (u *Declaration) Accept(visitor AstVisitor) any {
	return visitor.VisitDeclaration(u)
}

// func (p *Program) Position() errors.Location {
// 	return p.Loc
// }

// func (f *Function) Position() errors.Location {
// 	return f.Loc
// }

// func (s *StmtBlock) Position() errors.Location {
// 	return s.Loc
// }

// func (s *DeclarationBlock) Position() errors.Location {
// 	return s.Loc
// }

// func (s *ReturnStmt) Position() errors.Location {
// 	return s.Loc
// }
// func (s *ExpressionStmt) Position() errors.Location {
// 	return s.Loc
// }

// func (s *NullStmt) Position() errors.Location {
// 	return s.Loc
// }

// func (b *BinaryExp) Position() errors.Location {
// 	return b.Loc
// }

// func (n *FactorExp) Position() errors.Location {
// 	return n.Loc
// }

// func (i *IntLiteral) Position() errors.Location {
// 	return i.Loc
// }

// func (u *UnaryFactor) Position() errors.Location {
// 	return u.Loc
// }

// func (u *NestedExp) Position() errors.Location {
// 	return u.Loc
// }

// func (u *AssignmentExp) Position() errors.Location {
// 	return u.Loc
// }

func (u *IdentifierFactor) Position() errors.Location {
	return u.Loc
}

func (u *Declaration) Position() errors.Location {
	return u.Loc
}

func (StmtBlock) block()        {}
func (DeclarationBlock) block() {}

func (ReturnStmt) stmt()     {}
func (ExpressionStmt) stmt() {}
func (NullStmt) stmt()       {}

func (BinaryExp) exp()     {}
func (FactorExp) exp()     {}
func (AssignmentExp) exp() {}

func (IntLiteral) factor()       {}
func (UnaryFactor) factor()      {}
func (NestedExp) factor()        {}
func (IdentifierFactor) factor() {}
