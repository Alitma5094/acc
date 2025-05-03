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
	VisitIfStatement(node *IfStmt) any
	VisitNullStatement(node *NullStmt) any
	VisitDeclaration(node *Declaration) any
	VisitBinaryExp(node *BinaryExp) any
	VisitAssignmentExp(node *AssignmentExp) any
	VisitConditionalExp(node *ConditionalExp) any
	VisitUnaryFactor(node *UnaryFactor) any
	VisitIdentifierFactor(node *IdentifierFactor) any
	VisitIntLiteral(node *IntLiteral) any
	VisitBlock(node *Block) any
	VisitBreakStatement(node *BreakStmt) any
	VisitContinueStatement(node *ContinueStmt) any
	VisitWhileStatement(node *WhileStmt) any
	VisitDoWhileStatement(node *DoWhileStmt) any
	VisitForStatement(node *ForStmt) any
}

type Node interface {
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

type ForInit interface {
	Node
	forInit()
}

type Program struct {
	Loc      errors.Location
	Function *Function
}

type Function struct {
	Loc  errors.Location
	Name IdentifierFactor
	Body Block
}

type Block struct {
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

type IfStmt struct {
	Loc       errors.Location
	Condition Expression
	Then      Statement
	Else      Statement
}

type CompoundStmt struct {
	Block Block
}

type BreakStmt struct {
	Label string
}

type ContinueStmt struct {
	Label string
}

type WhileStmt struct {
	Label     string
	Condition Expression
	Body      Statement
}

type DoWhileStmt struct {
	Label     string
	Body      Statement
	Condition Expression
}

type ForStmt struct {
	Label     string
	Init      ForInit
	Condition Expression
	Post      Expression
	Body      Statement
}

type InitDecl struct {
	Declaration Declaration
}
type InitExp struct {
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

type ConditionalExp struct {
	Loc         errors.Location
	Condition   Expression
	Expression1 Expression
	Expression2 Expression
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

func (s *IfStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitIfStatement(s)
}

func (s *CompoundStmt) Accept(visitor AstVisitor) any {
	return s.Block.Accept(visitor)
}

func (b *Block) Accept(visitor AstVisitor) any {
	return visitor.VisitBlock(b)
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

func (n *ConditionalExp) Accept(visitor AstVisitor) any {
	return visitor.VisitConditionalExp(n)
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

func (b *BreakStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitBreakStatement(b)
}
func (b *ContinueStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitContinueStatement(b)
}
func (b *WhileStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitWhileStatement(b)
}
func (b *DoWhileStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitDoWhileStatement(b)
}
func (b *ForStmt) Accept(visitor AstVisitor) any {
	return visitor.VisitForStatement(b)
}

func (f *InitDecl) Accept(visitor AstVisitor) any {
	return f.Declaration.Accept(visitor)
}
func (f *InitExp) Accept(visitor AstVisitor) any {
	return f.Expression.Accept(visitor)
}

func (StmtBlock) block()        {}
func (DeclarationBlock) block() {}

func (ReturnStmt) stmt()     {}
func (ExpressionStmt) stmt() {}
func (IfStmt) stmt()         {}
func (CompoundStmt) stmt()   {}
func (BreakStmt) stmt()      {}
func (ContinueStmt) stmt()   {}
func (WhileStmt) stmt()      {}
func (DoWhileStmt) stmt()    {}
func (ForStmt) stmt()        {}
func (NullStmt) stmt()       {}

func (BinaryExp) exp()      {}
func (FactorExp) exp()      {}
func (AssignmentExp) exp()  {}
func (ConditionalExp) exp() {}

func (IntLiteral) factor()       {}
func (UnaryFactor) factor()      {}
func (NestedExp) factor()        {}
func (IdentifierFactor) factor() {}

func (InitDecl) forInit() {}
func (InitExp) forInit()  {}
