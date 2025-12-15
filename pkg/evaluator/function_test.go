package evaluator

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestFunctionObject(t *testing.T) {
	input := "fungsi(x) x + 2 akhir;"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"fungsi identitas(x) x akhir; identitas(5);", 5},
		{"fungsi identitas(x) kembalikan x; akhir; identitas(5);", 5},
		{"fungsi double(x) x * 2; akhir; double(5);", 10},
		{"fungsi add(x, y) x + y; akhir; add(5, 5);", 10},
		{"fungsi add(x, y) x + y; akhir; add(5 + 5, add(5, 5));", 20},
		{"fungsi(x) x; akhir(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	fungsi newAdder(x)
		fungsi(y) x + y akhir
	akhir

	addTwo = newAdder(2);
	addTwo(2);
	`

	testIntegerObject(t, testEval(input), 4)
}
