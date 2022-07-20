package loader

import (
	"testing"
	"time"
)

// Helper functions

// create a FileParser and return the error
func newTestFileParser(fileName string) (*FileParser, error) {
	i, _ := time.ParseDuration("1s")
	f := NewFileParser(fileName)
	err := f.Initialize(i)
	return f, err
}

// create a FileParser and raise a test error if it fails
func newGoodFileParser(t testing.TB, fileName string) *FileParser {
	i, _ := time.ParseDuration("1s")
	f := NewFileParser(fileName)
	if err := f.Initialize(i); err != nil {
		t.Error(err)
	}
	return f
}

// create a FileParser with a 1m interval and raise a test error if it fails
func newGoodFile1mParser(t testing.TB, fileName string) *FileParser {
	i, _ := time.ParseDuration("1m")
	f := NewFileParser(fileName)
	err := f.Initialize(i)
	if err != nil {
		t.Error(err)
	}
	return f
}

// given a FileParser, return a channel that fully parses
func parseCompleteFile(t testing.TB, fp *FileParser) chan *Sample {
	datas := make(chan *Sample)

	go func() {
		for {
			sd := fp.GetNextSample()
			if sd == nil {
				// EOF
				close(datas)
				break
			}
			if sd.Error() != nil {
				t.Fatal(sd.Error())
			}
			datas <- sd
		}
	}()
	return datas
}

// given a FileParser, fully parse while confirming uptime is changing and we get the expected number of MyqData samples
func checkFileParserExpected(t *testing.T, fp *FileParser, expected int) {
	i := 0
	samples := make(chan *Sample)

	go func() {
		for {
			sd := fp.GetNextSample()
			if sd != nil {
				if sd.Error() != nil {
					t.Error(sd.Error())
				}
				samples <- sd
				continue
			} else {
				// EOF
				close(samples)
				break
			}
		}
	}()

	var prev *Sample
	for sample := range samples {
		t.Log("New sample", i, len(sample.Data), sample.Data[`uptime`])
		if prev != nil {
			t.Log("\tPrev", i, len(prev.Data), prev.Data[`uptime`])

			if prev.Data[`uptime`] == sample.Data[`uptime`] {
				t.Fatal("previous has same uptime")
			}
		}

		if prev != nil && len(prev.Data) > 0 && len(prev.Data) > len(sample.Data) {
			t.Log(prev.Data[`uptime`], " (previous) had", len(prev.Data), "keys.  Current current has", len(sample.Data))
			for pkey := range prev.Data {
				_, ok := sample.Data[pkey]
				if !ok {
					t.Log("Missing", pkey, "from current sample")
				}
			}
			t.Fatal("")
		}
		prev = sample
		i++
	}

	if i != expected {
		t.Errorf("Got unexpected number of datas: %d", i)
	}
}

// given a MyqData of STATUS output, check that some keys parse as expected
func checkStatusParseTypes(t *testing.T, data *Sample) {
	typeTests := map[string]string{
		"connections":                "int64",
		"compression":                "string",
		"wsrep_local_send_queue_avg": "float64",
		"binlog_snapshot_file":       "string",
	}

	ssp := NewSampleSet()
	ssp.SetSample(`status`, data)

	for varname, expectedtype := range typeTests {
		i, ierr := ssp.GetInt(SourceKey{`status`, varname})
		if ierr == nil {
			if expectedtype != "int64" {
				t.Fatal("Found integer, expected", expectedtype, "for", varname, "value: `", i, "`")
			} else {
				continue
			}
		}

		f, ferr := ssp.GetFloat(SourceKey{`status`, varname})
		if ferr == nil {
			if expectedtype != "float64" {
				t.Fatal("Found float, expected", expectedtype, "for", varname, "value: `", f, "`")
			} else {
				continue
			}
		}

		s, serr := ssp.GetString(SourceKey{`status`, varname})
		if serr == nil {
			if expectedtype != "string" {
				t.Fatal("Found string, expected", expectedtype, "for", varname, "value: `", s, "`")
			} else {
				continue
			}
		}
	}
}

// Actual tests

// Test missing file
func TestNewFileParserFail(t *testing.T) {
	_, err := newTestFileParser("/fooey/kablooie")
	if err == nil {
		t.Error("No error!")
	}
}

// Test empty file
func TestNewFileParserEmpty(t *testing.T) {
	fp := newGoodFileParser(t, "/dev/null")

	data := fp.GetNextSample() // EOF should return nil, nil
	if data != nil {
		t.Error("How did we get a sample?") // Any result is a failure
	}
}

// Test bad interval
func TestFileParserBadInterval(t *testing.T) {
	i, _ := time.ParseDuration("1ns")
	f := NewFileParser("/dev/null")
	err := f.Initialize(i)
	if err == nil {
		t.Error("Allowed 1ns interval")
	}
}

// Test mysqladmin output files

// Confirm parsing yields good data
func TestSingleSample(t *testing.T) {
	fp := newGoodFileParser(t, "./testdata/mysqladmin.single")
	data := fp.GetNextSample()
	if data.Error() != nil {
		t.Error(data.Error())
	}
	checkStatusParseTypes(t, data)
}

// Check two
func TestTwoSamples(t *testing.T) {
	fp := newGoodFileParser(t, "./testdata/mysqladmin.two")
	checkFileParserExpected(t, fp, 2)
}

// Check many
func TestManySamples(t *testing.T) {
	if testing.Short() {
		return
	}

	fp := newGoodFileParser(t, "./testdata/mysqladmin.lots")
	checkFileParserExpected(t, fp, 220)
}

// Check 1m interval
func Test1mSamples(t *testing.T) {
	fp := newGoodFile1mParser(t, "./testdata/mysqladmin.lots")
	checkFileParserExpected(t, fp, 5)
}

// Test mysql batch output files

// Confirm parsing yields good data
func TestSingleBatchSample(t *testing.T) {
	fp := newGoodFileParser(t, "./testdata/mysql.single")

	data := fp.GetNextSample()
	if data.Error() != nil {
		t.Error(data.Error())
	}

	checkStatusParseTypes(t, data)
}

// Check two
func TestTwoBatchSamples(t *testing.T) {
	fp := newGoodFileParser(t, "./testdata/mysql.two")
	checkFileParserExpected(t, fp, 2)
}

// Check many
func TestManyBatchSamples(t *testing.T) {
	if testing.Short() {
		return
	}

	fp := newGoodFileParser(t, "./testdata/mysql.lots")
	checkFileParserExpected(t, fp, 215)
}

// Check 1m interval
func Test1mBatchSamples(t *testing.T) {
	fp := newGoodFile1mParser(t, "./testdata/mysql.lots")
	checkFileParserExpected(t, fp, 4)
}

// Check Toku DB output
func TestTokuSample(t *testing.T) {
	fp := newGoodFileParser(t, "./testdata/mysql.toku")
	checkFileParserExpected(t, fp, 2)
}

// Benchmarking

// Benchmark a given fileName
func benchmarkFile(b *testing.B, fileName string) {
	for i := 0; i < b.N; i++ {
		fp := newGoodFileParser(b, fileName)
		for range parseCompleteFile(b, fp) {
		}
	}
}

// Benchmark a single mysqladmin
func BenchmarkParseStatus(b *testing.B) {
	benchmarkFile(b, "./testdata/mysqladmin.single")
}

// Benchmark a single mysql batch
func BenchmarkParseStatusBatch(b *testing.B) {
	benchmarkFile(b, "./testdata/mysql.single")
}

// Benchmark a single variables batch
func BenchmarkParseVariablesBatch(b *testing.B) {
	benchmarkFile(b, "./testdata/variables")
}

// Benchmark a single mysqladmin variables
func BenchmarkParseVariablesTabular(b *testing.B) {
	benchmarkFile(b, "./testdata/variables.tab")
}

// Benchmark a large batch output file
func BenchmarkParseManyBatchSamples(b *testing.B) {
	benchmarkFile(b, "./testdata/mysql.lots")
}

// Benchmark a large mysqladmin output file
func BenchmarkParseManySamples(b *testing.B) {
	benchmarkFile(b, "./testdata/mysqladmin.lots")
}

// Benchmark large mysqladmin by minutes
func BenchmarkParseManySamplesLongInterval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fp := newGoodFile1mParser(b, "./testdata/mysqladmin.lots")
		for range parseCompleteFile(b, fp) {
		}
	}
}
