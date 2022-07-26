package viewer

import "github.com/jayjanssen/myq-tools2/loader"

// "Mix-in" style structs with helper functions around Sources

// A col with one key
type oneKeyCol struct {
	Key loader.SourceKey `yaml:"key"`
}

func (okc oneKeyCol) GetSources() []loader.SourceName {
	return []loader.SourceName{okc.Key.SourceName}
}

// A col with Keys that expand
type expandableKeysCol struct {
	Keys         []loader.SourceKey `yaml:"keys"`
	expandedKeys []loader.SourceKey
}

func (ekc expandableKeysCol) GetSources() (result []loader.SourceName) {
	for _, key := range ekc.Keys {
		result = append(result, key.SourceName)
	}
	return
}
