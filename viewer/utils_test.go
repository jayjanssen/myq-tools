package viewer

import "testing"

func TestFitString(t *testing.T) {
	out := fitString("fooey", 4)
	if len(out) != 4 && out != "fooe" {
		t.Errorf("truncated string improperly: '%s'", out)
	}

	out = fitString("f", 4)
	if len(out) != 4 && out != "   f" {
		t.Errorf("padded string improperly: '%s'", out)
	}

	out = fitStringLeft("f", 4)
	if len(out) != 4 && out != "f   " {
		t.Errorf("padded string left improperly: '%s'", out)
	}
}
