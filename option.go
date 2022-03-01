// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:47 下午 - UTC/GMT+08:00

package bottle

import (
	"fmt"
	"os"
	"strings"
)

// Option bottle setting option
type Option struct {
	Directory       string `yaml:"directory"`          // data directory
	DataFileMaxSize int64  `yaml:"data_file_max_size"` // data file max size
}

var (
	// DefaultOption default initialization option
	DefaultOption = Option{
		Directory:       "./testdata",
		DataFileMaxSize: 10240,
	}
)

func (o *Option) Validation() {

	if o.Directory == "" {
		panic("The data file directory cannot be empty")
	}

	// The first one does not determine whether there is a backslash
	o.Directory = pathBackslashes(o.Directory)

	// Record the location of the data file
	o.Directory = strings.TrimSpace(o.Directory)

	Root = o.Directory

	dataDirectory = fmt.Sprintf("%sdata/", Root)

	indexDirectory = fmt.Sprintf("%sindex/", Root)

	if o.DataFileMaxSize != 0 {
		defaultMaxFileSize = o.DataFileMaxSize
	}

}

// SetEncryptor Set up a custom encryption implementation
func SetEncryptor(encryptor Encryptor, secret []byte) {
	if len(secret) < 16 {
		panic("The encryption key contains less than 16 characters")
	}
	encoder.enable = true
	encoder.Encryptor = encryptor
	Secret = secret
}

// pathBackslashes Check directory ending backslashes
func pathBackslashes(path string) string {
	if !strings.HasSuffix(path, "/") {
		return fmt.Sprintf("%s/", path)
	}
	return path
}

// Checks whether the target path exists
// 如果返回的错误为nil,说明文件或文件夹存在
// 如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
// 如果返回的错误为其它类型,则不确定是否在存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
