package clientconf

import (
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
		t.Fatalf(err.Error())
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
	if config.FormatDSN() != `jayj@tcp(127.0.0.1:3306)/` {
		t.Errorf(`Unexpected dsn: %s`, config.FormatDSN())
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
