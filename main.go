package main

import (
	"log"
	"os"
)

func main() {
	sourceb, err := os.ReadFile("test.c.out")
	if err != nil {
		log.Fatal(err)
	}
	source := string(sourceb)
	lexer := NewLexer(source)
	err = lexer.lex()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q\n", lexer.tokens)
}
