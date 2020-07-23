package event

// Topic is the factory of event Iterator
type Topic interface {
	// ParseFilter 从指定的bytes buffer反序列化topic过滤器
	// 返回的参数会作为入参传递给NewIterator的filter参数
	ParseFilter(buf []byte) (interface{}, error)

	// MarshalEvent encode event payload returns from Iterator.Data()
	MarshalEvent(x interface{}) ([]byte, error)

	// NewIterator make a new Iterator base on filter
	NewIterator(filter interface{}) (Iterator, error)
}

// Iterator is the event iterator, must be closed after use
type Iterator interface {
	Next() bool
	Data() interface{}
	Error() error
	Close()
}
