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

// Options The default configuration
type Options struct {
	FileMaxSize int    `json:"file_max_size,omitempty"`
	Path        string `json:"path,omitempty"`
}

func (o *Options) Validation() {

	if o.Path == "" {
		panic("The data file directory cannot be empty")
	}

	if o.FileMaxSize == 0 {
		o.FileMaxSize = defaultMaxFileSize
	}

	globalConfig = o
}

// SafeOptions Secure configuration
type SafeOptions struct {
	Options
	enable bool   `json:"enable,omitempty"`
	Secret []byte `json:"secret,omitempty"`
}

func (s *SafeOptions) Validation() {
	s.enable = true
	if len(s.Secret) < 16 {
		panic("The encryption key contains less than 16 characters")
	}

	secret = s.Secret
	globalConfig = s
}

// SetEncryptor Set up a custom encryption implementation
func SetEncryptor(encryptor Encryptor) {
	if encoder == nil {
		encoder = DefaultEncoder()
	}
	encoder.enable = true
	encoder.Encryptor = encryptor
}

// Configure Global configuration constraint interfaces
type Configure interface {
	Validation()
}
