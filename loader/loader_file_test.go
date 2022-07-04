package loader

import (
	"testing"
	"time"
)

var sources_file_test = []SourceKey{`status`, `variables`}

func NewTestFileLoader(statusFile, varFile string) (*FileLoader, error) {
	i, _ := time.ParseDuration("1s")
	fl := NewFileLoader(statusFile, varFile)
	err := fl.Initialize(i, sources_file_test)
	return fl, err
}

func NewGoodFileLoader(t testing.TB, statusFile, varFile string) *FileLoader {
	i, _ := time.ParseDuration("1s")
	fl := NewFileLoader(statusFile, varFile)

	err := fl.Initialize(i, sources_file_test)

	if err != nil {
		t.Fatalf("File open error: %v", err)
	}
	return fl
}

// NewFileLoader
// - should return an error on a bad dsn
func TestNewFileLoaderFail(t *testing.T) {
	_, err := NewTestFileLoader("/fooey/kablooie", "/bad/var")
	if err == nil {
		t.Error("No error!")
	}
}

// -- empty files should not return samples
func TestNewFileLoaderEmpty(t *testing.T) {
	l := NewGoodFileLoader(t, "/dev/null", "")

	ch := l.GetStateChannel()
	select {
	case s := <-ch:
		if s != nil {
			t.Errorf("How did we get a state? %v", s) // Any result is a failure
		}
	case <-time.After(2 * time.Second):
	}
}

// File Loader implements the Loader interface
func TestFileLoaderImplementsLoader(t *testing.T) {
	var _ Loader = NewGoodFileLoader(t, "/dev/null", "")
}

// Ensure variables are loaded properly
func TestFileLoaderVariables(t *testing.T) {
	l := NewGoodFileLoader(t, "./testdata/mysql.single", "./testdata/variables")
	ch := l.GetStateChannel()

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.Current
		errs := curr.GetErrors()
		if errs != nil {
			t.Errorf("Sample returned error: %v", errs)
		} else if _, ok := curr.Samples[`variables`]; !ok {
			t.Error("Missing variables")
		} else {
			varss := curr.Samples[`variables`]
			mc, _ := varss.GetInt(`max_connections`)
			if mc != 151 {
				t.Error("Expected 151 max_connections")
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}

// Ensure missing variables are handled appropriately
func TestFileLoaderNilVarfile(t *testing.T) {
	l := NewGoodFileLoader(t, "./testdata/mysql.single", "")
	ch := l.GetStateChannel()

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.Current
		if errs := curr.GetErrors(); errs != nil {
			t.Errorf("Sample returned error: %v", errs)
		} else if _, ok := curr.Samples[`variables`]; ok {
			t.Error("found unexpected variables")
		} else {
			statuss := curr.Samples[`status`]
			mc, _ := statuss.GetInt(`questions`)
			if mc != 914 {
				t.Error("Expected 914 questions in sample")
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}
