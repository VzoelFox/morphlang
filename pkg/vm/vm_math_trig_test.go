package vm

import (
	"math"
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestTrigonometry(t *testing.T) {
	// Tolerance for float comparison
	epsilon := 0.000001

	tests := []struct {
		input    string
		expected float64
	}{
		{"sin(0)", 0.0},
		{"sin(pi()/2)", 1.0},
		{"cos(0)", 1.0},
		{"cos(pi())", -1.0},
		{"tan(0)", 0.0},
		// Inverse
		{"asin(0)", 0.0},
		{"acos(1)", 0.0},
		{"atan(0)", 0.0},
		// Pow/Sqrt with floats
		{"pow(2.0, 3.0)", 8.0},
		{"sqrt(4.0)", 2.0},
		{"sqrt(2)", math.Sqrt(2)},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		lastPop := vm.GetLastPopped()
		testFloatObjectApprox(t, lastPop, tt.expected, epsilon)
	}
}

func testFloatObjectApprox(t *testing.T, obj object.Object, expected float64, epsilon float64) {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return
	}

	if math.Abs(result.GetValue()-expected) > epsilon {
		t.Errorf("object has wrong value. got=%f, want=%f (epsilon=%f)", result.GetValue(), expected, epsilon)
	}
}
