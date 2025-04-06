package main

import (
	"acc/acc"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	lex := flag.Bool("lex", false, "stop after lexing")
	parse := flag.Bool("parse", false, "stop after parsing")
	tac := flag.Bool("tacky", false, "stop after tack generation")
	codeGen := flag.Bool("codegen", false, "stop before code emmision")
	flag.Parse()

	inputFile := flag.Arg(0)
	if inputFile == "" {
		log.Print("Must provide a file path")
		os.Exit(1)
	}
	basePath := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
	outputFile := fmt.Sprintf("%s.i", basePath)

	cmd := exec.Command("gcc", "-E", "-P", inputFile, "-o", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command execution failed: %v\nOutput: %s", err, string(output))
		os.Exit(1)
	}

	source, err := os.ReadFile(outputFile)
	if err != nil {
		log.Println(err)
		os.Remove(outputFile)
		os.Exit(1)
	}
	os.Remove(outputFile)

	lexer := acc.NewLexer(string(source))
	err = lexer.Lex()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if *lex {
		return
	}

	parser := acc.NewParser(lexer.Tokens)
	err = parser.Parse()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if *parse {
		return
	}

	tacParser := acc.NewTacParser(parser.Tree)
	err = tacParser.Parse()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if *tac {
		return
	}

	generator := acc.NewAsmParser(tacParser.Tree)
	err = generator.Parse()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if *codeGen {
		return
	}

	inputFile = fmt.Sprintf("%s.s", basePath)
	os.WriteFile(inputFile, []byte(generator.Emit()), 0644)

	cmd = exec.Command("gcc", inputFile, "-o", basePath)
	output, err = cmd.CombinedOutput()

	os.Remove(inputFile)
	if err != nil {
		log.Printf("Command execution failed: %v\nOutput: %s", err, string(output))
		os.Exit(1)
	}

}
