// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/27 - 4:26 下午 - UTC/GMT+08:00

package bottle

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
)

// Encoder bytes data encoder
type Encoder struct {
	Encryptor      // encryptor concrete implementation
	enable    bool // whether to enable data encryption and decryption
}

// AESEncoder enable the AES encryption encoder
func AESEncoder() *Encoder {
	return &Encoder{
		enable:    true,
		Encryptor: new(AESEncryptor),
	}
}

// DefaultEncoder disable the AES encryption encoder
func DefaultEncoder() *Encoder {
	return &Encoder{
		enable:    false,
		Encryptor: nil,
	}
}

// Write to entity's current activation file
func (e *Encoder) Write(item *Item, file *os.File) (int, error) {

	// whether encryption is enabled
	if e.enable && e.Encryptor != nil {
		// building source data
		sd := &SourceData{
			Secret: defaultSecret,
			Data:   item.Value,
		}
		if err := e.Encode(sd); err != nil {
			return 0, errors.New("an error occurred in the encryption encoder")
		}
		item.Value = sd.Data
		return 0, nil
	}

	return 0, nil
}

// BinaryEncode you can parse an entity into binary slices
func BinaryEncode(item *Item) []byte {

	buf := make([]byte, itemPadding+item.KeySize+item.ValueSize)

	// | CRC 4 | TS 8  | KS 4 | VS 4  | KEY ? | VALUE ? |
	// ItemPadding = 8 + 12 = 20 bit
	binary.LittleEndian.PutUint64(buf[4:12], item.TimeStamp)
	binary.LittleEndian.PutUint32(buf[12:16], item.KeySize)
	binary.LittleEndian.PutUint32(buf[16:20], item.ValueSize)

	//buf = append(buf, item.Key...)
	//buf = append(buf, item.Value...)

	// add key data to the buffer
	copy(buf[itemPadding:itemPadding+item.KeySize], item.Key)
	// add value data to the buffer
	copy(buf[itemPadding+item.KeySize:itemPadding+item.KeySize+item.ValueSize], item.Value)

	// add crc32 code to the buffer
	binary.LittleEndian.PutUint32(buf[:4], crc32.ChecksumIEEE(buf[4:]))

	return buf
}
