package viewer

import (
	"testing"

	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// Funcs to get some test columns
func getTestCol() defaultCol {
	return defaultCol{
		Name:        "cons",
		Description: "Connections per second",
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

func TestColGetDetailedHelp(t *testing.T) {
	col := getTestCol()

	help := col.GetDetailedHelp()
	if len(help) != 1 {
		t.Errorf("detailed help unexpected line count: %d", len(help))
	}

	if help[0] != "cons: Connections per second" {
		t.Error("bad detailed help")
	}
}

func getTestCache() *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)
	return cache
}

func TestColGetHeader(t *testing.T) {
	col := getTestCol()
	cache := getTestCache()
	headers := col.GetHeader(cache)

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

	// Test too long name
	col.Name = "consss"
	headers = col.GetHeader(cache)
	// Expect one line header
	if len(headers) != 1 {
		t.Errorf("Header more than 1 line: %d", len(headers))
	}

	header = headers[0]
	if len(header) != col.Length {
		t.Errorf("Got header of length: %d, expected: %d", len(header), col.Length)
	}

	if header != "cons" {
		t.Errorf("Expected header to be truncated to 'cons', not: %s", header)
	}
}
