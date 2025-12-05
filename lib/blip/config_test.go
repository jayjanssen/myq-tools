package blip

import (
	"testing"

	"github.com/cashapp/blip"
	"github.com/go-sql-driver/mysql"
)

func TestConfigFromMySQL_TCP(t *testing.T) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = "localhost:3306"
	mysqlCfg.User = "testuser"
	mysqlCfg.Passwd = "testpass"

	blipCfg, err := ConfigFromMySQL(mysqlCfg)
	if err != nil {
		t.Fatalf("ConfigFromMySQL failed: %v", err)
	}

	if blipCfg.Hostname != "localhost:3306" {
		t.Errorf("Expected Hostname 'localhost:3306', got '%s'", blipCfg.Hostname)
	}
	if blipCfg.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", blipCfg.Username)
	}
	if blipCfg.Password != "testpass" {
		t.Errorf("Expected Password 'testpass', got '%s'", blipCfg.Password)
	}
	if blipCfg.Socket != "" {
		t.Errorf("Expected Socket to be empty, got '%s'", blipCfg.Socket)
	}
	if blipCfg.MonitorId != "localhost:3306" {
		t.Errorf("Expected MonitorId 'localhost:3306', got '%s'", blipCfg.MonitorId)
	}
}

func TestConfigFromMySQL_Unix(t *testing.T) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.Net = "unix"
	mysqlCfg.Addr = "/tmp/mysql.sock"
	mysqlCfg.User = "testuser"
	mysqlCfg.Passwd = "testpass"

	blipCfg, err := ConfigFromMySQL(mysqlCfg)
	if err != nil {
		t.Fatalf("ConfigFromMySQL failed: %v", err)
	}

	if blipCfg.Socket != "/tmp/mysql.sock" {
		t.Errorf("Expected Socket '/tmp/mysql.sock', got '%s'", blipCfg.Socket)
	}
	if blipCfg.Hostname != "" {
		t.Errorf("Expected Hostname to be empty, got '%s'", blipCfg.Hostname)
	}
	if blipCfg.MonitorId != "/tmp/mysql.sock" {
		t.Errorf("Expected MonitorId '/tmp/mysql.sock', got '%s'", blipCfg.MonitorId)
	}
}

func TestConfigFromMySQL_MonitorIdFallback(t *testing.T) {
	tests := []struct {
		name       string
		setupCfg   func() *mysql.Config
		expectedId string
	}{
		{
			name: "socket priority",
			setupCfg: func() *mysql.Config {
				cfg := mysql.NewConfig()
				cfg.Net = "unix"
				cfg.Addr = "/var/run/mysql.sock"
				return cfg
			},
			expectedId: "/var/run/mysql.sock",
		},
		{
			name: "hostname when no socket",
			setupCfg: func() *mysql.Config {
				cfg := mysql.NewConfig()
				cfg.Net = "tcp"
				cfg.Addr = "db.example.com:3306"
				return cfg
			},
			expectedId: "db.example.com:3306",
		},
		{
			name: "localhost fallback",
			setupCfg: func() *mysql.Config {
				cfg := mysql.NewConfig()
				cfg.Net = ""
				cfg.Addr = ""
				return cfg
			},
			expectedId: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlCfg := tt.setupCfg()
			blipCfg, err := ConfigFromMySQL(mysqlCfg)
			if err != nil {
				t.Fatalf("ConfigFromMySQL failed: %v", err)
			}
			if blipCfg.MonitorId != tt.expectedId {
				t.Errorf("Expected MonitorId '%s', got '%s'", tt.expectedId, blipCfg.MonitorId)
			}
		})
	}
}

func TestMakeDSN_TCP(t *testing.T) {
	cfg := blip.ConfigMonitor{
		Hostname: "localhost:3306",
		Username: "testuser",
		Password: "testpass",
	}

	dsn, err := MakeDSN(cfg)
	if err != nil {
		t.Fatalf("MakeDSN failed: %v", err)
	}

	// Parse the DSN to verify it's correct
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN failed: %v", err)
	}

	if parsed.Net != "tcp" {
		t.Errorf("Expected Net 'tcp', got '%s'", parsed.Net)
	}
	if parsed.Addr != "localhost:3306" {
		t.Errorf("Expected Addr 'localhost:3306', got '%s'", parsed.Addr)
	}
	if parsed.User != "testuser" {
		t.Errorf("Expected User 'testuser', got '%s'", parsed.User)
	}
	if parsed.Passwd != "testpass" {
		t.Errorf("Expected Passwd 'testpass', got '%s'", parsed.Passwd)
	}
}

func TestMakeDSN_Unix(t *testing.T) {
	cfg := blip.ConfigMonitor{
		Socket:   "/tmp/mysql.sock",
		Username: "testuser",
		Password: "testpass",
	}

	dsn, err := MakeDSN(cfg)
	if err != nil {
		t.Fatalf("MakeDSN failed: %v", err)
	}

	// Parse the DSN to verify it's correct
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN failed: %v", err)
	}

	if parsed.Net != "unix" {
		t.Errorf("Expected Net 'unix', got '%s'", parsed.Net)
	}
	if parsed.Addr != "/tmp/mysql.sock" {
		t.Errorf("Expected Addr '/tmp/mysql.sock', got '%s'", parsed.Addr)
	}
}

func TestMakeDSN_NoHostOrSocket(t *testing.T) {
	cfg := blip.ConfigMonitor{
		Username: "testuser",
		Password: "testpass",
	}

	_, err := MakeDSN(cfg)
	if err == nil {
		t.Fatal("Expected error when neither hostname nor socket specified, got nil")
	}

	expectedErr := "neither hostname nor socket specified"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can convert MySQL->Blip->DSN and get a valid result
	tests := []struct {
		name     string
		setupCfg func() *mysql.Config
	}{
		{
			name: "TCP connection",
			setupCfg: func() *mysql.Config {
				cfg := mysql.NewConfig()
				cfg.Net = "tcp"
				cfg.Addr = "127.0.0.1:3306"
				cfg.User = "root"
				cfg.Passwd = "secret"
				return cfg
			},
		},
		{
			name: "Unix socket connection",
			setupCfg: func() *mysql.Config {
				cfg := mysql.NewConfig()
				cfg.Net = "unix"
				cfg.Addr = "/var/lib/mysql/mysql.sock"
				cfg.User = "root"
				cfg.Passwd = "secret"
				return cfg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.setupCfg()

			// MySQL -> Blip
			blipCfg, err := ConfigFromMySQL(original)
			if err != nil {
				t.Fatalf("ConfigFromMySQL failed: %v", err)
			}

			// Blip -> DSN
			dsn, err := MakeDSN(blipCfg)
			if err != nil {
				t.Fatalf("MakeDSN failed: %v", err)
			}

			// Parse DSN and verify it matches original
			parsed, err := mysql.ParseDSN(dsn)
			if err != nil {
				t.Fatalf("ParseDSN failed: %v", err)
			}

			if parsed.Net != original.Net {
				t.Errorf("Net mismatch: expected '%s', got '%s'", original.Net, parsed.Net)
			}
			if parsed.Addr != original.Addr {
				t.Errorf("Addr mismatch: expected '%s', got '%s'", original.Addr, parsed.Addr)
			}
			if parsed.User != original.User {
				t.Errorf("User mismatch: expected '%s', got '%s'", original.User, parsed.User)
			}
			if parsed.Passwd != original.Passwd {
				t.Errorf("Passwd mismatch: expected '%s', got '%s'", original.Passwd, parsed.Passwd)
			}
		})
	}
}
