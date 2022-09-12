package mr

// Mapper is the interface that wraps the basic Map method.
type Mapper interface {
	Map(data []byte) ([]KeyValue, error)
}

// Reducer is the interface that wraps the basic Reduce method.
type Reducer interface {
	Reduce(key string, values []string) (string, error)
}

// MapReducer is the interface that groups the basic Map and Reduce methods.
type MapReducer interface {
	Mapper
	Reducer
}

// KeyValue is the data unit returned by any Mapper.
type KeyValue struct {
	Key   string
	Value string
}

// Encoder is the interface that wraps the basic Encode method.
type Encoder interface {
	Encode(v any) error
}

// Decoder is the interface that wraps the basic Decode method.
type Decoder interface {
	Decode(v any) error
}
