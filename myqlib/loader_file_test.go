package myqlib

import "testing"

func TestBadFile(t *testing.T) {
  _, err := GetSamplesFile("/fooey/kablooie")

  if err == nil {
    t.Error( "Somehow able to open /fooey/kablooie" ) 
  }
}

func TestEmpty(t *testing.T) {
  ch, err := GetSamplesFile("/dev/null")
  if err != nil {
    t.Error( "Got error opening /dev/null:", err)
  }
  _, ok := <- ch
  if ok {
    t.Error( "How did we get a sample?" ) // Any result is a failure
  }
}
