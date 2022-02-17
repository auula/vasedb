// MIT License

// Copyright (c) 2022 Leon Ding

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// encoding.go
// File data encryption and decryption encoder implementation.

package bottle

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
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
func (e *Encoder) Write(entity *Entity, file *ActiveFile) (int, error) {

	// whether encryption is enabled
	if e.enable && e.Encryptor != nil {
		// building source data
		sd := &SourceData{
			Secret: secret,
			Data:   entity.Value,
		}
		if err := e.Encode(sd); err != nil {
			return 0, ErrSourceDataEncodeFail
		}
		entity.Value = sd.Data
		return bufToFile(BinaryEncode(entity), file)
	}

	return bufToFile(BinaryEncode(entity), file)
}

func (e *Encoder) Read(record *Record) (*Entity, error) {

	// 解析到数据实体
	entity := parseEntity(record)

	if e.enable && e.Encryptor != nil && entity != nil {
		// Decryption operation
		sd := &SourceData{
			Secret: secret,
			Data:   entity.Value,
		}
		if err := e.Decode(sd); err != nil {
			return nil, ErrSourceDataDecodeFail
		}
		entity.Value = sd.Data
		return entity, nil
	}

	return nil, ErrNoDataEntityWasFound
}

// parseEntity parse data entities from files
func parseEntity(record *Record) *Entity {
	// 通过记录文件标识符查找到这个文件
	if file, ok := fileLists[record.FID]; ok {
		// 截取数据段 尺寸窗口
		file.Seek(int64(record.Offset), 0)
		data := make([]byte, record.Size)
		file.Read(data)
		return BinaryDecode(data)
	}
	return nil
}

// dfp
func dfp(fid string) string {
	return fmt.Sprintf("%s%s%s", dataPath, fid, dataFileSuffix)
}

// BinaryDecode you can parse  binary data into entity
func BinaryDecode(data []byte) *Entity {
	var entity Entity

	// | CRC32 4 | TS 4 | KS 4 | VS 4  | KEY ? | VALUE ? |
	entity.TimeStamp = binary.LittleEndian.Uint32(data[4:8])
	entity.KeySize = binary.LittleEndian.Uint32(data[8:12])
	entity.ValueSize = binary.LittleEndian.Uint32(data[12:16])
	// 校验crc32
	if binary.LittleEndian.Uint32(data[:4]) != crc32.ChecksumIEEE(data[4:]) {
		return nil
	}

	// 解析数据
	entity.Key, entity.Value = make([]byte, entity.KeySize), make([]byte, entity.ValueSize)
	copy(entity.Key, data[EntityPadding:EntityPadding+entity.KeySize])
	copy(entity.Value, data[EntityPadding+entity.KeySize:EntityPadding+entity.KeySize+entity.ValueSize])

	return &entity
}

// BinaryEncode you can parse an entity into binary slices
func BinaryEncode(entity *Entity) []byte {

	buf := make([]byte, bufferSize(entity.Key, entity.Value))
	ks, vs := len(entity.Key), len(entity.Value)

	// | CRC32 4 | TS 4 | KS 4 | VS 4  | KEY ? | VALUE ? |
	binary.LittleEndian.PutUint32(buf[4:8], entity.TimeStamp) // PUT TS
	binary.LittleEndian.PutUint32(buf[8:12], uint32(ks))      // PUT KS
	binary.LittleEndian.PutUint32(buf[12:16], uint32(vs))     // PUT VS

	// add key data to the buffer
	copy(buf[EntityPadding:EntityPadding+ks], entity.Key)
	// add value data to the buffer
	copy(buf[EntityPadding+ks:EntityPadding+ks+vs], entity.Value)
	// add crc32 code to the buffer
	binary.LittleEndian.PutUint32(buf[:4], crc32.ChecksumIEEE(buf[4:]))

	return buf
}

// bufferSize calculate buffer size
func bufferSize(key, value []byte) int {
	return EntityPadding + len(key) + len(value)
}

// bufToFile entity records are written from the buffer to the file
func bufToFile(data []byte, file *ActiveFile) (int, error) {
	if n, err := file.Write(data); err == nil {
		return n, nil
	}
	return 0, ErrEntityDataBufToFile
}
