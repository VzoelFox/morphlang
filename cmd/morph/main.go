package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/VzoelFox/morphlang/pkg/analysis"
	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/parser"
	"github.com/VzoelFox/morphlang/pkg/vm"
)

func main() {
	// Panic Recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\n\033[31m[INTERNAL ERROR] Morph Compiler Crashed!\033[0m")
			fmt.Printf("Panic: %v\n", r)
			fmt.Println("Please report this issue with the code that caused it.")
			os.Exit(1)
		}
	}()

	debug := flag.Bool("debug", false, "Enable debug output")
	check := flag.Bool("check", false, "Check syntax only")
	useVM := flag.Bool("vm", false, "Run using Bytecode VM")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: morph [options] <file>")
		os.Exit(1)
	}

	cmd := args[0]
	filename := ""

	if cmd == "compile" {
		if len(args) < 2 {
			fmt.Println("Usage: morph compile <file>")
			os.Exit(1)
		}
		filename = args[1]
	} else {
		filename = cmd
	}

	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	input := string(content)

	// Lexer for Parsing
	l := lexer.New(input)

	// Parser
	p := parser.New(l)
	program := p.ParseProgram()

	// Debug Output
	if *debug {
		fmt.Printf("=== Morph Compiler Debug Output ===\n")
		fmt.Printf("File: %s\n", filename)
		fmt.Printf("Compiled: %s\n", time.Now().Format(time.RFC3339))

		fmt.Println("\n--- Lexer Output ---")
		l2 := lexer.New(input)
		for {
			tok := l2.NextToken()
			if tok.Type == lexer.EOF {
				break
			}
			fmt.Printf("%d:%d\t%s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
		}

		fmt.Println("\n--- Parser Output ---")
		if len(p.Errors()) > 0 {
			fmt.Println("Errors encountered:")
			for _, msg := range p.Errors() {
				fmt.Println(msg)
			}
		} else {
			fmt.Println("AST generated successfully")
			fmt.Println(program.String())
		}
	}

	// Analysis & Context Generation (Always generate context first)
	ctx, err := analysis.GenerateContext(program, filename, input, p.Errors())
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
	} else {
		// Write .fox.vz
		outPath := filename + ".vz"
		file, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating context file: %v\n", err)
		} else {
			enc := json.NewEncoder(file)
			enc.SetIndent("", "  ")
			if err := enc.Encode(ctx); err != nil {
				fmt.Printf("Error writing context: %v\n", err)
			}
			file.Close()

			if *debug {
				fmt.Printf("\n--- Context File ---\n")
				fmt.Printf("Generated: %s\n", outPath)
			}
		}
	}

	if len(p.Errors()) > 0 {
		fmt.Printf("Parsing failed with %d errors.\n", len(p.Errors()))
		for _, msg := range p.Errors() {
			fmt.Println(msg)
		}
		os.Exit(1)
	}

	// VM Execution
	if *useVM {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("Compilation failed:\n%s\n", err)
			os.Exit(1)
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Printf("Runtime error:\n%s\n", err)
			os.Exit(1)
		}
		return
	}

	if !*check {
		fmt.Printf("Successfully compiled %s\n", filename)
	}
}
