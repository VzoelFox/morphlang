package main

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/evaluator"
	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

func main() {
	input := `panjang("")`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	result := evaluator.Eval(program, env)
	fmt.Printf("Result: %s\n", result.Inspect())
}
