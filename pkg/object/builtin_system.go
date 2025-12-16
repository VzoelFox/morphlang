package object

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func init() {
	// Register System Builtins
	RegisterBuiltin("info_memori", func(args ...Object) Object {
		v, err := mem.VirtualMemory()
		if err != nil {
			return &Error{Message: fmt.Sprintf("gopsutil error: %v", err)}
		}

		pairs := make(map[HashKey]HashPair)

		// Keys to export
		data := map[string]int64{
			"total":    int64(v.Total),
			"terpakai": int64(v.Used),
			"bebas":    int64(v.Free),
			"persen":   int64(v.UsedPercent), // Rounded down/truncated
		}

		for k, val := range data {
			kObj := &String{Value: k}
			hKey := kObj.HashKey()
			pairs[hKey] = HashPair{Key: kObj, Value: &Integer{Value: val}}
		}

		return &Hash{Pairs: pairs}
	})

	RegisterBuiltin("info_cpu", func(args ...Object) Object {
		// Interval 0 means return immediately?
		// Note: First call might be inaccurate or require blocking.
		// For now we use 0 (non-blocking).
		percentages, err := cpu.Percent(0, false)
		if err != nil {
			return &Error{Message: fmt.Sprintf("gopsutil error: %v", err)}
		}

		if len(percentages) == 0 {
			return &Integer{Value: 0}
		}

		// Return aggregate percent as Integer
		return &Integer{Value: int64(math.Round(percentages[0]))}
	})
}
