package loader

import "time"

// The current and most recent SampleSets
type MyqState struct {
	Current, Previous *MyqSampleSet
}

// A collection of Samples at a given time
type MyqSampleSet struct {
	Timestamp time.Time
	Samples   map[MyqSchemaName]*MyqSample
}

// The values for a Schema for a specifc time
type MyqSample struct {
	Timestamp time.Time
	Schema    *MyqSchema
	Data      map[MyqSchemaKey]interface{}
}

// The name of a schema
type MyqSchemaName string

// The key identifier in a schema
type MyqSchemaKey string

// The type of a SchemaKey
type MyqSchemaType int

const (
	INT MyqSchemaType = iota
	FLOAT
	STRING
)

// A set of keys and types that we would fetch from a Source for a given interval to produce a Sample.
type MyqSchema struct {
	Name   MyqSchemaName
	Keys   map[MyqSchemaKey]MyqSchemaType
	Loader *MyqLoader
}
