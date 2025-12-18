package object

import (
	"testing"
)

func TestSystemBuiltins(t *testing.T) {
	// Check info_memori
	idx := GetBuiltinByName("info_memori")
	if idx == -1 {
		t.Fatal("info_memori builtin not found")
	}

	memFn := Builtins[idx].Builtin.Fn
	res := memFn()

	if isError(res) {
		t.Fatalf("info_memori returned error: %s", res.Inspect())
	}

	hashObj, ok := res.(*Hash)
	if !ok {
		t.Fatalf("info_memori should return Hash, got %T", res)
	}

	// Check keys presence
	foundTotal := false
	for _, pair := range hashObj.GetPairs() {
		keyStr := pair.Key.(*String).GetValue()
		if keyStr == "total" {
			foundTotal = true
			val := pair.Value.(*Integer).GetValue()
			if val <= 0 {
				t.Errorf("Memory total should be > 0, got %d", val)
			}
		}
	}
	if !foundTotal {
		t.Error("info_memori result missing 'total' key")
	}

	// Check info_cpu
	idxCpu := GetBuiltinByName("info_cpu")
	if idxCpu == -1 {
		t.Fatal("info_cpu builtin not found")
	}
	cpuFn := Builtins[idxCpu].Builtin.Fn
	resCpu := cpuFn()

	if isError(resCpu) {
		t.Fatalf("info_cpu returned error: %s", resCpu.Inspect())
	}

	_, ok = resCpu.(*Integer) // Expect Integer percent
	if !ok {
		t.Fatalf("info_cpu should return Integer, got %T", resCpu)
	}
}

func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}
