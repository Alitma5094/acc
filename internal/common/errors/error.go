package errors

import "fmt"

type CompilerError struct {
	Message  string
	Location Location
	Phase    CompilationPhase
}

type CompilationPhase int

const (
	LexPhase CompilationPhase = iota
	ParsePhase
	IRGenPhase
	CodeGenPhase
)

func (p CompilationPhase) String() string {
	switch p {
	case LexPhase:
		return "Lexical Analysis"
	case ParsePhase:
		return "Syntax Analysis"
	case IRGenPhase:
		return "IR Generation"
	case CodeGenPhase:
		return "Code Generation"
	default:
		return "Unknown Phase"
	}
}

// Impliment the error interface
func (e *CompilerError) Error() string {
	return fmt.Sprintf("%s error at %s: %s", e.Phase.String(), e.Location.String(), e.Message)
}

func NewLexError(msg string, loc Location) *CompilerError {
	return &CompilerError{
		Message:  msg,
		Location: loc,
		Phase:    LexPhase,
	}
}

func NewParseError(msg string, loc Location) *CompilerError {
	return &CompilerError{
		Message:  msg,
		Location: loc,
		Phase:    ParsePhase,
	}
}

func NewIRGenError(msg string) *CompilerError {
	return &CompilerError{
		Message: msg,
		Phase:   IRGenPhase,
	}
}

func NewCodeGenError(msg string) *CompilerError {
	return &CompilerError{
		Message: msg,
		Phase:   CodeGenPhase,
	}
}
