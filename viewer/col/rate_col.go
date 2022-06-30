package col

type RateCol struct {
	numCol
}

// // Data for this view based on the state
// func (c defaultCol) GetData(sr loader.StateReader) (result []string) {
// 	var raw string
// 	switch c.Type {
// 	case `Rate`:
// 		raw = c.getRate(sr)
// 	}

// 	var res []string
// 	res = append(res, raw)
// 	return res
// }

// func (c numCol) getRate(sr loader.StateReader) (result string) {
// 	// This sucks because currently Col has sources and a key, but that should:
// 	// a) allow a mechanism to specify a source/key in one attribute
// 	// b) allow different types of Cols to take one key, two keys, or a list of depending on what the calculation is.    This implies this function should be in a subclass
// 	// cur, prev := sr.GetKeyCurPrev

// 	// Apply precision
// 	// return fmt.Sprintf("%.2s")
// 	return "   5"
// }
