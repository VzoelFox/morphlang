package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

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
		{"jika (benar) { 10 }", 10},
		{"jika (benar) { 10 } lainnya { 20 }", 10},
		{"jika (salah) { 10 } lainnya { 20 }", 20},
		{"jika (1 < 2) { 10 }", 10},
		{"jika (1 < 2) { 10 } lainnya { 20 }", 10},
		{"jika (1 > 2) { 10 } lainnya { 20 }", 20},
		{"jika (1 > 2) { 10 }", nil}, // Null
		{"jika (salah) { 10 }", nil}, // Null
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
	// Cannot use input string because Parser doesn't support arrays.
	// Construct bytecode manually.
	tests := []struct {
		instructions compiler.Instructions
		constants    []object.Object
		expected     []int64
	}{
		{
			concatInstructions(
				compiler.Make(compiler.OpLoadConst, 0), // 1
				compiler.Make(compiler.OpLoadConst, 1), // 2
				compiler.Make(compiler.OpArray, 2),
				compiler.Make(compiler.OpPop),
			),
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
			},
			[]int64{1, 2},
		},
		{
			concatInstructions(
				compiler.Make(compiler.OpLoadConst, 0), // 1
				compiler.Make(compiler.OpLoadConst, 1), // 2
				compiler.Make(compiler.OpAdd),          // 3
				compiler.Make(compiler.OpLoadConst, 2), // 4
				compiler.Make(compiler.OpLoadConst, 3), // 5
				compiler.Make(compiler.OpMul),          // 20
				compiler.Make(compiler.OpArray, 2),
				compiler.Make(compiler.OpPop),
			),
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
				&object.Integer{Value: 4},
				&object.Integer{Value: 5},
			},
			[]int64{3, 20},
		},
	}

	for _, tt := range tests {
		bytecode := &compiler.Bytecode{
			Instructions: tt.instructions,
			Constants:    tt.constants,
		}
		runVmTestWithBytecode(t, bytecode, tt.expected)
	}
}

func TestHashLiterals(t *testing.T) {
	// Construct bytecode manually for {1: 2, 3: 4}
	tests := []struct {
		instructions compiler.Instructions
		constants    []object.Object
		expected     map[object.HashKey]int64
	}{
		{
			concatInstructions(
				compiler.Make(compiler.OpLoadConst, 0), // 1
				compiler.Make(compiler.OpLoadConst, 1), // 2
				compiler.Make(compiler.OpLoadConst, 2), // 3
				compiler.Make(compiler.OpLoadConst, 3), // 4
				compiler.Make(compiler.OpHash, 4),      // 4 elements (2 pairs)
				compiler.Make(compiler.OpPop),
			),
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
				&object.Integer{Value: 3},
				&object.Integer{Value: 4},
			},
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 3}).HashKey(): 4,
			},
		},
	}

	for _, tt := range tests {
		bytecode := &compiler.Bytecode{
			Instructions: tt.instructions,
			Constants:    tt.constants,
		}
		runVmTestWithBytecode(t, bytecode, tt.expected)
	}
}

func TestIndexExpressions(t *testing.T) {
	// [1, 2, 3][1] -> 2
	tests := []struct {
		instructions compiler.Instructions
		constants    []object.Object
		expected     interface{}
	}{
		{
			concatInstructions(
				compiler.Make(compiler.OpLoadConst, 0), // 1
				compiler.Make(compiler.OpLoadConst, 1), // 2
				compiler.Make(compiler.OpLoadConst, 2), // 3
				compiler.Make(compiler.OpArray, 3),
				compiler.Make(compiler.OpLoadConst, 3), // Index 1
				compiler.Make(compiler.OpIndex),
				compiler.Make(compiler.OpPop),
			),
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
				&object.Integer{Value: 3},
				&object.Integer{Value: 1}, // Index
			},
			2,
		},
		{
			// {1: 2}[1] -> 2
			concatInstructions(
				compiler.Make(compiler.OpLoadConst, 0), // Key 1
				compiler.Make(compiler.OpLoadConst, 1), // Value 2
				compiler.Make(compiler.OpHash, 2),
				compiler.Make(compiler.OpLoadConst, 0), // Key 1
				compiler.Make(compiler.OpIndex),
				compiler.Make(compiler.OpPop),
			),
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
			},
			2,
		},
	}

	for _, tt := range tests {
		bytecode := &compiler.Bytecode{
			Instructions: tt.instructions,
			Constants:    tt.constants,
		}
		runVmTestWithBytecode(t, bytecode, tt.expected)
	}
}

func concatInstructions(s ...[]byte) []byte {
	var out []byte
	for _, b := range s {
		out = append(out, b...)
	}
	return out
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
	default:
		t.Fatalf("unsupported test type")
	}
}

func runVmTest(t *testing.T, input string, expected interface{}) {
	program := parse(input)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	bytecode := comp.Bytecode()
	runVmTestWithBytecode(t, bytecode, expected)
}

func runVmTestWithBytecode(t *testing.T, bytecode *compiler.Bytecode, expected interface{}) {
	vm := New(bytecode)
	err := vm.Run()
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
		if len(result.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(result.Elements))
			return
		}
		for i, expectedVal := range expected {
			testIntegerObject(t, result.Elements[i], expectedVal)
		}
	case map[object.HashKey]int64:
		result, ok := obj.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", obj, obj)
			return
		}
		if len(result.Pairs) != len(expected) {
			t.Errorf("wrong num of pairs. want=%d, got=%d", len(expected), len(result.Pairs))
			return
		}
		for key, expectedVal := range expected {
			pair, ok := result.Pairs[key]
			if !ok {
				t.Errorf("no pair for key %v", key)
				continue
			}
			testIntegerObject(t, pair.Value, expectedVal)
		}
	case nil:
		if obj == nil {
			return
		}
		if obj.Type() != object.NULL_OBJ {
			t.Errorf("object is not Null. got=%T (%+v)", obj, obj)
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

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
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

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
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

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
	}
}
