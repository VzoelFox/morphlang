package evaluator

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestStringConcatenation(t *testing.T) {
	input := `"Halo" + " " + "Dunia!"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.GetValue() != "Halo Dunia!" {
		t.Errorf("String has wrong value. got=%q", str.GetValue())
	}
}

func TestStringComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"a" == "a"`, true},
		{`"a" != "b"`, true},
		{`"a" == "b"`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}
