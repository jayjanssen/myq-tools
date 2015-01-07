package myqlib

import "testing"

func TestOne(t *testing.T) {

	assert := func(test_name, expected string, units UnitsDef, val float64, precision, width int64) {
		str := collapse_number(val, width, precision, units)
		if str != expected {
			t.Errorf("%s err: `%s` != `%s`", test_name, str, expected)
		}
	}

	assert(`one is the lonliest number`, `1`, NumberUnits, 1, 0, 3)
	assert(`one point oh`, `1.0`, NumberUnits, 1, 1, 3)
	assert(`five hundred`, `500`, NumberUnits, 500, 0, 3)

	assert(`one kay`, `1k`, NumberUnits, 1000, 0, 3)
	assert(`one zero zero zero`, `1000`, NumberUnits, 1000, 0, 4)

	assert(`round up to 1k`, `1k`, NumberUnits, 501, 0, 2)
	assert(`round down to 0k`, `0k`, NumberUnits, 500, 0, 2)
	assert(`twelve k`, `12k`, NumberUnits, 12300, 0, 4)
	assert(`one twenty three k`, `123k`, NumberUnits, 123000, 0, 4)

	assert(`twelve m`, `12m`, NumberUnits, 12300000, 0, 4)
	assert(`twelve point three m`, `12.3m`, NumberUnits, 12300000, 0, 5)

	assert(`one bee`, `1b`, MemoryUnits, 1, 0, 3)
	assert(`one point nil`, `1b`, MemoryUnits, 1, 1, 3)
	assert(`one point oh`, `1.0b`, MemoryUnits, 1, 1, 4)

	assert(`five oh oh rounded down`, `0K`, MemoryUnits, 500, 0, 3)
	assert(`five fifty rounded up`, `1K`, MemoryUnits, 550, 0, 3)
	assert(`five fifty bee`, `550b`, MemoryUnits, 550, 0, 4)

	assert(`one kay`, `1K`, MemoryUnits, 1000, 0, 3)
	assert(`one zero zero zero bee`, `1000b`, MemoryUnits, 1000, 0, 5)
	assert(`one point oh kay`, `1.0K`, MemoryUnits, 1000, 0, 4)

	assert(`one oh oh one bee`, `1001b`, MemoryUnits, 1001, 0, 5)
	assert(`one poing oh kay`, `1.0K`, MemoryUnits, 1001, 0, 4)

	assert(`round up to one kay`, `1K`, MemoryUnits, 550, 0, 2)
	assert(`round down to 0K`, `0K`, MemoryUnits, 500, 0, 2)

	assert(`twelve kay`, `12K`, MemoryUnits, 12300, 0, 4)
	assert(`one twenty three kay`, `120K`, MemoryUnits, 123000, 0, 4)

	assert(`twelve em`, `12M`, MemoryUnits, 12300000, 0, 4)
	assert(`eleven point seven em`, `11.7M`, MemoryUnits, 12300000, 0, 5)

}
