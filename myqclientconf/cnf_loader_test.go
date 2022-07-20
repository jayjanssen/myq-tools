package myqclientconf

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

func TestLoadFiles(t *testing.T) {
	files := []string{
		`./testcnf/my.cnf`,
		`./testcnf/.my.cnf`,
		`/dev/null`,
	}

	cfg, err := loadFiles(files)
	if err != nil {
		t.Error(err)
	}

	if !cfg.HasSection(`client`) {
		t.Fatalf(`no [client] section found: %v`, cfg.SectionStrings())
	}

	client, err := cfg.GetSection(`client`)
	if err != nil {
		t.Error(err)
	}

	clientMap := client.KeysHash()

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
	// t.Errorf(`client: %v`, client.KeysHash())
}
