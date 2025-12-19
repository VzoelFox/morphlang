package object

import "testing"

func TestEnvironmentBasic(t *testing.T) {
	env := NewEnvironment()
	val := NewInteger(10)

	env.Set("x", val)

	got, ok := env.Get("x")
	if !ok {
		t.Fatalf("Get('x') failed")
	}
	if got != val {
		t.Errorf("Get('x') wrong value. expected=%v, got=%v", val, got)
	}

	_, ok = env.Get("y")
	if ok {
		t.Errorf("Get('y') should fail")
	}
}

func TestEnvironmentEnclosed(t *testing.T) {
	outer := NewEnvironment()
	outerVal := NewInteger(5)
	outer.Set("outer", outerVal)

	inner := NewEnclosedEnvironment(outer)

	// Test reading from outer
	got, ok := inner.Get("outer")
	if !ok {
		t.Fatalf("inner.Get('outer') failed")
	}
	if got != outerVal {
		t.Errorf("inner.Get('outer') wrong value")
	}
}

func TestEnvironmentClosureUpdate(t *testing.T) {
	outer := NewEnvironment()
	val1 := NewInteger(1)
	outer.Set("x", val1)

	inner := NewEnclosedEnvironment(outer)
	val2 := NewInteger(2)
	inner.Set("x", val2) // Updates outer x (Closure semantics)

	// Inner should see val2
	got, ok := inner.Get("x")
	if !ok {
		t.Fatalf("inner.Get('x') failed")
	}
	if got.(*Integer).GetValue() != 2 {
		t.Errorf("inner.Get('x') should be 2, got %s", got.Inspect())
	}

	// Outer should ALSO see val2 (Updated)
	gotOuter, ok := outer.Get("x")
	if !ok {
		t.Fatalf("outer.Get('x') failed")
	}
	if gotOuter.(*Integer).GetValue() != 2 {
		t.Errorf("outer.Get('x') should be 2 (updated), got %s", gotOuter.Inspect())
	}
}

func TestEnvironmentDeepNesting(t *testing.T) {
	global := NewEnvironment()
	global.Set("g", NewInteger(100))

	middle := NewEnclosedEnvironment(global)
	middle.Set("m", NewInteger(50))

	local := NewEnclosedEnvironment(middle)
	local.Set("l", NewInteger(10))

	// Local should see all
	if _, ok := local.Get("g"); !ok {
		t.Error("local cant see global")
	}
	if _, ok := local.Get("m"); !ok {
		t.Error("local cant see middle")
	}
	if _, ok := local.Get("l"); !ok {
		t.Error("local cant see local")
	}

	// Middle cant see local
	if _, ok := middle.Get("l"); ok {
		t.Error("middle saw local variable")
	}
}
