package parser

type Node interface {
	// Position() errors.Location
	Accept(visitor AstVisitor) any
}

type AstVisitor interface {
	VisitProgram(node *Program) any
	VisitFunction(node *Function) any
	VisitStatement(node *Statement) any
	VisitBinaryExp(node *BinaryExp) any
	VisitUnaryFactor(node *UnaryFactor) any
	VisitIntLiteral(node *IntLiteral) any
}

type Program struct {
	Function *Function
}

func (p *Program) Accept(visitor AstVisitor) any {
	return visitor.VisitProgram(p)
}

type Function struct {
	Name string
	Body *Statement
}

func (f *Function) Accept(visitor AstVisitor) any {
	return visitor.VisitFunction(f)
}

type Statement struct {
	Expression *Expression
}

func (s *Statement) Accept(visitor AstVisitor) any {
	return visitor.VisitStatement(s)
}

type Expression interface {
	Node
	exp()
}

type BinaryExp struct {
	Left  Expression
	Op    BinopType
	Right Expression
}

func (BinaryExp) exp() {}
func (b *BinaryExp) Accept(visitor AstVisitor) any {
	return visitor.VisitBinaryExp(b)
}

type FactorExp struct {
	Factor Factor
}

func (FactorExp) exp() {}
func (n *FactorExp) Accept(visitor AstVisitor) any {
	return n.Factor.Accept(visitor)
}

type Factor interface {
	Node
	factor()
}

type IntLiteral struct {
	Value int
}

func (IntLiteral) factor() {}

func (i *IntLiteral) Accept(visitor AstVisitor) any {
	return visitor.VisitIntLiteral(i)
}

type UnaryFactor struct {
	Op    UnopType
	Value Factor
}

func (UnaryFactor) factor() {}

func (u *UnaryFactor) Accept(visitor AstVisitor) any {
	return visitor.VisitUnaryFactor(u)
}

type NestedExp struct {
	Expr Expression
}

func (NestedExp) factor() {}

func (u *NestedExp) Accept(visitor AstVisitor) any {
	return u.Expr.Accept(visitor)
}

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
