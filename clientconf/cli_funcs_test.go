package clientconf

import "testing"

func TestSetMySQLFlags(t *testing.T) {
	SetMySQLFlags()
}

func TestGenerateConfig(t *testing.T) {
	userFlag = "testuser"

	config, err := GenerateConfig()
	if err != nil {
		t.Error(err)
	}

	if config.FormatDSN() != `testuser@tcp(127.0.0.1:3306)/` {
		t.Errorf(`Unexpected dsn: %s`, config.FormatDSN())
	}
}

func TestSocketConfig(t *testing.T) {
	userFlag = "testuser"
	socketFlag = "/var/lib/mysql/mysql.sock"

	config, err := GenerateConfig()
	if err != nil {
		t.Error(err)
	}

	if config.FormatDSN() != `testuser@unix(/var/lib/mysql/mysql.sock)/` {
		t.Errorf(`Unexpected dsn: %s`, config.FormatDSN())
	}
}
