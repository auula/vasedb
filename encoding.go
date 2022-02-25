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
func (e *Encoder) Write(entity *Item, file *activeFile) (int, error) {

	// whether encryption is enabled
	if e.enable && e.Encryptor != nil {
		// building source data
		sd := &SourceData{
			Secret: globalOption.Secret,
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

func (e *Encoder) Read(record *Record) (*Item, error) {

	// 解析到数据实体
	entity := parseEntity(record)

	if e.enable && e.Encryptor != nil && entity != nil {
		// Decryption operation
		sd := &SourceData{
			Secret: globalOption.Secret,
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
func parseEntity(record *Record) *Item {
	// 通过记录文件标识符查找到这个文件
	if file, ok := fileLists[record.FID]; ok {
		// 截取数据段 尺寸窗口
		_, _ = file.Seek(int64(record.Offset), 0)
		data := make([]byte, record.Size)
		_, _ = file.Read(data)
		return BinaryDecode(data)
	}
	return nil
}

// BinaryDecode you can parse  binary data into entity
func BinaryDecode(data []byte) *Item {
	// 校验crc32
	if binary.LittleEndian.Uint32(data[:4]) != crc32.ChecksumIEEE(data[4:]) {
		return nil
	}

	var entity Item
	// | CRC32 4 | TS 4 | KS 4 | VS 4  | KEY ? | VALUE ? |
	entity.TimeStamp = binary.LittleEndian.Uint32(data[4:8])
	entity.KeySize = binary.LittleEndian.Uint32(data[8:12])
	entity.ValueSize = binary.LittleEndian.Uint32(data[12:16])

	// 解析数据
	entity.Key, entity.Value = make([]byte, entity.KeySize), make([]byte, entity.ValueSize)
	copy(entity.Key, data[EntityPadding:EntityPadding+entity.KeySize])
	copy(entity.Value, data[EntityPadding+entity.KeySize:EntityPadding+entity.KeySize+entity.ValueSize])

	return &entity
}

// BinaryEncode you can parse an entity into binary slices
func BinaryEncode(entity *Item) []byte {

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
func bufToFile(data []byte, file *activeFile) (int, error) {
	if n, err := file.Write(data); err == nil {
		return n, nil
	}
	return 0, ErrEntityDataBufToFile
}

// WriteIndex the index entry to the target file
func (e *Encoder) WriteIndex(item indexItem, file *os.File) (int, error) {
	// | CRC32 4 | ET 4 | SZ 4  | OF 4 | IDX 8 |FID 8 | timestamp 4 |
	buf := make([]byte, 36)

	binary.LittleEndian.PutUint32(buf[4:8], item.ExpireTime)
	binary.LittleEndian.PutUint32(buf[8:12], item.Size)
	binary.LittleEndian.PutUint32(buf[12:16], item.Offset)
	binary.LittleEndian.PutUint64(buf[16:24], item.idx)

	// 这里如果是文件名长度是不统一的，用的是
	binary.LittleEndian.PutUint64(buf[24:32], item.fid)
	binary.LittleEndian.PutUint32(buf[32:36], item.Timestamp)

	binary.LittleEndian.PutUint32(buf[:4], crc32.ChecksumIEEE(buf[4:]))

	return file.Write(buf)
}

func (e *Encoder) ReadIndex(buf []byte, index map[uint64]*Record) error {

	// | CRC32 4 | ET 4 | SZ 4  | OF 4 | IDX 8 |FID 8 |
	var (
		item indexItem
	)

	if binary.LittleEndian.Uint32(buf[:4]) != crc32.ChecksumIEEE(buf[4:]) {
		return ErrRecoveryIndexFail
	}

	item.idx = binary.LittleEndian.Uint64(buf[16:24])
	item.fid = binary.LittleEndian.Uint64(buf[24:32])
	item.Size = binary.LittleEndian.Uint32(buf[8:12])
	item.Offset = binary.LittleEndian.Uint32(buf[12:16])
	item.Timestamp = binary.LittleEndian.Uint32(buf[32:36])
	item.ExpireTime = binary.LittleEndian.Uint32(buf[4:8])

	index[item.idx] = &Record{
		FID:        item.fid,
		Size:       item.Size,
		Offset:     item.Offset,
		Timestamp:  item.Timestamp,
		ExpireTime: item.ExpireTime,
	}

	return nil
}

// Save index files to the data directory
func saveIndexToFile(index map[uint64]*Record) error {

	var channel = make(chan indexItem, 1024)

	go func() {
		for sum64, record := range index {
			channel <- indexItem{
				fid:        record.FID,
				idx:        sum64,
				Size:       record.Size,
				Offset:     record.Offset,
				ExpireTime: record.ExpireTime,
			}
		}
		close(channel)
	}()

	if file, err := os.OpenFile(indexFilePath(globalOption.Path), fileOnlyReadANDWrite, perm); err != nil {
		return err
	} else {
		for v := range channel {
			_, encodeIndexErr := encoder.WriteIndex(v, file)
			if encodeIndexErr != nil {
				return ErrIndexEncode
			}
		}
	}

	return nil
}
