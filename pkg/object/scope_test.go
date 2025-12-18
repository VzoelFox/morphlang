package object

import (
	"testing"
)

func TestScopeUpdate(t *testing.T) {
	globalEnv := NewEnvironment()
	globalEnv.Set("x", NewInteger(10))

	firstLocalEnv := NewEnclosedEnvironment(globalEnv)
	// This should update x in globalEnv
	firstLocalEnv.Set("x", NewInteger(20))

	// Verify globalEnv updated
	obj, ok := globalEnv.Get("x")
	if !ok {
		t.Fatalf("x not found in globalEnv")
	}
	if val := obj.(*Integer).GetValue(); val != 20 {
		t.Errorf("expected global x to be 20, got %d", val)
	}

	// Verify firstLocalEnv sees update (via look up)
	objLocal, ok := firstLocalEnv.Get("x")
	if !ok {
		t.Fatalf("x not found in firstLocalEnv")
	}
	if val := objLocal.(*Integer).GetValue(); val != 20 {
		t.Errorf("expected local x to be 20, got %d", val)
	}

	// New variable in local should NOT affect global
	firstLocalEnv.Set("y", NewInteger(50))
	_, ok = globalEnv.Get("y")
	if ok {
		t.Errorf("variable y should not exist in globalEnv")
	}
}

func TestScopeShadowingWithDeclaration(t *testing.T) {
	global := NewEnvironment()
	global.Set("g", NewInteger(1))

	mid := NewEnclosedEnvironment(global)
	mid.Set("m", NewInteger(2))
	mid.Set("g", NewInteger(10)) // Should update global

	inner := NewEnclosedEnvironment(mid)
	inner.Set("i", NewInteger(3))
	inner.Set("m", NewInteger(20)) // Should update mid
	inner.Set("g", NewInteger(100)) // Should update global

	// Check inner
	val, _ := inner.Get("g")
	if val.(*Integer).GetValue() != 100 { t.Errorf("inner g wrong") }

	// Check mid
	val, _ = mid.Get("m")
	if val.(*Integer).GetValue() != 20 { t.Errorf("mid m wrong") }
	val, _ = mid.Get("g")
	if val.(*Integer).GetValue() != 100 { t.Errorf("mid g wrong (should see global update)") }

	// Check global
	val, _ = global.Get("g")
	if val.(*Integer).GetValue() != 100 { t.Errorf("global g wrong") }
}
