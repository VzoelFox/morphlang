package object

import "testing"

func TestEnvironmentBasic(t *testing.T) {
	env := NewEnvironment()
	val := &Integer{Value: 10}

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
	outerVal := &Integer{Value: 5}
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

func TestEnvironmentShadowing(t *testing.T) {
	outer := NewEnvironment()
	val1 := &Integer{Value: 1}
	outer.Set("x", val1)

	inner := NewEnclosedEnvironment(outer)
	val2 := &Integer{Value: 2}
	inner.Set("x", val2) // Shadows outer x

	// Inner should see val2
	got, ok := inner.Get("x")
	if !ok {
		t.Fatalf("inner.Get('x') failed")
	}
	if got.(*Integer).Value != 2 {
		t.Errorf("inner.Get('x') should be 2, got %s", got.Inspect())
	}

	// Outer should still see val1
	gotOuter, ok := outer.Get("x")
	if !ok {
		t.Fatalf("outer.Get('x') failed")
	}
	if gotOuter.(*Integer).Value != 1 {
		t.Errorf("outer.Get('x') should be 1, got %s", gotOuter.Inspect())
	}
}

func TestEnvironmentDeepNesting(t *testing.T) {
	global := NewEnvironment()
	global.Set("g", &Integer{Value: 100})

	middle := NewEnclosedEnvironment(global)
	middle.Set("m", &Integer{Value: 50})

	local := NewEnclosedEnvironment(middle)
	local.Set("l", &Integer{Value: 10})

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
