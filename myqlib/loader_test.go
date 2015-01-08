package myqlib

import (
	"testing"
	"time"
)

func TestBadFile(t *testing.T) {
	l := FileLoader{loaderInterval(1 * time.Second), "/fooey/kablooie", ""}
	_, err := GetState(l)

	if err == nil {
		t.Error("Somehow able to open /fooey/kablooie")
	}
}

func TestEmpty(t *testing.T) {
	l := FileLoader{loaderInterval(1 * time.Second), "/dev/null", ""}
	ch, err := l.getStatus()
	if err != nil {
		t.Error("Got error opening /dev/null:", err)
	}
	_, ok := <-ch
	if ok {
		t.Error("How did we get a sample?") // Any result is a failure
	}
}

func TestMetric(t *testing.T) {
	sample := make(MyqSample)

	sample["key"] = "value"

	if sample.Length() != 1 {
		t.Fatal("Expecting 1 KV, got", sample.Length())
	}
}
