package viewer

import "testing"

func TestcalculateDiff(t *testing.T) {
	diff := calculateDiff(200, 100)
	if diff != 100 {
		t.Errorf(`unexpected diff: %f`, diff)
	}

	diff = calculateDiff(50, 200)
	if diff != 50 {
		t.Errorf(`unexpected diff with a low bigger number: %f`, diff)
	}
}

func TestcalculateRate(t *testing.T) {
	rate := calculateRate(200, 100, 5)
	if rate != 20 {
		t.Errorf(`unexpected rate: %f`, rate)
	}

	rate = calculateRate(200, 100, -5)
	if rate != 100 {
		t.Errorf(`unexpected rate with negative seconds: %f`, rate)
	}
}

func TestFitString(t *testing.T) {
	out := FitString("fooey", 4)
	if len(out) != 4 && out != "fooe" {
		t.Errorf("truncated string improperly: '%s'", out)
	}

	out = FitString("f", 4)
	if len(out) != 4 && out != "   f" {
		t.Errorf("padded string improperly: '%s'", out)
	}
}

func TestfitStringLeft(t *testing.T) {
	out := fitStringLeft("fooey", 4)
	if len(out) != 4 && out != "fooe" {
		t.Errorf("truncated string improperly: '%s'", out)
	}
	out = fitStringLeft("f", 4)
	if len(out) != 4 && out != "f   " {
		t.Errorf("padded string left improperly: '%s'", out)
	}
}
