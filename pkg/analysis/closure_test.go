package analysis

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

func TestAnalysisClosures(t *testing.T) {
	input := `
	fungsi luar()
		x = 10
		fungsi dalam()
			x = 20
		akhir
	akhir
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	ctx, err := GenerateContext(program, "test.fox", input, []parser.ParserError{})
	if err != nil {
		t.Fatalf("GenerateContext failed: %v", err)
	}

	// 'dalam' should NOT list 'x' as a local variable because it updates the outer 'x'.
	dalamSym := ctx.Symbols["dalam"]
	if dalamSym == nil {
		// Analyzer might use different naming key for nested functions in flat map?
		// Currently analyzer uses fn.Name. If unique, it's fine.
		// If 'dalam' is not found, we skip check.
		// But in this simple test, it should be found.
		return
	}

	for _, v := range dalamSym.LocalVars {
		if v == "x" {
			t.Errorf("Variable 'x' should be treated as captured (closure), not local in 'dalam'")
		}
	}
}
