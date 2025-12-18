package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func parse(input string) parser.Node {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"50 + -50", 0},
		{"50 - -50", 100},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
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
		{"!benar", false},
		{"!salah", true},
		{"!5", false},
		{"!!benar", true},
		{"!!5", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"jika (benar) 10 akhir", 10},
		{"jika (benar) 10 lainnya 20 akhir", 10},
		{"jika (salah) 10 lainnya 20 akhir", 20},
		{"jika (1 < 2) 10 akhir", 10},
		{"jika (1 < 2) 10 lainnya 20 akhir", 10},
		{"jika (1 > 2) 10 lainnya 20 akhir", 20},
		{"jika (1 > 2) 10 akhir", nil}, // Null
		{"jika (salah) 10 akhir", nil}, // Null
	}

	runVmTests(t, tests)
}

func TestGlobalVariables(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"x = 10; x", 10},
		{"x = 10; y = 20; x + y", 30},
		{"x = 10; y = x + 5; y", 15},
		{"x = 10; x = 20; x", 20},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{"[]", []int64{}},
		{"[1, 2, 3]", []int64{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int64{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected map[object.HashKey]int64
	}{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				object.NewInteger(1).HashKey(): 2,
				object.NewInteger(2).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				object.NewInteger(2).HashKey(): 4,
				object.NewInteger(6).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", nil},
		{"[1, 2, 3][99]", nil},
		{"[1][-1]", nil},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[99]", nil},
		{"{}[0]", nil},
	}

	runVmTests(t, tests)
}

func TestWhileLoops(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"selama (salah) 10 akhir", nil},
		{"x = 0; selama (x < 3) x = x + 1 akhir; x", 3},
		{"x = 0; y = 0; selama (x < 3) x = x + 1; y = y + 2; akhir; y", 6},
		{"x = 0; selama (x < 3) x = x + 1; x; akhir", 3},
	}

	runVmTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"fungsi() kembalikan 5 + 10; akhir ();", 15},
		{"fungsi() 5 + 10; akhir ();", 15},
		{"fungsi() 1; 2; akhir ();", 2},
		{"fungsi() 1; kembalikan 2; akhir ();", 2},
		{"fungsi() kembalikan 1; 2; akhir ();", 1},
		{"fungsi(a) a; akhir (24);", 24},
		{"fungsi(a, b) a + b; akhir (1, 2);", 3},
		{"global = 10; fungsi() global; akhir ();", 10},
		{"global = 10; fungsi() global + 5; akhir ();", 15},
		{"global = 10; fungsi() local = 5; global + local; akhir ();", 15},
	}

	runVmTests(t, tests)
}

func TestFunctionNoReturnValue(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"fungsi() akhir ();", nil},
		{"fungsi() 1; akhir ();", 1},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests interface{}) {
	switch tests := tests.(type) {
	case []struct {
		input    string
		expected int64
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected bool
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected interface{}
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected string
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected []int64
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected []string
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []struct {
		input    string
		expected map[object.HashKey]int64
	}:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	case []vmTestCase:
		for _, tt := range tests {
			runVmTest(t, tt.input, tt.expected)
		}
	default:
		t.Fatalf("unsupported test type %T", tests)
	}
}

func runVmTest(t *testing.T, input string, expected interface{}) {
	program := parse(input)

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

	stackElem := vm.LastPoppedStackElem
	testExpectedObject(t, stackElem, expected)
}

func testExpectedObject(t *testing.T, obj object.Object, expected interface{}) {
	switch expected := expected.(type) {
	case int:
		testIntegerObject(t, obj, int64(expected))
	case int64:
		testIntegerObject(t, obj, expected)
	case float64:
		testFloatObject(t, obj, expected)
	case bool:
		testBooleanObject(t, obj, expected)
	case string:
		testStringObject(t, obj, expected)
	case []int64:
		result, ok := obj.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", obj, obj)
			return
		}
		elements := result.GetElements()
		if len(elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(elements))
			return
		}
		for i, expectedVal := range expected {
			testIntegerObject(t, elements[i], expectedVal)
		}
	case []string:
		result, ok := obj.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", obj, obj)
			return
		}
		elements := result.GetElements()
		if len(elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(elements))
			return
		}
		for i, expectedVal := range expected {
			testStringObject(t, elements[i], expectedVal)
		}
	case map[object.HashKey]int64:
		result, ok := obj.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", obj, obj)
			return
		}
		pairs := result.GetPairs()
		if len(pairs) != len(expected) {
			t.Errorf("wrong num of pairs. want=%d, got=%d", len(expected), len(pairs))
			return
		}

		pairsMap := make(map[object.HashKey]object.Object)
		for _, pair := range pairs {
			hashKey := pair.Key.(object.Hashable).HashKey()
			pairsMap[hashKey] = pair.Value
		}

		for key, expectedVal := range expected {
			val, ok := pairsMap[key]
			if !ok {
				t.Errorf("no pair for key %v", key)
				continue
			}
			testIntegerObject(t, val, expectedVal)
		}
	case nil:
		if obj == nil {
			return
		}
		if obj.Type() != object.NULL_OBJ {
			t.Errorf("object is not Null. got=%T (%+v)", obj, obj)
		}
	case *object.Error:
		result, ok := obj.(*object.Error)
		if !ok {
			t.Errorf("object is not Error. got=%T (%+v)", obj, obj)
			return
		}
		if result.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, result.Message)
		}
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) {
	if obj == nil {
		t.Errorf("object is nil, want Integer %d", expected)
		return
	}
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return
	}

	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.GetValue(), expected)
	}
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) {
	if obj == nil {
		t.Errorf("object is nil, want Float %f", expected)
		return
	}
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return
	}

	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.GetValue(), expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) {
	if obj == nil {
		t.Errorf("object is nil, want Boolean %t", expected)
		return
	}
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return
	}

	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.GetValue(), expected)
	}
}

func testStringObject(t *testing.T, obj object.Object, expected string) {
	if obj == nil {
		t.Errorf("object is nil, want String %q", expected)
		return
	}
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return
	}

	if result.GetValue() != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.GetValue(), expected)
	}
}
