package loader

// Source to collect a Sample
type Source struct {
	Name        SourceName
	Description string
	// Needs some attributes that describe how to load this source, live or file
}

// A SourceName identifies some unique portion of data gathered from a Source
type SourceName string
