package viewer

import (
	"fmt"
	"testing"

	"github.com/jayjanssen/myq-tools/lib/loader"
)

func getTestPercentCol() PercentCol {
	rc := PercentCol{}
	rc.Name = "dirt"
	rc.Description = "Buffer pool percent dirty"
	rc.Type = "Percent"
	rc.Numerator = loader.SourceKey{SourceName: "status", Key: "innodb_buffer_pool_pages_dirty"}
	rc.Denominator = loader.SourceKey{SourceName: "status", Key: "innodb_buffer_pool_pages_total"}
	rc.Length = 4
	rc.Units = PERCENT
	rc.Precision = 0

	return rc
}

func TestPercentColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestPercentCol()
}

func getTestPercentState(numerator, denominator string) loader.StateReader {
	sp := loader.NewState()

	cursamp := loader.NewSample()
	cursamp.Data[`innodb_buffer_pool_pages_dirty`] = numerator
	cursamp.Data[`innodb_buffer_pool_pages_total`] = denominator

	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	return sp
}

func TestPercentColgetPercent(t *testing.T) {
	col := getTestPercentCol()
	state := getTestPercentState(`86716`, `15999992`)

	percent, err := col.getPercent(state)
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%.5f", percent) != `0.54198` {
		t.Errorf(`unexpected percent: '%s'`, fmt.Sprintf("%.5f", percent))
	}

	data := col.GetData(state)
	if data[0] != `  1%` {
		t.Errorf(`unexpected data: '%s'`, data)
	}

	state = getTestPercentState(`86716`, `notanum`)
	_, err = col.getPercent(state)
	if err == nil {
		t.Error(`expected denominator error`)
	}

	state = getTestPercentState(`notanum`, `15999992`)
	_, err = col.getPercent(state)
	if err == nil {
		t.Error(`expected numerator error`)
	}

}
