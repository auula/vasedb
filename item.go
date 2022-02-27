// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:32 下午 - UTC/GMT+08:00

package bottle

// Value the key value data
type Value struct {
	Key, Value []byte
}

// Item each data operation log item
type Item struct {
	TimeStamp uint64 // Create timestamp
	CRC32     uint32 // Cyclic check code
	KeySize   uint32 // The size of the key
	ValueSize uint32 // The size of the value
	Value            // Key string, value serialization
}

// NewItem build a data log item
func NewItem(key, value []byte, timestamp uint64) *Item {
	return &Item{
		TimeStamp: timestamp,
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
		Value: Value{
			Key:   key,
			Value: value,
		},
	}
}

// Data returns to the upper-level data item
type Data struct {
	Err error
	*Item
}

// buildData build upper-level data item
func buildData(err error, item *Item) *Data {
	var data Data
	if err != nil {
		data.Err = err
		return &data
	}
	data.Item = item
	return &data
}

// Value data value
func (d Data) Value() []byte {
	return d.Item.Value.Value
}

// Key data key
func (d Data) Key() []byte {
	return d.Item.Value.Key
}

// Error return an error
func (d Data) Error() error {
	return d.Err
}
