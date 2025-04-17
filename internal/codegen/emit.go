package codegen

import "fmt"

func (program *Program) EmitAsm() string {
	return program.Function.EmitAsm()
}

func (function *Function) EmitAsm() string {
	i := ""
	for _, in := range function.Instructions {
		i += in.EmitAsm()
	}
	return fmt.Sprintf("\t.global _%s\n_%s:\n\tpushq\t%%rbp\n\tmovq\t%%rsp, %%rbp\n%s", function.Name, function.Name, i)
}

func (move *Mov) EmitAsm() string {
	return fmt.Sprintf("\tmovl\t%s, %s\n", move.Src.EmitAsm(), move.Dst.EmitAsm())
}

func (r *Unary) EmitAsm() string {
	return fmt.Sprintf("\t%s\t%s\n", r.Operator.EmitAsm(), r.Operand.EmitAsm())
}
func (r *Binary) EmitAsm() string {
	return fmt.Sprintf("\t%s\t%s, %s\n", r.Operator.EmitAsm(), r.Operand1.EmitAsm(), r.Operand2.EmitAsm())
}
func (r *Idiv) EmitAsm() string {
	return fmt.Sprintf("\tidivl\t%s\n", r.Operand.EmitAsm())
}
func (r *Cdq) EmitAsm() string {
	return "\tcdq\n"
}

func (r *AllocateStack) EmitAsm() string {
	return fmt.Sprintf("\tsubq\t$%d, %%rsp\n", r.Val)
}

func (r *Ret) EmitAsm() string {
	return "\tmovq\t%rbp, %rsp\n\tpopq\t%rbp\n\tret\n"
}

func (r *Imn) EmitAsm() string {
	return fmt.Sprintf("$%d", r.Val)
}

func (r *Reg) EmitAsm() string {
	switch r.Reg {
	case regAX:
		return "%eax"
	case regDX:
		return "%edx"
	case regR10:
		return "%r10d"
	case regR11:
		return "%r10d"
	default:
		panic(fmt.Sprintf("invalid register type: %d", r.Reg))
	}
}

func (r *Pseudo) EmitAsm() string {
	panic("pseudo registers not allowed in final asm")
}

func (o *Stack) EmitAsm() string {
	return fmt.Sprintf("-%d(%%rbp)", o.Val)
}
func (o UnaryOp) EmitAsm() string {
	switch o {
	case opNeg:
		return "negl"
	case opNot:
		return "notl"
	default:
		panic(fmt.Sprintf("invalid unary operator type: %d", 0))
	}
}

func (o BinaryOp) EmitAsm() string {
	switch o {
	case opAdd:
		return "addl"
	case opSub:
		return "subl"
	case opMult:
		return "imull"
	default:
		panic(fmt.Sprintf("invalid binary operator type: %d", 0))
	}
}
