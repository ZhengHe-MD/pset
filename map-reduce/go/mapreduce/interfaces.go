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

// ByKey is used to sort KeyValue.
type ByKey []KeyValue

func (b ByKey) Len() int           { return len(b) }
func (b ByKey) Less(i, j int) bool { return b[i].Key < b[j].Key }
func (b ByKey) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// Encoder is the interface that wraps the basic Encode method.
type Encoder interface {
	Encode(v any) error
}

// Decoder is the interface that wraps the basic Decode method.
type Decoder interface {
	Decode(v any) error
}
