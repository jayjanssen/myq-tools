package myqlib

import (
	"testing"
	"time"
)

func TestSingleSample(t *testing.T) {
	l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.single", ""}
	samples, err := l.getStatus()
	if err != nil {
		t.Error(err)
	}

	// Check some types on some known metrics to verify autodetection
	sample := <-samples
	typeTests := map[string]string{
		"connections":                "int64",
		"compression":                "string",
		"wsrep_local_send_queue_avg": "float64",
		"binlog_snapshot_file":       "string",
	}

	for varname, expectedtype := range typeTests {
		i, ierr := sample.getInt(varname)
		if ierr == nil {
			if expectedtype != "int64" {
				t.Fatal("Found integer, expected", expectedtype, "for", varname, "value: `", i, "`")
			} else {
				continue
			}
		}
		
		f, ferr := sample.getFloat(varname)
		if ferr == nil {
			if expectedtype != "float64" {
				t.Fatal("Found float, expected", expectedtype, "for", varname, "value: `", f, "`")
			} else {
				continue
			}
		}
		
		s, serr := sample.getString(varname)
		if serr == nil {
			if expectedtype != "string" {
				t.Fatal("Found string, expected", expectedtype, "for", varname, "value: `", s, "`")
			} else {
				continue
			}
		}
	}
}

func TestTwoSamples(t *testing.T) {
	l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.two", ""}
	samples, err := l.getStatus()

	if err != nil {
		t.Error(err)
	}

	checksamples(t, samples, 2)
}

func TestManySamples(t *testing.T) {
	if testing.Short() {
		return
	}

	l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.lots", ""}
	samples, err := l.getStatus()

	if err != nil {
		t.Error(err)
	}

	checksamples(t, samples, 220)
}

func checksamples(t *testing.T, samples chan MyqSample, expected int) {
	i := 0
	var prev MyqSample
	for sample := range samples {
		t.Log("New MyqSample", i, len(sample), sample["uptime"])
		if prev != nil {
			t.Log("\tPrev", i, len(prev), prev["uptime"])

			if prev["uptime"] == sample["uptime"] {
				t.Fatal("previous has same uptime")
			}
		}

		if len(prev) > 0 && len(prev) > len(sample) {
			t.Log(prev["uptime"], "(previous) had", len(prev), "keys.  Current current has", len(sample))
			for pkey, _ := range prev {
				_, ok := (sample)[pkey]
				if !ok {
					t.Log("Missing", pkey, "from current sample")
				}
			}
			t.Fatal()
		}
		prev = sample
		i++
	}

	if i != expected {
		t.Errorf("Got unexpected number of samples: %d", i)
	}
}

func BenchmarkParseStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.single", ""}
		samples, err := l.getStatus()

		if err != nil {
			b.Error(err)
		}
		<-samples
	}
}

func BenchmarkParseVariables(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := FileLoader{loaderInterval(1 * time.Second), "../testdata/variables", ""}
		samples, err := l.getStatus()

		if err != nil {
			b.Error(err)
		}
		<-samples
	}
}