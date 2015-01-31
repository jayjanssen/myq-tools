package myqlib

import (
	"testing"
	"time"
)

func TestExpand(t *testing.T) {
	l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.single", ""}
	samples, err := l.getStatus()
	if err != nil {
		t.Error(err)
	}
	sample := <-samples
	
	assert := func(test_name string, variables []string, expected int) {
		expanded := expand_variables( variables, sample )
		if len( expanded ) != expected {
			t.Fatal( test_name, `Failed, got: `, len(expanded), ", expected:", expected )
		}
	}
	
	assert( `dmls`, []string{`com_insert.*`, `com_update.*`, `com_delete.*`, `Com_load`, `Com_replace.*`, `Com_truncate`}, 6 )
	assert( `no_regex`, []string{`com_select`, `qcache_hits`}, 2 )
}

func BenchmarkVariableExpand(b *testing.B) {
	l := FileLoader{loaderInterval(1 * time.Second), "../testdata/mysqladmin.single", ""}
	samples, err := l.getStatus()
	if err != nil {
		b.Error(err)
	}
	sample := <-samples
	
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = expand_variables( []string{`com_insert.*`, `com_update.*`, `com_delete.*`, `Com_load`, `Com_replace.*`, `Com_truncate`}, sample )
	}
}