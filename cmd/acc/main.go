package main

import (
	"acc/internal/codegen"
	"acc/internal/common/config"
	"acc/internal/ir"
	"acc/internal/lexer"
	"acc/internal/parser"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command line flags
	cfg := config.NewCompilerConfig()
	cfg.RegisterFlags()
	flag.Parse()

	// Get input file path
	inputFile := flag.Arg(0)
	if inputFile == "" {
		log.Fatal("Must provide a file path")
	}

	// Preprocess with GCC
	preprocessedSource, err := preprocess(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	// Run the compiler pipeline
	if err := runCompiler(preprocessedSource, inputFile, cfg); err != nil {
		log.Fatal(err)
	}
}

func preprocess(inputFile string) (string, error) {
	basePath := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
	outputFile := fmt.Sprintf("%s.i", basePath)

	cmd := exec.Command("gcc", "-E", "-P", inputFile, "-o", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("preprocessing failed: %v\nOutput: %s", err, string(output))
	}

	source, err := os.ReadFile(outputFile)
	if err != nil {
		os.Remove(outputFile)
		return "", err
	}
	os.Remove(outputFile)

	return string(source), nil
}

func runCompiler(source, inputFile string, cfg *config.CompilerConfig) error {
	// Create lexer
	l := lexer.NewLexer(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return err
	}

	if cfg.StopAfterLexing {
		return nil // Exit after lexical analysis
	}

	// Create parser
	p := parser.NewParser(tokens)
	ast, err := p.Parse()
	if err != nil {
		return err
	}
	fmt.Println(ast)

	if cfg.StopAfterParsing {
		return nil // Exit after parsing
	}

	// Generate TAC
	tacGen := ir.NewTACGenerator()
	tacProgram, err := tacGen.Generate(ast)
	if err != nil {
		return err
	}
	fmt.Println(tacProgram)

	if cfg.StopAfterTAC {
		return nil // Exit after TAC generation
	}

	// Generate assembly
	asmGen := codegen.NewASMGenerator()
	err = asmGen.Generate(tacProgram)
	if err != nil {
		return err
	}

	asmGen.FixInstructions()

	if cfg.StopAfterCodeGen {
		return nil // Exit before assembly file creation
	}

	// Write assembly to file and assemble
	basePath := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
	asmFile := fmt.Sprintf("%s.s", basePath)

	if err := os.WriteFile(asmFile, []byte(asmGen.Program.EmitAsm()), 0644); err != nil {
		return err
	}

	// Assemble with GCC
	cmd := exec.Command("gcc", asmFile, "-o", basePath)
	output, err := cmd.CombinedOutput()
	os.Remove(asmFile) // Clean up

	if err != nil {
		return fmt.Errorf("assembly failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}
