//go:build integration
// +build integration

package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/cashapp/blip"
	_ "github.com/go-sql-driver/mysql"
)

// GetMySQLDSN returns a DSN for connecting to MySQL.
// It uses environment variables or defaults for CI environments.
func GetMySQLDSN() string {
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("MYSQL_PASSWORD")

	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}

	if password == "" {
		return user + "@tcp(" + host + ":" + port + ")/"
	}
	return user + ":" + password + "@tcp(" + host + ":" + port + ")/"
}

// ConnectMySQL attempts to connect to MySQL and returns the DB connection.
// Returns nil if connection fails (allows tests to skip gracefully).
func ConnectMySQL(t *testing.T) *sql.DB {
	dsn := GetMySQLDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Failed to open MySQL connection: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("Failed to ping MySQL server: %v", err)
		return nil
	}

	return db
}

// GetTestConfig returns a blip.ConfigMonitor configured from environment variables
func GetTestConfig() blip.ConfigMonitor {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}

	username := os.Getenv("MYSQL_USER")
	if username == "" {
		username = "root"
	}

	password := os.Getenv("MYSQL_PASSWORD")

	return blip.ConfigMonitor{
		MonitorId: host + ":" + port,
		Hostname:  host + ":" + port,
		Username:  username,
		Password:  password,
	}
}
