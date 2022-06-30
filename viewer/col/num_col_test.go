package col

import "testing"

// Funcs to get some test columns
func getTestNumCol() numCol {
	sources := []string{"status"}
	return numCol{
		defaultCol: defaultCol{
			Name:        "cons",
			Description: "Connections per second",
			Sources:     sources,
			Length:      4,
		},
		Units:     NUMBER,
		Precision: 0,
	}
}

func TestNumColfitNumber(t *testing.T) {
	col := getTestNumCol()
	out := col.fitNumber(5, col.Precision)

	if len(out) != 4 || out != "   5" {
		t.Errorf("Unexpected fitNumber: '%s'", out)
	}
}
