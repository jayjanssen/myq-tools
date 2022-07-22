package loader

import (
	"testing"
	"time"
)

var sources_file_test = []SourceName{`status`, `variables`}

func NewTestFileLoader(statusFile, varFile string) (*FileLoader, error) {
	i, _ := time.ParseDuration("1s")
	fl := NewFileLoader(statusFile, varFile)
	err := fl.Initialize(i, sources_file_test)
	return fl, err
}

func NewGoodFileLoader(t testing.TB, statusFile, varFile, intervalStr string) *FileLoader {
	i, _ := time.ParseDuration(intervalStr)
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
	l := NewGoodFileLoader(t, "/dev/null", "", "1s")

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
	var _ Loader = NewGoodFileLoader(t, "/dev/null", "", "1s")
}

// Ensure variables are loaded properly
func TestFileLoaderVariables(t *testing.T) {
	l := NewGoodFileLoader(t, "./testdata/mysql.single", "./testdata/variables", "1s")
	ch := l.GetStateChannel()

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.GetCurrent()
		errs := curr.GetErrors()
		if errs != nil {
			t.Errorf("Sample returned error: %v", errs)
		} else {
			mc, err := curr.GetInt(SourceKey{`variables`, `max_connections`})
			if err != nil {
				t.Error(err)
			} else if mc != 151 {
				t.Error("Expected 151 max_connections")
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}

// Ensure missing variables are handled appropriately
func TestFileLoaderNilVarfile(t *testing.T) {
	l := NewGoodFileLoader(t, "./testdata/mysql.single", "", "1s")
	ch := l.GetStateChannel()

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.GetCurrent()
		if errs := curr.GetErrors(); errs != nil {
			t.Fatalf("Sample returned error: %v", errs)
		}

		if curr.HasSource(`variables`) {
			t.Error("found unexpected variables")
		}

		mc, _ := curr.GetInt(SourceKey{`status`, `questions`})
		if mc != 914 {
			t.Errorf("Expected 914 questions in sample, got `%d`", mc)
		}
	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}

func TestFileLoader5sInterval(t *testing.T) {
	l := NewGoodFileLoader(t, "./testdata/mysqladmin.lots", "", "5s")
	ch := l.GetStateChannel()

	// throw away the first state
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Error("sample missing")
	}

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.GetCurrent()
		if errs := curr.GetErrors(); errs != nil {
			t.Fatalf("Sample returned error: %v", errs)
		}

		mc, _ := curr.GetInt(SourceKey{`status`, `questions`})
		if mc != 65218910 {
			t.Errorf("Expected 65218910 questions in sample, got `%d`", mc)
		}

		diff := s.SecondsDiff()
		if diff != 5 {
			t.Errorf("unexpected SecondsDiff: %f", diff)
		}

		timestamp := s.GetTimeString()
		if timestamp != `5s` {
			t.Errorf("unexpected GetTimeString: %s", timestamp)
		}

	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}
