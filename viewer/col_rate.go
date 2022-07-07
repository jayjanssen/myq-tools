package viewer

import "github.com/jayjanssen/myq-tools2/loader"

type RateCol struct {
	numCol
	Key SourceKey
}

// Data for this view based on the state
func (c RateCol) GetData(sr loader.StateReader) (result []string) {
	raw := c.getRate(sr)

	var res []string
	res = append(res, raw)
	return res
}

func (c RateCol) getRate(sr loader.StateReader) (result string) {
	cur, prev := sr.GetKeyCurPrev()

	// Apply precision
	// return fmt.Sprintf("%.2s")
	return "   5"
}
