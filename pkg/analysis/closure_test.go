package analysis

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

func TestAnalyzerClosureSemantics(t *testing.T) {
	input := `
	x = 10

	fungsi update_x()
		x = 20 # Should update global x, NOT declare local x
	akhir

	fungsi local_y()
		y = 30 # Should declare local y
	akhir
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	ctx, err := GenerateContext(program, "closure_test.fox", input, []parser.ParserError{})
	if err != nil {
		t.Fatalf("GenerateContext failed: %v", err)
	}

	// 1. Check update_x
	updateXSym, ok := ctx.Symbols["update_x"]
	if !ok {
		t.Fatalf("Symbol update_x not found")
	}

	// In old semantics, x would be in LocalVars. In new semantics, it should NOT be.
	for _, v := range updateXSym.LocalVars {
		if v == "x" {
			t.Errorf("FAIL: 'x' found in update_x LocalVars. It should be treated as closure update.")
		}
	}

	// 2. Check local_y
	localYSym, ok := ctx.Symbols["local_y"]
	if !ok {
		t.Fatalf("Symbol local_y not found")
	}

	foundY := false
	for _, v := range localYSym.LocalVars {
		if v == "y" {
			foundY = true
			break
		}
	}
	if !foundY {
		t.Errorf("FAIL: 'y' NOT found in local_y LocalVars. It should be declared as local.")
	}
}

func TestAnalyzerNestedScope(t *testing.T) {
	input := `
	fungsi outer()
		a = 1
		fungsi inner()
			a = 2 # Update outer a
			b = 3 # Local inner b
		akhir
	akhir
	`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	ctx, err := GenerateContext(program, "nested_test.fox", input, []parser.ParserError{})
	if err != nil {
		t.Fatalf("GenerateContext failed: %v", err)
	}

	// Currently the analyzer does NOT recursively analyze inner functions into Symbols map
	// (it flattens them or ignores inner function literals if not assigned to top level names?).
	// Looking at analyzer.go:
	// It analyzes function declarations. But Morph only has function literals.
	// `analyzeFunction` walks the block. If it finds `FunctionLiteral`?
	// `walkExpression` -> `case *parser.FunctionLiteral: a.analyzeFunction(e)`
	// So it DOES analyze nested functions.

	// Check inner
	// But what is the name of inner function?
	// `fungsi inner()` -> FunctionLiteral with Name="inner".
	// Analyzer puts it in ctx.Symbols["inner"]. Note: it flattens names (doesn't do "outer.inner").

	innerSym, ok := ctx.Symbols["inner"]
	if !ok {
		t.Fatalf("Symbol inner not found")
	}

	// 'a' should NOT be local in inner
	for _, v := range innerSym.LocalVars {
		if v == "a" {
			t.Errorf("FAIL: 'a' found in inner LocalVars. Should be closure capture.")
		}
	}

	// 'b' should be local in inner
	foundB := false
	for _, v := range innerSym.LocalVars {
		if v == "b" {
			foundB = true
		}
	}
	if !foundB {
		t.Errorf("FAIL: 'b' not found in inner LocalVars.")
	}
}
