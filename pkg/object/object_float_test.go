package object

import "testing"

func TestFloatInspect(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1.5, "1.5"},
		{1.0, "1"},
		{0.0001, "0.0001"},
		{3.14159, "3.14159"},
	}

	for _, tt := range tests {
		obj := &Float{Value: tt.input}
		if obj.Inspect() != tt.expected {
			t.Errorf("wrong inspect output. want=%q, got=%q", tt.expected, obj.Inspect())
		}
	}
}
