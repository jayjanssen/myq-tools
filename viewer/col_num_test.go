package viewer

import (
	"testing"
)

// Funcs to get some test columns
func getTestcolNum(units UnitsType, precision, width int) colNum {
	return colNum{
		defaultCol: defaultCol{
			Name:   "test",
			Length: width,
		},
		Units:     units,
		Precision: precision,
	}
}

func TestNumbers(t *testing.T) {

	assert := func(test_name, expected string, units UnitsType, val float64, precision, width int) {
		col := getTestcolNum(units, precision, width)
		str := col.fitNumber(val, col.Precision)
		if str != expected {
			t.Errorf("%s err: `%s` != `%s`", test_name, str, expected)
		}
	}

	assert(`one is the loneliest number`, `1`, NUMBER, 1, 0, 3)
	assert(`one point oh`, `1.0`, NUMBER, 1, 1, 3)
	assert(`five hundred`, `500`, NUMBER, 500, 0, 3)

	assert(`one kay`, `1k`, NUMBER, 1000, 0, 3)
	assert(`one zero zero zero`, `1000`, NUMBER, 1000, 0, 4)

	assert(`round up to 1k`, `1k`, NUMBER, 501, 0, 2)
	assert(`cant fit 500 into two`, `##`, NUMBER, 500, 0, 2)
	assert(`twelve k`, `12k`, NUMBER, 12300, 0, 4)
	assert(`one twenty three k`, `123k`, NUMBER, 123000, 0, 4)

	assert(`point one em`, `.1m`, NUMBER, 123000, 0, 3)
	assert(`point six em`, `.6m`, NUMBER, 550000, 0, 3)

	assert(`twelve m`, `12m`, NUMBER, 12300000, 0, 4)
	assert(`twelve point three m`, `12.3m`, NUMBER, 12300000, 0, 5)

	assert(`three hundred kay`, `300k`, NUMBER, 300000, 0, 4)

	assert(`wayyy to big`, `####`, NUMBER, 3000000000000000, 0, 4)

	assert(`one bee`, `1b`, MEMORY, 1, 0, 3)
	assert(`one point nil`, `1b`, MEMORY, 1, 1, 3)
	assert(`one point oh`, `1.0b`, MEMORY, 1, 1, 4)

	assert(`five oh oh rounded down`, `.5K`, MEMORY, 500, 0, 3)
	assert(`five fifty fit`, `.5K`, MEMORY, 550, 0, 3)
	assert(`five fifty bee`, `550b`, MEMORY, 550, 0, 4)

	assert(`one kay`, `1K`, MEMORY, 1000, 0, 3)
	assert(`one zero zero zero bee`, `1000b`, MEMORY, 1000, 0, 5)
	assert(`one point oh kay`, `1.0K`, MEMORY, 1000, 0, 4)

	assert(`one oh oh one bee`, `1001b`, MEMORY, 1001, 0, 5)
	assert(`one poing oh kay`, `1.0K`, MEMORY, 1001, 0, 4)

	assert(`round up to one kay`, `1K`, MEMORY, 550, 0, 2)
	assert(`cant fit 500b into two`, `##`, MEMORY, 500, 0, 2)

	assert(`twelve kay`, `12K`, MEMORY, 12300, 0, 4)
	assert(`one twenty three kay`, `120K`, MEMORY, 123000, 0, 4)

	assert(`twelve em`, `12M`, MEMORY, 12300000, 0, 4)
	assert(`eleven point seven em`, `11.7M`, MEMORY, 12300000, 0, 5)

	assert(`zero en ess`, `0.0ns`, SECOND, 0, 0, 5)
	assert(`four seven five mu ess`, `476µs`, SECOND, 0.000476, 0, 5)

	assert(`zero en ess`, `0ns`, NANOSECOND, 0.000000, 0, 5)
	assert(`zero pee ess`, `0ps`, PICOSECOND, 0.000000, 0, 5)

}
