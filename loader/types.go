package loader

// Source to collect a Sample
type Source struct {
	Name        SourceName
	Description string
	Query       string
}

// A SourceName identifies some unique portion of data gathered from a Source
type SourceName string
