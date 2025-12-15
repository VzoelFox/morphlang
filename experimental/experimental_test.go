package experimental

import (
	"math"
	"testing"
)

func TestTrigonometry(t *testing.T) {
	val := &ExperimentalFloat{Value: math.Pi / 2} // 90 degrees

	// Sin(90) = 1
	resSin := Sin(val)
	if math.Abs(resSin.Value-1.0) > 0.0001 {
		t.Errorf("Sin(PI/2) wrong. expected=1, got=%f", resSin.Value)
	}

	// Cos(90) = 0 (approaching 0)
	resCos := Cos(val)
	if math.Abs(resCos.Value) > 0.0001 {
		t.Errorf("Cos(PI/2) wrong. expected=0, got=%f", resCos.Value)
	}
}

func TestSets(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet(3, 4, 5)

	// Union: {1, 2, 3, 4, 5}
	u := Union(s1, s2)
	if len(u.Elements) != 5 {
		t.Errorf("Union size wrong. expected=5, got=%d", len(u.Elements))
	}
	expectedUnion := "{1, 2, 3, 4, 5}"
	if u.Inspect() != expectedUnion {
		t.Errorf("Union inspect wrong. expected=%q, got=%q", expectedUnion, u.Inspect())
	}

	// Intersection: {3}
	i := Intersection(s1, s2)
	if len(i.Elements) != 1 {
		t.Errorf("Intersection size wrong. expected=1, got=%d", len(i.Elements))
	}
	if !i.Elements[3] {
		t.Errorf("Intersection should contain 3")
	}
}
