package loader

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func testExpectedSkey(t *testing.T, sk SourceKey, source, key string) {
	if sk.SourceName != SourceName(source) && sk.Key != key {
		t.Errorf("Invalid parsing of sk: %v", sk)
	}
}

func TestSourceKeyParse(t *testing.T) {
	var test_yaml = `---
- source/key
- source/key/extra
`

	skeys := make([]SourceKey, 0)
	err := yaml.Unmarshal([]byte(test_yaml), &skeys)
	if err != nil {
		t.Error(err)
	}

	if len(skeys) != 2 {
		t.Errorf("didn't parse all the keys: %d", len(skeys))
	}

	testExpectedSkey(t, skeys[0], "source", "key")
	testExpectedSkey(t, skeys[1], "source", "key/extra")
}

func TestSourceKeyParseErr(t *testing.T) {
	bad_yaml := `?`
	test_yaml := `---
- sourcekey
`

	skeys := make([]SourceKey, 0)
	err := yaml.Unmarshal([]byte(bad_yaml), &skeys)
	if err == nil {
		t.Error(`expected error on bad yaml`)
	}

	err = yaml.Unmarshal([]byte(test_yaml), &skeys)
	if err == nil {
		t.Error(`expected err on bad sourcekey`)
	} else {
		if err.Error() != "sourcekey invalid format: sourcekey" {
			t.Errorf("unexpected error str: %s", err)
		}
	}

}
