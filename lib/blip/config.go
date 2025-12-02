package blip

import (
	"fmt"

	"github.com/cashapp/blip"
	"github.com/go-sql-driver/mysql"
)

// ConfigFromMySQL converts a MySQL config to a blip ConfigMonitor
func ConfigFromMySQL(mysqlCfg *mysql.Config) (blip.ConfigMonitor, error) {
	cfg := blip.ConfigMonitor{
		MonitorId: mysqlCfg.Addr,
		Hostname:  mysqlCfg.Addr,
		Username:  mysqlCfg.User,
		Password:  mysqlCfg.Passwd,
		Socket:    "",
	}

	// Extract host and port if available
	if mysqlCfg.Net == "tcp" && mysqlCfg.Addr != "" {
		// Addr is in format "host:port"
		cfg.Hostname = mysqlCfg.Addr
	} else if mysqlCfg.Net == "unix" {
		cfg.Socket = mysqlCfg.Addr
		cfg.Hostname = ""
	}

	// Set MonitorId to something meaningful
	if cfg.Socket != "" {
		cfg.MonitorId = cfg.Socket
	} else if cfg.Hostname != "" {
		cfg.MonitorId = cfg.Hostname
	} else {
		cfg.MonitorId = "localhost"
	}

	return cfg, nil
}

// MakeDSN creates a DSN from a blip ConfigMonitor for use with sql.Open
func MakeDSN(cfg blip.ConfigMonitor) (string, error) {
	mysqlCfg := mysql.NewConfig()

	if cfg.Socket != "" {
		mysqlCfg.Net = "unix"
		mysqlCfg.Addr = cfg.Socket
	} else if cfg.Hostname != "" {
		mysqlCfg.Net = "tcp"
		mysqlCfg.Addr = cfg.Hostname
	} else {
		return "", fmt.Errorf("neither hostname nor socket specified")
	}

	mysqlCfg.User = cfg.Username
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.DBName = "" // Don't need to specify a database

	return mysqlCfg.FormatDSN(), nil
}
