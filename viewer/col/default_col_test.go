package col

import (
	"fmt"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

// Funcs to get some test columns
func getTestCol() defaultCol {
	sources := []string{"status"}
	return defaultCol{
		Name:        "cons",
		Description: "Connections per second",
		Sources:     sources,
		Type:        RATE,
		Length:      4,
	}
}

func getBadTestCol() defaultCol {
	sources := []string{"fooey"}
	return defaultCol{
		Name:        "cons",
		Description: "Connections per second",
		Sources:     sources,
		Type:        RATE,
		Length:      4,
	}
}

func TestColGetShortHelp(t *testing.T) {
	col := getTestCol()

	help := col.GetShortHelp()
	if help != "cons: Connections per second" {
		t.Error("Bad short help!")
	}
}

func TestColGetSources(t *testing.T) {
	loader.LoadDefaultSources()

	col := getTestCol()
	sources, err := col.GetSources()

	if err != nil {
		t.Error(err)
	}

	fmt.Printf("sources: %v\n", sources)

	if len(sources) != 1 {
		t.Errorf("Got the wrong number of sources: %d", len(sources))
	}
}

func TestColGetSourcesErr(t *testing.T) {
	loader.LoadDefaultSources()

	bcol := getBadTestCol()
	_, err := bcol.GetSources()
	if err == nil {
		t.Error("Expected error!")
	}
}

// Test object that implements loader.StateReader
type testStateReader struct {
	rate float64
}

func (sr testStateReader) GetKeyCurPrev(source, key string) (string, string) {
	if source == "status" && key == "connections" {
		return "100", "105"
	}
	return "", ""
}

func (sr testStateReader) SecondsDiff() int64 {
	return 1
}

func getTestState() loader.StateReader {
	return testStateReader{
		rate: 10,
	}
}

func TestColGetHeader(t *testing.T) {
	col := getTestCol()
	state := getTestState()
	headers := col.GetHeader(state)

	// Expect one line header
	if len(headers) != 1 {
		t.Errorf("Header more than 1 line: %d", len(headers))
	}

	header := headers[0]
	if len(header) != col.Length {
		t.Errorf("Got header of length: %d, expected: %d", len(header), col.Length)
	}

	if header != "cons" {
		t.Errorf("Expected header to be 'cons', not: %s", header)
	}
}

func TestColGetData(t *testing.T) {
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
