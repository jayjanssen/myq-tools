package myqlib

import (
	"fmt"
	"sort"
)

type UnitsDef map[float64]string

var (
	NumberUnits UnitsDef = UnitsDef{
		1:          ``,
		1000:       `k`,
		1000000:    `m`,
		1000000000: `g`,
	}
	MemoryUnits UnitsDef = UnitsDef{
		1:             `b`,
		1024:          `K`,
		1048576:       `M`,
		1073741824:    `G`,
		1099511627776: `T`,
	}
	SecondUnits UnitsDef = UnitsDef{
		1000:        `ks`,
		1:           `s`,
		0.001:       `ms`,
		0.000001:    `µs`,
		0.000000001: `ns`,
	}
	MicroSecondUnits UnitsDef = UnitsDef{
		1000000000: `ks`,
		1000000:    `s`,
		1000:       `ms`,
		1:          `µs`,
	}
	NanoSecondUnits UnitsDef = UnitsDef{
		1000000000: `s`,
		1000000:    `ms`,
		1000:       `µs`,
		1:          `ns`,
	}
	PercentUnits UnitsDef = UnitsDef{
		1: `%`,
	}
)

func collapse_number(value float64, width int64, precision int64, units UnitsDef) string {
	// To store the keys in slice in sorted order
	var factors []float64
	for k := range units {
		factors = append(factors, k)
	}
	sort.Float64s(factors)

	for _, factor := range factors {
		unit := units[factor]
		raw := value / factor
		str := fmt.Sprintf(fmt.Sprint(`%.`, precision, `f%s`), raw, unit)

		if raw != 0 && int64(len(str)) <= width+precision {
			left := width - int64(len(str))
			if left < 0 {
				if precision > 0 {
					return collapse_number(value, width, precision-1, units)
				} else {
					return str
				}
			} else if left > 1 && factor != 1 {
				dec := left - 1
				return fmt.Sprintf(fmt.Sprint(`%.`, dec, `f%s`), raw, unit)
			} else {
				return str
			}
		}
	}
	str := fmt.Sprintf(fmt.Sprint(`%.`, precision, `f`), value)
	if int64(len(str)) <= width && precision > 0 {
		return collapse_number(value, width, precision-1, units)
	} else {
		return str
	}
}
