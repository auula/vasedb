// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:32 下午 - UTC/GMT+08:00

package bottle

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

// Log the key value data
type Log struct {
	Key, Value []byte
}

// Item each data operation log item
// | TS 8 | CRC 4 | KS 4 | VS 4  | KEY ? | VALUE ? |
// ItemPadding = 8 + 12 = 20 bit
type Item struct {
	TimeStamp uint64 // Create timestamp
	CRC32     uint32 // Cyclic check code
	KeySize   uint32 // The size of the key
	ValueSize uint32 // The size of the value
	Log              // Key string, value serialization
}

// NewItem build a data log item
func NewItem(key, value []byte, timestamp uint64) *Item {
	return &Item{
		TimeStamp: timestamp,
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
		Log: Log{
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

// IsError return an error
func (d Data) IsError() bool {
	return d.Err != nil
}

func (d *Data) Unwrap(v interface{}) {
	if d.Item != nil {
		_ = bson.Unmarshal(d.Value, v)
	}
}

func (d Data) String() string {
	if d.Item != nil {
		return string(d.Value)
	}
	return ""
}

func (d Data) Int() int {
	if d.Item != nil {
		num, err := strconv.Atoi(string(d.Value))
		if err != nil {
			return 0
		}
		return num
	}
	return 0
}

func (d Data) Float() float64 {
	if d.Item != nil {
		num, err := strconv.ParseFloat(string(d.Value), 64)
		if err != nil {
			return 0
		}
		return num
	}
	return 0
}

func (d Data) Bool() bool {
	if d.Item != nil {
		b, err := strconv.ParseBool(string(d.Value))
		if err != nil {
			return b
		}
		return false
	}
	return false
}

func Bson(v interface{}) []byte {
	if v == nil {
		return nil
	}
	bytes, _ := bson.Marshal(v)
	return bytes
}
