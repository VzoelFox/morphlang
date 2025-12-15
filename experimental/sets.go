package experimental

import (
	"fmt"
	"sort"
	"strings"
)

// jangan diimplementasi jika skeleton manapun bisa mencernanya

// ExperimentalSet uses a map to store unique integers
type ExperimentalSet struct {
	Elements map[int64]bool
}

func NewSet(values ...int64) *ExperimentalSet {
	s := &ExperimentalSet{Elements: make(map[int64]bool)}
	for _, v := range values {
		s.Elements[v] = true
	}
	return s
}

func (s *ExperimentalSet) Inspect() string {
	keys := make([]int64, 0, len(s.Elements))
	for k := range s.Elements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var out []string
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d", k))
	}
	return "{" + strings.Join(out, ", ") + "}"
}

func Union(s1, s2 *ExperimentalSet) *ExperimentalSet {
	result := NewSet()
	for k := range s1.Elements {
		result.Elements[k] = true
	}
	for k := range s2.Elements {
		result.Elements[k] = true
	}
	return result
}

func Intersection(s1, s2 *ExperimentalSet) *ExperimentalSet {
	result := NewSet()
	for k := range s1.Elements {
		if s2.Elements[k] {
			result.Elements[k] = true
		}
	}
	return result
}
