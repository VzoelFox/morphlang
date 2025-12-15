package evaluator

import (
	"testing"
)

func TestWhileExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			`
			x = 0;
			selama x < 10
				x = x + 1;
			akhir
			x;
			`,
			10,
		},
		{
			`
			result = 0;
			i = 1;
			selama i < 5
				result = result + i;
				i = i + 1;
			akhir
			result;
			`,
			10, // 1 + 2 + 3 + 4 = 10
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}
