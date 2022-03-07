// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/3/4 - 2:56 下午 - UTC/GMT+08:00

package bottle

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestAESEncryptor(t *testing.T) {
	aes := AES()
	data := &SourceData{
		Secret: []byte("1234567890123456"),
		Data:   []byte("TEST"),
	}
	aes.Encode(data)
	t.Log(data.Data)
	aes.Decode(data)
	t.Log(data.Data)
}

func TestEncoding(t *testing.T) {
	initialize()
	// 测试数据编码器
	encoder := DefaultEncoder()
	// 构建数据项
	item := NewItem([]byte("foo"), []byte("bar"), uint64(time.Now().Unix()))
	// 打开一个数据文件
	if file, err := os.OpenFile("./test.data", FRW, Perm); err == nil {
		num, err := encoder.Write(item, file)
		fileList[int64(HashedFunc.Sum64([]byte("foo")))] = file

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, num, 26)
	}

	HashedFunc = DefaultHashFunc()

	record := &record{
		FID:    int64(HashedFunc.Sum64([]byte("foo"))),
		Size:   26,
		Offset: 0,
	}

	data, _ := encoder.Read(record)

	assert.Equal(t, data.Value, item.Value)
}

func TestBinaryEncode(t *testing.T) {
	// 构建数据项
	item := NewItem([]byte("foo"), []byte("bar"), uint64(time.Now().Unix()))
	bytes := binaryEncode(item)
	t.Log(bytes)
	data := binaryDecode(bytes)
	t.Log(data.Log)
}
