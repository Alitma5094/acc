package acc

type asmNode interface {
	emitASM() string
}

type asmProgram struct {
	function asmFunction
}

func (p *asmProgram) emitASM() string {
	return ""
}

type asmFunction struct {
	name         string
	instructions []asmInstruction
}

func (f *asmFunction) emitASM() string {
	return ""
}

type asmInstruction interface {
	asmNode
}

type asmMov struct {
	src asmOperand
	dst asmOperand
}

func (m *asmMov) emitASM() string {
	return ""
}

type asmRet struct {
}

func (r *asmRet) emitASM() string {
	return ""
}

type asmOperand interface {
	asmNode
}

type asmImn struct {
	value int
}

func (i *asmImn) emitASM() string {
	return ""
}

type asmRegister struct{}

func (r *asmRegister) emitASM() string {
	return ""
}
