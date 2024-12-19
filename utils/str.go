package utils

import (
	"math/rand"
)

const Charset = "#$@!abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ123456789"

// RandomString 返回指定长度的字符串
func RandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = Charset[rand.Intn(len(Charset))]
	}
	return string(result)
}
