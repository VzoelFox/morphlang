package evaluator

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"benar", true},
		{"salah", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"benar == benar", true},
		{"salah == salah", true},
		{"benar == salah", false},
		{"benar != salah", true},
		{"(1 < 2) == benar", true},
		{"(1 < 2) == salah", false},
		{"(1 > 2) == benar", false},
		{"(1 > 2) == salah", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!benar", false},
		{"!salah", true},
		{"!5", false},
		{"!!benar", true},
		{"!!salah", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"jika (benar) 10 akhir", 10},
		{"jika (salah) 10 akhir", nil},
		{"jika (1) 10 akhir", 10},
		{"jika (1 < 2) 10 akhir", 10},
		{"jika (1 > 2) 10 akhir", nil},
		{"jika (1 > 2) 10 lainnya 20 akhir", 20},
		{"jika (1 < 2) 10 lainnya 20 akhir", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"kembalikan 10;", 10},
		{"kembalikan 10; 9;", 10},
		{"kembalikan 2 * 5; 9;", 10},
		{"9; kembalikan 2 * 5; 9;", 10},
		{
			`
			jika (10 > 1)
				jika (10 > 1)
					kembalikan 10
				akhir
				kembalikan 1
			akhir
			`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestAssignmentStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 5; x;", 5},
		{"x = 5 * 5; x;", 25},
		{"x = 5; y = x; y;", 5},
		{"x = 5; y = x + 5; y;", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + benar;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + benar; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-benar",
			"unknown operator: -BOOLEAN",
		},
		{
			"benar + salah;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; benar + salah; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"jika (10 > 1) benar + salah akhir",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			jika (10 > 1)
				jika (10 > 1)
					kembalikan benar + salah
				akhir
				kembalikan 1
			akhir
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.GetMessage() != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, errObj.GetMessage())
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.GetValue(), expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.GetValue(), expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj.Type() != object.NULL_OBJ {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
