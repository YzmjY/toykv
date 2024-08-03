package x

type Iterator interface {
	Next()
	Vaild() bool
	Rewind()
	Seek(key []byte)
	Key() []byte
	Value() ValueStruct
	Close()
}
