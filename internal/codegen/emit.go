package codegen

import (
	"fmt"
	"runtime"
)

func (program *Program) EmitAsm() string {
	ending := ""
	if runtime.GOOS == "linux" {
		ending = "\n\t.section .note.GNU-stack,\"\",@progbits"
	}
	return fmt.Sprint(program.Function.EmitAsm(), ending)
}

func (function *Function) EmitAsm() string {
	i := ""
	for _, in := range function.Instructions {
		i += in.EmitAsm()
	}
	name := function.Name
	if runtime.GOOS == "darwin" {
		name = fmt.Sprint("_", name)
	}
	return fmt.Sprintf("\t.global %s\n%s:\n\tpushq\t%%rbp\n\tmovq\t%%rsp, %%rbp\n%s", name, name, i)
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

func (i *Cmp) EmitAsm() string {
	return fmt.Sprintf("\tcmpl\t%s, %s\n", i.Operand1.EmitAsm(), i.Operand2.EmitAsm())
}

func (r *Idiv) EmitAsm() string {
	return fmt.Sprintf("\tidivl\t%s\n", r.Operand.EmitAsm())
}

func (r *Cdq) EmitAsm() string {
	return "\tcdq\n"
}

func (i *Jmp) EmitAsm() string {
	identifier := i.Identifier
	if runtime.GOOS == "darwin" {
		identifier = fmt.Sprint("L", identifier)
	} else {
		// Linux
		identifier = fmt.Sprint(".L", identifier)
	}

	return fmt.Sprintf("\tjmp\t%s\n", identifier)
}
func (i *JmpCC) EmitAsm() string {
	identifier := i.Identifier
	if runtime.GOOS == "darwin" {
		identifier = fmt.Sprint("L", identifier)
	} else {
		// Linux
		identifier = fmt.Sprint(".L", identifier)
	}

	return fmt.Sprintf("\tj%s\t%s\n", i.Condition.EmitAsm(), identifier)
}
func (i *SetCC) EmitAsm() string {
	op := i.Operand.EmitAsm()
	if reg, opIsReg := i.Operand.(*Reg); opIsReg {
		op = reg.EmitAsm1Bit()
	}
	return fmt.Sprintf("\tset%s\t%s\n", i.Condition.EmitAsm(), op)
}
func (i *Label) EmitAsm() string {
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("L%s:\n", i.Identifier)
	} else {
		// Linux
		return fmt.Sprintf(".L%s:\n", i.Identifier)
	}
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
		return "%r11d"
	default:
		panic(fmt.Sprintf("invalid register type: %d", r.Reg))
	}
}

func (r *Reg) EmitAsm1Bit() string {
	switch r.Reg {
	case regAX:
		return "%al"
	case regDX:
		return "%dl"
	case regR10:
		return "%r10b"
	case regR11:
		return "%r11b"
	default:
		panic(fmt.Sprintf("invalid 1-bit register type: %d", r.Reg))
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

func (o CondCode) EmitAsm() string {
	switch o {
	case CondE:
		return "e"
	case CondNE:
		return "ne"
	case CondL:
		return "l"
	case CondLE:
		return "le"
	case CondG:
		return "g"
	case CondGE:
		return "ge"
	default:
		panic(fmt.Sprintf("invalid binary operator type: %d", 0))
	}
}
