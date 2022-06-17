package loader

// Loads data from somewhere to produce samples
type MyqLoader interface {

	// Setup the loader and error-check
	Initialize() error

	// ProduceSamples
}
