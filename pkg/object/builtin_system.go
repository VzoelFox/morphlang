package object

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func init() {
	RegisterBuiltin("info_memori", func(args ...Object) Object {
		v, err := mem.VirtualMemory()
		if err != nil {
			return NewError(fmt.Sprintf("gopsutil error: %v", err), ErrCodeRuntime, 0, 0)
		}

		data := map[string]int64{
			"total":    int64(v.Total),
			"terpakai": int64(v.Used),
			"bebas":    int64(v.Free),
			"persen":   int64(v.UsedPercent),
		}

		pairs := make([]HashPair, 0, len(data))
		for k, val := range data {
			kObj := NewString(k)
			pairs = append(pairs, HashPair{Key: kObj, Value: NewInteger(val)})
		}

		return NewHash(pairs)
	})

	RegisterBuiltin("info_cpu", func(args ...Object) Object {
		percentages, err := cpu.Percent(0, false)
		if err != nil {
			return NewError(fmt.Sprintf("gopsutil error: %v", err), ErrCodeRuntime, 0, 0)
		}

		if len(percentages) == 0 {
			return NewInteger(0)
		}

		return NewInteger(int64(math.Round(percentages[0])))
	})
}
