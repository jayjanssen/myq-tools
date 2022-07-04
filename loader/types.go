package loader

// Source to collect a Sample
type Source struct {
	Name        SourceKey
	Description string
	// Needs some attributes that describe how to load this source, live or file
}

// A SourceKey identifies some unique portion of data gathered from a Source
type SourceKey string
