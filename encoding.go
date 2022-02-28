// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/27 - 4:26 下午 - UTC/GMT+08:00

package bottle

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"time"
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
func (e *Encoder) Write(item *Item) (int, error) {

	// whether encryption is enabled
	if e.enable && e.Encryptor != nil {
		// building source data
		sd := &SourceData{
			Secret: Secret,
			Data:   item.Value,
		}
		if err := e.Encode(sd); err != nil {
			return 0, errors.New("an error occurred in the encryption encoder")
		}
		item.Value = sd.Data
		return bufToFile(BinaryEncode(item))
	}

	return bufToFile(BinaryEncode(item))
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

// bufToFile entity records are written from the buffer to the file
func bufToFile(data []byte) (int, error) {
	if n, err := active.Write(data); err == nil {
		return n, nil
	}
	return 0, errors.New("error writing encode buffer data to log")
}

func (e *Encoder) Read(rec *record) (*Item, error) {

	// 解析到数据实体
	item := parseLog(rec)

	if e.enable && e.Encryptor != nil && item != nil {
		// Decryption operation
		sd := &SourceData{
			Secret: Secret,
			Data:   item.Value,
		}
		if err := e.Decode(sd); err != nil {
			return nil, errors.New("a data decryption error occurred")
		}
		item.Value = sd.Data
		return item, nil
	}

	return item, nil
}

// parseLog parse data item from files
func parseLog(rec *record) *Item {
	// 通过记录文件标识符查找到这个文件
	if file, ok := fileList[rec.FID]; ok {
		// 截取数据段 尺寸窗口
		_, _ = file.Seek(int64(rec.Offset), 0)
		data := make([]byte, rec.Size)
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

	var item Item
	// | CRC 4 | TS 8  | KS 4 | VS 4  | KEY ? | VALUE ? |
	item.TimeStamp = binary.LittleEndian.Uint64(data[4:12])
	item.KeySize = binary.LittleEndian.Uint32(data[12:16])
	item.ValueSize = binary.LittleEndian.Uint32(data[16:20])
	item.CRC32 = binary.LittleEndian.Uint32(data[:4])

	// 解析数据
	item.Key, item.Value = make([]byte, item.KeySize), make([]byte, item.ValueSize)
	copy(item.Key, data[itemPadding:itemPadding+item.KeySize])
	copy(item.Value, data[itemPadding+item.KeySize:itemPadding+item.KeySize+item.ValueSize])
	return &item
}

// WriteIndex the index entry to the target file
func (e *Encoder) WriteIndex(item indexItem, file *os.File) (int, error) {

	// | CRC32 4 | IDX 8 | FID 8  | TS 4 | ET 4 | SZ 4 | OF 4 |
	buf := make([]byte, 36)

	binary.LittleEndian.PutUint64(buf[4:12], item.idx)
	binary.LittleEndian.PutUint64(buf[12:20], uint64(item.FID))
	binary.LittleEndian.PutUint32(buf[20:24], item.Timestamp)
	binary.LittleEndian.PutUint32(buf[24:28], item.ExpireTime)
	binary.LittleEndian.PutUint32(buf[28:32], item.Size)
	binary.LittleEndian.PutUint32(buf[32:36], item.Offset)

	binary.LittleEndian.PutUint32(buf[:4], crc32.ChecksumIEEE(buf[4:]))

	return file.Write(buf)
}

func (e *Encoder) ReadIndex(buf []byte) error {
	var (
		item indexItem
	)

	if binary.LittleEndian.Uint32(buf[:4]) != crc32.ChecksumIEEE(buf[4:]) {
		return errors.New("index record verification failed")
	}

	item.record = new(record)

	item.idx = binary.LittleEndian.Uint64(buf[4:12])
	item.FID = int64(binary.LittleEndian.Uint64(buf[12:20]))
	item.Timestamp = binary.LittleEndian.Uint32(buf[20:24])
	item.ExpireTime = binary.LittleEndian.Uint32(buf[24:28])
	item.Size = binary.LittleEndian.Uint32(buf[28:32])
	item.Offset = binary.LittleEndian.Uint32(buf[32:36])

	// Determine expiration date
	if uint32(time.Now().Unix()) <= item.ExpireTime {
		index[item.idx] = &record{
			FID:        item.FID,
			Size:       item.Size,
			Offset:     item.Offset,
			Timestamp:  item.Timestamp,
			ExpireTime: item.ExpireTime,
		}
	}

	return nil
}
