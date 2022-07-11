package viewer

import "github.com/jayjanssen/myq-tools2/loader"

type RateCol struct {
	numCol
	key loader.SourceKey
}

// Data for this view based on the state
func (c RateCol) GetData(sr loader.StateReader) []string {
	raw := c.getRate(sr)
	str := c.fitNumber(raw, c.Precision)

	return []string{str}
}

func (c RateCol) getRate(sr loader.StateReader) (result float64) {
	cur, prev := sr.GetKeyCurPrev(c.key)

	diff := CalculateDiff(cur, prev)
	secs := sr.SecondsDiff()

	if secs <= 0 {
		return diff
	} else {
		return diff / float64(secs)
	}
}
