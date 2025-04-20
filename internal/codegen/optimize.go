package codegen

type stackAllocator struct {
	Variables    map[string]int
	CurrentIndex int
}

func (g *AsmGenerator) FixInstructions() {
	stackAllocator := &stackAllocator{
		Variables:    make(map[string]int),
		CurrentIndex: 4,
	}

	for i := 0; i < len(g.Program.Function.Instructions); i++ {
		inst := g.Program.Function.Instructions[i]

		switch inst := inst.(type) {
		case *Mov:
			i += g.fixMovInstruction(inst, i, stackAllocator)
		case *Unary:
			g.fixUnaryInstruction(inst, i, stackAllocator)
		case *Binary:
			i += g.fixBinaryInstruction(inst, i, stackAllocator)
		case *Idiv:
			i += g.fixIdivInstruction(inst, i, stackAllocator)
		case *Cmp:
			i += g.fixCmpInstruction(inst, i, stackAllocator)
		case *SetCC:
			g.fixSetCCInstruction(inst, i, stackAllocator)
		}
	}

	// Insert stack allocation instruction at the beginning
	g.Program.Function.Instructions = append(
		[]Instruction{&AllocateStack{Val: stackAllocator.CurrentIndex}},
		g.Program.Function.Instructions...,
	)
}

func (sa *stackAllocator) allocateVar(identifier string) *Stack {
	if val, exists := sa.Variables[identifier]; exists {
		return &Stack{Val: val}
	}

	stack := &Stack{Val: sa.CurrentIndex}
	sa.Variables[identifier] = sa.CurrentIndex
	sa.CurrentIndex += 4
	return stack
}

func (g *AsmGenerator) fixMovInstruction(inst *Mov, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	modified := false

	// Handle source operand
	if src, ok := inst.Src.(*Pseudo); ok {
		inst.Src = sa.allocateVar(src.Identifier)
		modified = true
	}

	// Handle destination operand
	if dst, ok := inst.Dst.(*Pseudo); ok {
		inst.Dst = sa.allocateVar(dst.Identifier)
		modified = true
	}

	if modified {
		g.Program.Function.Instructions[index] = inst
	}

	// Handle stack-to-stack moves
	if _, srcIsStack := inst.Src.(*Stack); srcIsStack {
		if _, dstIsStack := inst.Dst.(*Stack); dstIsStack {
			// Replace with two instructions using temporary register
			g.Program.Function.Instructions[index] = &Mov{
				Src: inst.Src,
				Dst: &Reg{Reg: regR10},
			}

			g.Program.Function.Instructions = append(
				g.Program.Function.Instructions[:index+1],
				append(
					[]Instruction{
						&Mov{
							Src: &Reg{Reg: regR10},
							Dst: inst.Dst,
						},
					},
					g.Program.Function.Instructions[index+1:]...,
				)...,
			)
			return 1 // Indicate that extra instruction was added
		}
	}

	return 0
}

func (g *AsmGenerator) fixUnaryInstruction(inst *Unary, index int, sa *stackAllocator) {
	// Replace pseudoregisters
	if operand, ok := inst.Operand.(*Pseudo); ok {
		inst.Operand = sa.allocateVar(operand.Identifier)
		g.Program.Function.Instructions[index] = inst
	}
}

func (g *AsmGenerator) fixBinaryInstruction(inst *Binary, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	if operand1, ok := inst.Operand1.(*Pseudo); ok {
		inst.Operand1 = sa.allocateVar(operand1.Identifier)
		g.Program.Function.Instructions[index] = inst
	}

	if operand2, ok := inst.Operand2.(*Pseudo); ok {
		inst.Operand2 = sa.allocateVar(operand2.Identifier)
		g.Program.Function.Instructions[index] = inst
	}

	// Can't have mem address as both src and dst
	if _, dstIsOp := inst.Operand2.(*Stack); dstIsOp {
		if inst.Operator == opMult {
			// imul cant have mem address as dst, reguardless of source
			g.Program.Function.Instructions[index] = &Mov{
				Src: inst.Operand2,
				Dst: &Reg{Reg: regR11},
			}

			g.Program.Function.Instructions = append(
				g.Program.Function.Instructions[:index+1],
				append(
					[]Instruction{
						&Binary{
							Operator: inst.Operator,
							Operand1: inst.Operand1,
							Operand2: &Reg{Reg: regR11},
						},
						&Mov{Src: &Reg{Reg: regR11}, Dst: inst.Operand2},
					},
					g.Program.Function.Instructions[index+1:]...,
				)...,
			)
			return 2
		}

		if _, srcIsOp := inst.Operand1.(*Stack); srcIsOp {
			g.Program.Function.Instructions[index] = &Mov{
				Src: inst.Operand1,
				Dst: &Reg{Reg: regR10},
			}

			g.Program.Function.Instructions = append(
				g.Program.Function.Instructions[:index+1],
				append(
					[]Instruction{
						&Binary{
							Operator: inst.Operator,
							Operand1: &Reg{Reg: regR10},
							Operand2: inst.Operand2,
						},
					},
					g.Program.Function.Instructions[index+1:]...,
				)...,
			)
			return 1
		}
	}

	return 0
}

func (g *AsmGenerator) fixIdivInstruction(inst *Idiv, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	if operand, ok := inst.Operand.(*Pseudo); ok {
		inst.Operand = sa.allocateVar(operand.Identifier)
		g.Program.Function.Instructions[index] = inst
		return 0
	}

	// idivl can't operate on constants, copy value into scratch register
	if constant, ok := inst.Operand.(*Imn); ok {
		g.Program.Function.Instructions[index] = &Mov{
			Src: constant,
			Dst: &Reg{Reg: regR10},
		}

		g.Program.Function.Instructions = append(
			g.Program.Function.Instructions[:index+1],
			append(
				[]Instruction{
					&Idiv{
						Operand: &Reg{Reg: regR10},
					},
				},
				g.Program.Function.Instructions[index+1:]...,
			)...,
		)
		return 1
	}

	return 0
}

func (g *AsmGenerator) fixCmpInstruction(inst *Cmp, index int, sa *stackAllocator) int {
	// Replace pseudoregisters
	if operand1, ok := inst.Operand1.(*Pseudo); ok {
		inst.Operand1 = sa.allocateVar(operand1.Identifier)
		g.Program.Function.Instructions[index] = inst
	}

	if operand2, ok := inst.Operand2.(*Pseudo); ok {
		inst.Operand2 = sa.allocateVar(operand2.Identifier)
		g.Program.Function.Instructions[index] = inst
	}

	// Can't have mem address as both src and dst
	if _, dstIsOp := inst.Operand2.(*Stack); dstIsOp {

		if _, srcIsOp := inst.Operand1.(*Stack); srcIsOp {
			g.Program.Function.Instructions[index] = &Mov{
				Src: inst.Operand1,
				Dst: &Reg{Reg: regR10},
			}

			g.Program.Function.Instructions = append(
				g.Program.Function.Instructions[:index+1],
				append(
					[]Instruction{
						&Cmp{
							Operand1: &Reg{Reg: regR10},
							Operand2: inst.Operand2,
						},
					},
					g.Program.Function.Instructions[index+1:]...,
				)...,
			)
			return 1
		}
	} else if _, dstIsConst := inst.Operand2.(*Imn); dstIsConst {
		// cmp cant have mem address as dst
		g.Program.Function.Instructions[index] = &Mov{
			Src: inst.Operand2,
			Dst: &Reg{Reg: regR11},
		}

		g.Program.Function.Instructions = append(
			g.Program.Function.Instructions[:index+1],
			append(
				[]Instruction{
					&Cmp{
						Operand1: inst.Operand1,
						Operand2: &Reg{Reg: regR11},
					},
				},
				g.Program.Function.Instructions[index+1:]...,
			)...,
		)
		return 1
	}

	return 0
}

func (g *AsmGenerator) fixSetCCInstruction(inst *SetCC, index int, sa *stackAllocator) {
	// Replace pseudoregister
	if operand, ok := inst.Operand.(*Pseudo); ok {
		inst.Operand = sa.allocateVar(operand.Identifier)
		g.Program.Function.Instructions[index] = inst
	}
}
