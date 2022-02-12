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

package bigmap

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
)

var (
	suffix = ".data"

	ErrPathIsExists         = errors.New("big map error: an empty path is illegal")
	ErrPathNotAvailable     = errors.New("big map error: the current directory path is unavailable")
	ErrCreateActiveFileFail = errors.New("big map error: failed to create a writable and readable active file")

	FileOnlyReadANDWrite = os.O_RDWR | os.O_APPEND | os.O_CREATE
)

type ActiveFile struct {
	FID string
	*os.File
}

func Open(path string) error {
	if path == "" {
		return ErrPathIsExists
	}
	if ok, err := PathExists(path); err != nil {
		return ErrPathNotAvailable
	} else if ok {
		// 文件夹存在
		// 1. 读取下面的索引文件，是否有索引文件查看
		// 2. 如果有索引，则回复到内存里面
	}

	// 文件夹不存在
	// 创建一个可写的文件 开始键索引
	if err := createActiveFile(path); err != nil {
		return ErrCreateActiveFileFail
	}
	// 从索引文件里面恢复数据文件，恢复内存索引
	return nil
}

// Close current active file
func Close() error {
	return currentActiveFile.Close()
}

func createActiveFile(path string) error {
	// 创建文件
	fid := uuid.NewString()
	filePath := fmt.Sprintf("%s/%s%s", path, fid, suffix)

	if file, err := os.OpenFile(filePath, FileOnlyReadANDWrite, 0666); err != nil {
		return err
	} else {
		// 可能需要加锁
		currentActiveFile.FID = fid
		currentActiveFile.File = file
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
