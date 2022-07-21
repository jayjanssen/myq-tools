package loader

import (
	"testing"
	"time"
)

var sources_live_test = []SourceName{`status`, `variables`}

func NewTestLiveLoader(dsn string) (*LiveLoader, error) {
	i, _ := time.ParseDuration("1s")
	db := NewLiveLoader(dsn)
	err := db.Initialize(i, sources_live_test)

	return db, err
}

func NewGoodLiveLoader(t testing.TB) *LiveLoader {
	db, err := NewTestLiveLoader("root@tcp(127.0.0.1:3306)/")
	if err != nil {
		t.Errorf("Connection error: %s", err)
	}
	return db
}

// NewLiveLoader
// - should return an error on a bad dsn
func TestNewLiveLoaderFail(t *testing.T) {
	_, err := NewTestLiveLoader("10.1.1.1")
	if err == nil {
		t.Error("No error!")
	}
}

// - should be able to make a successful connection
func TestNewLiveLoader(t *testing.T) {
	_, err := NewTestLiveLoader("tcp(127.0.0.1:3306)/")
	if err != nil {
		t.Error(err)
	}
}

// Sql Loader implements the Loader interface
func TestLiveLoaderImplementsLoader(t *testing.T) {
	var _ Loader = NewGoodLiveLoader(t)
}

// GetSample
func TestLiveLoaderGetSample(t *testing.T) {
	l := NewGoodLiveLoader(t)

	ch := l.GetStateChannel()

	// Block waiting for a sample from ch, or else a timeout
	select {
	case s := <-ch:
		curr := s.GetCurrent()
		errs := curr.GetErrors()
		if errs != nil {
			t.Fatalf("Sample returned error: %v", errs)
		}

		// variables/port == 3306
		port, err := curr.GetInt(SourceKey{`variables`, `port`})
		if err != nil {
			t.Error(err)
		} else if port != 3306 {
			t.Errorf("Expected port 3306, got: %d", port)
		}

		// status/uptime == Int
		uptime, err := curr.GetInt(SourceKey{`status`, `uptime`})
		if err != nil {
			t.Error(err)
		} else if uptime < 10 {
			t.Errorf("Expected uptime > 10, got: %d", uptime)
		}

	case <-time.After(2 * time.Second):
		t.Error("Sample missing")
	}
}

func Benchmark(b *testing.B) {
	l := NewGoodLiveLoader(b)

	for i := 0; i < b.N; i++ {
		l.getSample(STATUS_QUERY)
		l.getSample(VARIABLES_QUERY)
	}
}
