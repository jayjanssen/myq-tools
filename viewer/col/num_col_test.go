package col

import (
	"testing"
)

// Funcs to get some test columns
func getTestNumCol() numCol {
	sources := []string{"status"}
	return numCol{
		defaultCol: defaultCol{
			Name:        "cons",
			Description: "Connections per second",
			Sources:     sources,
			Type:        RATE,
			Length:      4,
		},
		Units:     NUMBER,
		Precision: 0,
	}
}

func TestNumColGetData(t *testing.T) {
	col := getTestCol()
	state := getTestState()

	datas := col.GetData(state)

	if len(datas) != 1 {
		t.Errorf("Header more than 1 line: %d", len(datas))
	}

	data := datas[0]
	if len(data) != col.Length {
		t.Errorf("Got data of length: %d, expected: '%d'", len(data), col.Length)
	}

	if data != "   5" {
		t.Errorf("Expected data to be '   5', not: '%s'", data)
	}

}
