package viewer

import (
	"fmt"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

// Funcs to get some test columns
func getTestCol() defaultCol {
	sources := []loader.SourceName{"status"}
	return defaultCol{
		Name:        "cons",
		Description: "Connections per second",
		Sources:     sources,
		Length:      4,
	}
}

func getBadTestCol() defaultCol {
	sources := []loader.SourceName{"fooey"}
	return defaultCol{
		Name:        "cons",
		Description: "Connections per second",
		Sources:     sources,
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

func TestColformatString(t *testing.T) {
	col := getTestCol()

	out := col.fitString("fooey")
	if len(out) != 4 && out != "fooe" {
		t.Errorf("truncated string improperly: '%s'", out)
	}

	out = col.fitString("f")
	if len(out) != 4 && out != "   f" {
		t.Errorf("padded string improperly: '%s'", out)
	}
}
