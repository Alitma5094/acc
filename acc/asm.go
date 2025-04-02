package acc

// import (
// 	"fmt"
// )

// type Generator struct {
// 	ast  *nodeProgram
// 	Tree asmProgram
// }

// func NewGenerator(tree *nodeProgram) *Generator {
// 	return &Generator{ast: tree}
// }

// func (g *Generator) Generate() error {
// 	asmFunc, err := g.convertFunction(g.ast.function)
// 	if err != nil {
// 		return err
// 	}

// 	g.Tree = asmProgram{function: *asmFunc}
// 	return nil
// }
// func (g *Generator) Emit() string {
// 	return g.Tree.emitASM()
// }

// type asmNode interface {
// 	emitASM() string
// }

// type asmProgram struct {
// 	function asmFunction
// }

// func (p *asmProgram) emitASM() string {
// 	return p.function.emitASM()
// }

// type asmFunction struct {
// 	name         string
// 	instructions []asmInstruction
// }

// func (f *asmFunction) emitASM() string {
// 	i := ""
// 	for _, in := range f.instructions {
// 		i += in.emitASM()
// 	}
// 	return fmt.Sprintf("\t.global _%s\n_%s:\n%s", f.name, f.name, i)
// }

// type asmInstruction interface {
// 	asmNode
// }

// type asmMov struct {
// 	src asmOperand
// 	dst asmOperand
// }

// func (m *asmMov) emitASM() string {
// 	return fmt.Sprintf("\tmovl\t%s, %s\n", m.src.emitASM(), m.dst.emitASM())
// }

// type asmRet struct {
// }

// func (r *asmRet) emitASM() string {
// 	return "\tret"
// }

// type asmOperand interface {
// 	asmNode
// }

// type asmImn struct {
// 	value int
// }

// func (i *asmImn) emitASM() string {
// 	return fmt.Sprintf("$%d", i.value)
// }

// type asmRegister struct{}

// func (r *asmRegister) emitASM() string {
// 	return "%eax"
// }

// func (g *Generator) convertFunction(n *nodeFunction) (*asmFunction, error) {
// 	if n == nil {
// 		return nil, fmt.Errorf("function node is nil")
// 	}

// 	funcName := n.name.val

// 	instrs, err := g.convertStatement(n.body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &asmFunction{
// 		name:         funcName,
// 		instructions: instrs,
// 	}, nil
// }

// func (g *Generator) convertStatement(s *nodeStatement) ([]asmInstruction, error) {
// 	if s == nil {
// 		return nil, fmt.Errorf("statement node is nil")
// 	}

// 	movInstr, err := g.convertExpression(s.expression)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return []asmInstruction{
// 		movInstr,
// 		&asmRet{},
// 	}, nil
// }

// func (g *Generator) convertExpression(e *nodeExpression) (*asmMov, error) {
// 	// if e == nil || e.constant == nil {
// 	// 	return nil, fmt.Errorf("expression node is nil")
// 	// }

// 	imm := &asmImn{value: 999}
// 	// imm := &asmImn{value: e.constant.val}

// 	reg := &asmRegister{}

// 	return &asmMov{
// 		src: imm,
// 		dst: reg,
// 	}, nil
// }
