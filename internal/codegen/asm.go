package codegen

type Register int

const (
	regAX Register = iota
	regDX
	regR10
	regR11
)

type UnaryOp int

const (
	opNeg UnaryOp = iota
	opNot
)

type BinaryOp int

const (
	opAdd BinaryOp = iota
	opSub
	opMult
)

type CondCode int

const (
	CondE CondCode = iota
	CondNE
	CondG
	CondGE
	CondL
	CondLE
)

type Instruction interface {
	instr()
	EmitAsm() string
}
type Operand interface {
	op()
	EmitAsm() string
}

type Program struct {
	Function Function
}

type Function struct {
	Name         string
	Instructions []Instruction
}

type Mov struct {
	Src Operand
	Dst Operand
}

type Unary struct {
	Operator UnaryOp
	Operand  Operand
}

type Binary struct {
	Operator BinaryOp
	Operand1 Operand
	Operand2 Operand
}

type Cmp struct {
	Operand1 Operand
	Operand2 Operand
}

type Idiv struct {
	Operand Operand
}

type Cdq struct {
}

type Jmp struct {
	Identifier string
}

type JmpCC struct {
	Condition  CondCode
	Identifier string
}

type SetCC struct {
	Condition CondCode
	Operand   Operand
}

type Label struct {
	Identifier string
}

type Ret struct {
}

type AllocateStack struct {
	Val int
}

type Imn struct {
	Val int
}

type Reg struct {
	Reg Register
}

type Pseudo struct {
	Identifier string
}

type Stack struct {
	Val int
}

func (i *Mov) instr()           {}
func (i *AllocateStack) instr() {}
func (i *Unary) instr()         {}
func (i *Binary) instr()        {}
func (i *Cmp) instr()           {}
func (i *Jmp) instr()           {}
func (i *JmpCC) instr()         {}
func (i *SetCC) instr()         {}
func (i *Label) instr()         {}
func (i *Idiv) instr()          {}
func (i *Cdq) instr()           {}
func (i *Ret) instr()           {}

func (o *Imn) op()    {}
func (o *Reg) op()    {}
func (o *Pseudo) op() {}
func (o *Stack) op()  {}
