package clientconf

import (
	"fmt"
	"os/user"
	"testing"
)

func TestGetCnfFiles(t *testing.T) {
	files := getCnfFiles()

	if files[0] != `/etc/my.cnf` {
		t.Errorf(`unexpected files[0] value: %s`, files[0])
	}

	if len(files) != 4 {
		t.Errorf(`unexpected files length: %d`, len(files))
	}
}

func TestAppendFiles(t *testing.T) {
	files := []string{
		`./testcnf/my.cnf`,
		`./testcnf/.my.cnf`,
		`/dev/null`,
	}

	cnf := initCnf()
	err := appendFiles(cnf, files)
	if err != nil {
		t.Fatal(err)
	}

	if !cnf.HasSection(`client`) {
		t.Fatalf(`no [client] section found: %v`, cnf.SectionStrings())
	}

	clientMap := cnf.Section(`client`).KeysHash()

	expectedMap := map[string]string{
		`port`:     `3306`,
		`password`: `my password`,
		`socket`:   `/tmp/mysql.sock`,
	}
	for k, v := range expectedMap {
		if clientMap[k] != v {
			t.Errorf(`unexpected value for key %s: %s`, k, clientMap[k])
		}
	}
}

func TestApplyFlags(t *testing.T) {
	cnf := initCnf()

	userFlag = "testuser"
	passwordFlag = "testpassword"
	hostFlag = "testhost"
	portFlag = "testport"
	socketFlag = "testsocket"

	applyFlags(cnf)

	if !cnf.HasSection(`client`) {
		t.Fatal(`missing client section`)
	}

	clientMap := cnf.Section(`client`).KeysHash()
	expectedMap := map[string]string{
		`user`:     `testuser`,
		`password`: `testpassword`,
		`host`:     `testhost`,
		`port`:     `testport`,
		`socket`:   `testsocket`,
	}
	for k, v := range expectedMap {
		if clientMap[k] != v {
			t.Errorf(`unexpected value for key %s: %s`, k, clientMap[k])
		}
	}
}

func TestCnfToConfig(t *testing.T) {
	cnf := initCnf()

	config, err := cnfToConfig(cnf)
	if err != nil {
		t.Fatal(err)
	}

	// Get the current username to build expected DSN
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	expectedDSN := fmt.Sprintf("%s@tcp(127.0.0.1:3306)/", currentUser.Username)

	if config.FormatDSN() != expectedDSN {
		t.Errorf(`Unexpected dsn: %s (expected: %s)`, config.FormatDSN(), expectedDSN)
	}

	// Second round
	cnf = initCnf()
	userFlag = "testuser"
	passwordFlag = "testpassword"
	hostFlag = "testhost"
	portFlag = "testport"
	socketFlag = ""
	applyFlags(cnf)
	config, err = cnfToConfig(cnf)
	if err != nil {
		t.Fatal(err)
	}
	if config.FormatDSN() != `testuser:testpassword@tcp(testhost:testport)/` {
		t.Errorf(`Unexpected dsn: %s`, config.FormatDSN())
	}

}

func TestCnfToConfigSSL(t *testing.T) {
	cnf := initCnf()
	userFlag = "jayj"
	passwordFlag = ""
	hostFlag = ""
	portFlag = ""
	socketFlag = ""
	sslCertFlag = "./testcnf/client-cert.pem"
	sslKeyFlag = "./testcnf/client-key.pem"
	sslCaFlag = "./testcnf/ca.pem"
	applyFlags(cnf)

	config, err := cnfToConfig(cnf)
	if err != nil {
		t.Fatal(err)
	}
	if config.FormatDSN() != `jayj@tcp(127.0.0.1:3306)/?tls=custom` {
		t.Errorf(`Unexpected dsn: %s`, config.FormatDSN())
	}

}

func TestLoosePrefix(t *testing.T) {
	// Test that loose- prefixed keys are recognized
	files := []string{
		`./testcnf/.my.cnf.loose`,
	}

	cnf := initCnf()
	err := appendFiles(cnf, files)
	if err != nil {
		t.Fatal(err)
	}

	if !cnf.HasSection(`client`) {
		t.Fatalf(`no [client] section found: %v`, cnf.SectionStrings())
	}

	clientMap := cnf.Section(`client`).KeysHash()

	// Test that loose- prefixed keys are accessible via getConfigValue
	if user, ok := getConfigValue(clientMap, `user`); !ok || user != `standarduser` {
		t.Errorf(`expected standarduser from standard key, got: %s`, user)
	}

	if password, ok := getConfigValue(clientMap, `password`); !ok || password != `loosepass` {
		t.Errorf(`expected loosepass from loose- prefix, got: %s`, password)
	}

	if host, ok := getConfigValue(clientMap, `host`); !ok || host != `loosehost.example.com` {
		t.Errorf(`expected loosehost.example.com from loose- prefix, got: %s`, host)
	}

	if port, ok := getConfigValue(clientMap, `port`); !ok || port != `3307` {
		t.Errorf(`expected 3307 from loose- prefix, got: %s`, port)
	}

	if sslmode, ok := getConfigValue(clientMap, `ssl-mode`); !ok || sslmode != `VERIFY_CA` {
		t.Errorf(`expected VERIFY_CA from loose- prefix, got: %s`, sslmode)
	}

	if _, ok := getConfigValue(clientMap, `enable-cleartext-plugin`); !ok {
		t.Errorf(`expected enable-cleartext-plugin to be found via loose- prefix`)
	}

	// Test cnfToConfig with loose- prefixed keys
	config, err := cnfToConfig(cnf)
	if err != nil {
		t.Fatal(err)
	}

	// Standard key should take precedence over loose-user
	if config.User != `standarduser` {
		t.Errorf(`expected standarduser, got: %s`, config.User)
	}

	// These should come from loose- prefixed keys
	if config.Passwd != `loosepass` {
		t.Errorf(`expected loosepass, got: %s`, config.Passwd)
	}

	expectedAddr := `loosehost.example.com:3307`
	if config.Addr != expectedAddr {
		t.Errorf(`expected %s, got: %s`, expectedAddr, config.Addr)
	}

	if !config.AllowCleartextPasswords {
		t.Errorf(`expected AllowCleartextPasswords to be true`)
	}
}
