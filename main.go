package main

import (
	"acc/acc"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

func main() {
	lex := flag.Bool("lex", false, "stop after lexing")
	parse := flag.Bool("parse", false, "stop after parsing")
	flag.Parse()

	inputFile := flag.Arg(0)
	if inputFile == "" {
		log.Print("Must provide a file path")
		os.Exit(1)
	}
	outputFile := fmt.Sprintf("%s.i", path.Base(inputFile))
	defer os.Remove(outputFile)

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

	lexer := acc.NewLexer(string(source))
	err = lexer.Lex()
	if err != nil {
		log.Println(err)
		os.Remove(outputFile)
		os.Exit(1)
	}

	log.Printf("%q\n", lexer.Tokens)

	if *lex {
		return
	}

	parser := acc.NewParser(lexer.Tokens)
	err = parser.Parse()

	if err != nil {
		log.Println(err)
		os.Remove(outputFile)
		os.Exit(1)
	}

	log.Printf("%q\n", parser.Tree)

	if *parse {
		return
	}
}
