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

// encrypted.go
// Specific encoder AES encryption decryption implementation.

package bottle

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// SourceData for encryption and decryption
type SourceData struct {
	Data   []byte
	Secret []byte
}

// Encryptor used for data encryption and decryption operation
type Encryptor interface {
	Encode(sd *SourceData) error
	Decode(sd *SourceData) error
}

// AESEncryptor Implement the Encryptor interface
type AESEncryptor struct{}

// Encode source data encode
func (A AESEncryptor) Encode(sd *SourceData) error {
	sd.Data = aesEncrypt(sd.Data, sd.Secret)
	return nil
}

// Decode source data decode
func (A AESEncryptor) Decode(sd *SourceData) error {
	sd.Data = aesDecrypt(sd.Data, sd.Secret)
	return nil
}

// aesEncrypt ASE encode
func aesEncrypt(origData []byte, key []byte) []byte {
	// 分组秘钥
	block, _ := aes.NewCipher(key)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 创建数组
	result := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(result, origData)
	return result
}

// aesDecrypt  aes decode
func aesDecrypt(bytes []byte, key []byte) []byte {
	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 创建数组
	orig := make([]byte, len(bytes))
	// 解密
	blockMode.CryptBlocks(orig, bytes)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return orig
}

// PKCS7Padding 补码
func PKCS7Padding(ciphertext []byte, blksize int) []byte {
	padding := blksize - len(ciphertext)%blksize
	plaintext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, plaintext...)
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	padding := int(origData[length-1])
	return origData[:(length - padding)]
}
