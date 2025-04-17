package config

import "flag"

type CompilerConfig struct {
	StopAfterLexing  bool
	StopAfterParsing bool
	StopAfterTAC     bool
	StopAfterCodeGen bool
}

func NewCompilerConfig() *CompilerConfig {
	return &CompilerConfig{}
}

func (c *CompilerConfig) RegisterFlags() {
	flag.BoolVar(&c.StopAfterLexing, "lex", false, "stop after lexing")
	flag.BoolVar(&c.StopAfterParsing, "parse", false, "stop after parsing")
	flag.BoolVar(&c.StopAfterTAC, "tacky", false, "stop after TAC generation")
	flag.BoolVar(&c.StopAfterCodeGen, "codegen", false, "stop before code emission")
}
