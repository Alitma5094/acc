package parser

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
	Function *Function
}

type Function struct {
	Name IdentifierFactor
	Body []BlockItem
}

type StmtBlock struct {
	Statement Statement
}
type DeclarationBlock struct {
	Declaration Declaration
}

type ReturnStmt struct {
	Expression Expression
}

type ExpressionStmt struct {
	Expression Expression
}

type NullStmt struct{}

type BinaryExp struct {
	Left  Expression
	Op    BinopType
	Right Expression
}

type FactorExp struct {
	Factor Factor
}

type IntLiteral struct {
	Value int
}

type UnaryFactor struct {
	Op    UnopType
	Value Factor
}

type NestedExp struct {
	Expr Expression
}

type AssignmentExp struct {
	Left  Expression
	Right Expression
}

type IdentifierFactor struct {
	Value string
}

type Declaration struct {
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
