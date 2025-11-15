package loader

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

var sources_live_test = []SourceName{`status`, `variables`}

// setupMySQLContainer starts a MySQL container for testing
func setupMySQLContainer(t testing.TB) (*testmysql.MySQLContainer, *mysql.Config) {
	t.Helper()

	// Allow version override via environment variable
	version := os.Getenv("MYSQL_VERSION")
	if version == "" {
		version = "8.0"
	}

	ctx := context.Background()

	container, err := testmysql.Run(ctx,
		fmt.Sprintf("mysql:%s", version),
		testmysql.WithDatabase("test"),
		testmysql.WithUsername("root"),
		testmysql.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("Failed to start MySQL container: %s", err)
	}

	// Ensure container cleanup
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	})

	// Get connection string and parse to mysql.Config
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %s", err)
	}

	config, err := mysql.ParseDSN(connStr)
	if err != nil {
		t.Fatalf("Failed to parse DSN: %s", err)
	}

	return container, config
}

func NewTestLiveLoader(config *mysql.Config) (*LiveLoader, error) {
	i, _ := time.ParseDuration("1s")
	db := NewLiveLoader(config)
	err := db.Initialize(i, sources_live_test)

	return db, err
}

func NewGoodLiveLoader(t testing.TB) *LiveLoader {
	_, config := setupMySQLContainer(t)
	db, err := NewTestLiveLoader(config)
	if err != nil {
		t.Fatalf("Connection error: %s", err)
	}
	return db
}

// NewLiveLoader
// - should return an error on a bad dsn
func TestNewLiveLoaderFail(t *testing.T) {
	config := mysql.NewConfig()
	config.Net = "tcp"
	config.Addr = "127.0.0.1:7777"
	_, err := NewTestLiveLoader(config)
	if err == nil {
		t.Error("No error!")
	}
}

// - should be able to make a successful connection
func TestNewLiveLoader(t *testing.T) {
	_, config := setupMySQLContainer(t)
	_, err := NewTestLiveLoader(config)
	if err != nil {
		t.Error(err)
	}
}

// - should return an error on a bad user
func TestNewLiveLoaderUserFail(t *testing.T) {
	_, config := setupMySQLContainer(t)
	config.User = "bad_user"
	_, err := NewTestLiveLoader(config)
	if err == nil {
		t.Error("No error!")
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

		// status/uptime == Int (should be >= 0, fresh containers have low uptime)
		uptime, err := curr.GetInt(SourceKey{`status`, `uptime`})
		if err != nil {
			t.Error(err)
		} else if uptime < 0 {
			t.Errorf("Expected uptime >= 0, got: %d", uptime)
		}

		// Com_select == Int
		comSelect, err := curr.GetString(SourceKey{`status`, `com_select`})
		if err != nil {
			t.Error(err)
		}

		t.Log("Com_select:", comSelect)

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
