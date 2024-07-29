package val

// ValueStruct represents the value info that can be associated with a key, but also the internal
// Meta field.
type ValueStruct struct {
	Meta      byte
	UserMeta  byte
	ExpiresAt uint64
	Value     []byte

	Version uint64 // This field is not serialized. Only for internal usage.
}