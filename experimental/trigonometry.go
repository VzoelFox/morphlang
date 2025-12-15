package experimental

import (
	"fmt"
	"math"
)

// jangan diimplementasi jika skeleton manapun bisa mencernanya

type ExperimentalFloat struct {
	Value float64
}

func (f *ExperimentalFloat) Inspect() string {
	return fmt.Sprintf("%f", f.Value)
}

func Sin(input *ExperimentalFloat) *ExperimentalFloat {
	return &ExperimentalFloat{Value: math.Sin(input.Value)}
}

func Cos(input *ExperimentalFloat) *ExperimentalFloat {
	return &ExperimentalFloat{Value: math.Cos(input.Value)}
}

func Tan(input *ExperimentalFloat) *ExperimentalFloat {
	return &ExperimentalFloat{Value: math.Tan(input.Value)}
}
