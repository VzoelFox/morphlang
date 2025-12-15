package object

import (
	"testing"
)

func TestScopeUpdate(t *testing.T) {
	globalEnv := NewEnvironment()
	globalEnv.Set("x", &Integer{Value: 10})

	firstLocalEnv := NewEnclosedEnvironment(globalEnv)
	// This should update x in globalEnv
	firstLocalEnv.Set("x", &Integer{Value: 20})

	// Verify globalEnv updated
	obj, ok := globalEnv.Get("x")
	if !ok {
		t.Fatalf("x not found in globalEnv")
	}
	if val := obj.(*Integer).Value; val != 20 {
		t.Errorf("expected global x to be 20, got %d", val)
	}

	// Verify firstLocalEnv sees update (via look up)
	objLocal, ok := firstLocalEnv.Get("x")
	if !ok {
		t.Fatalf("x not found in firstLocalEnv")
	}
	if val := objLocal.(*Integer).Value; val != 20 {
		t.Errorf("expected local x to be 20, got %d", val)
	}

	// New variable in local should NOT affect global
	firstLocalEnv.Set("y", &Integer{Value: 50})
	_, ok = globalEnv.Get("y")
	if ok {
		t.Errorf("variable y should not exist in globalEnv")
	}
}

func TestScopeShadowingWithDeclaration(t *testing.T) {
	// Note: Morph doesn't have explicit 'var', so strictly speaking we can't force shadowing
	// of an existing variable unless we define syntax for it.
	// With current "Set = Update or Create", strict shadowing of existing outer variable is HARDER
	// unless we introduce `local x = ...` or similar.
	// But let's verify deep nesting.

	global := NewEnvironment()
	global.Set("g", &Integer{Value: 1})

	mid := NewEnclosedEnvironment(global)
	mid.Set("m", &Integer{Value: 2})
	mid.Set("g", &Integer{Value: 10}) // Should update global

	inner := NewEnclosedEnvironment(mid)
	inner.Set("i", &Integer{Value: 3})
	inner.Set("m", &Integer{Value: 20}) // Should update mid
	inner.Set("g", &Integer{Value: 100}) // Should update global

	// Check inner
	val, _ := inner.Get("g")
	if val.(*Integer).Value != 100 { t.Errorf("inner g wrong") }

	// Check mid
	val, _ = mid.Get("m")
	if val.(*Integer).Value != 20 { t.Errorf("mid m wrong") }
	val, _ = mid.Get("g")
	if val.(*Integer).Value != 100 { t.Errorf("mid g wrong (should see global update)") }

	// Check global
	val, _ = global.Get("g")
	if val.(*Integer).Value != 100 { t.Errorf("global g wrong") }
}
