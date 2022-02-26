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

package bottle

import (
	"gopkg.in/mgo.v2/bson"
	"os"
	"strings"
)

// Options The default configuration
type Options struct {
	FileMaxSize int64  `json:"file_max_size,omitempty"`
	Path        string `json:"path,omitempty"`
}

func (o *Options) Validation() {

	if o.Path == "" {
		panic("The data file directory cannot be empty")
	}

	// The first one does not determine whether there is a backslash
	o.Path = pathBackslashes(o.Path)

	// Record the location of the data file
	o.Path = strings.TrimSpace(o.Path)

	if o.FileMaxSize == 0 {
		o.FileMaxSize = DefaultMaxFileSize
	}

	globalOption = o
}

// SetEncryptor Set up a custom encryption implementation
func SetEncryptor(encryptor Encryptor, secret []byte) {
	if len(secret) < 16 {
		panic("The encryption key contains less than 16 characters")
	}
	encoder.enable = true
	encoder.Encryptor = encryptor
	DefaultSecret = secret
}

// Configure Global configuration constraint interfaces
type Configure interface {
	Validation()
}

type Hint struct {
	IndexPath  string `json:"index_path"`
	activeFile string `json:"active_file_path"`
}

func saveHint(indexPath string, filePath string) error {

	data, err := bson.Marshal(&Hint{
		IndexPath:  indexPath,
		activeFile: filePath,
	})

	if err != nil {
		return ErrSaveHintFileFail
	}

	if file, err := os.OpenFile(hintPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm); err == nil {
		_, err := file.Write(data)
		if err != nil {
			return ErrSaveHintFileFail
		}
	}

	return nil
}
