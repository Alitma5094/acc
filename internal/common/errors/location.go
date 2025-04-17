package errors

import "fmt"

type Location struct {
	Line   int
	Column int
	File   string
}

func (l Location) String() string {
	if l.File != "" {
		return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Column)
	}
	return fmt.Sprintf("line %d, column %d", l.Line, l.Column)
}

func NewLocation(line, column int, file string) Location {
	return Location{
		Line:   line,
		Column: column,
		File:   file,
	}
}
