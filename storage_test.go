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
	"fmt"
	"hash/crc32"
	"os"
	"sync"
	"testing"
)

func TestOpen(t *testing.T) {

	//defer recoveryDir()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"failure", args{path: "./testdata/xxx/"}, false},
		{"successful", args{path: "./testdata/"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Open(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

type UserInfo struct {
	UserName string `json:"user_name,omitempty"`
	Email    string `json:"email,omitempty"`
	Age      uint8  `json:"age,omitempty"`
}

func TestPutANDGet(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	storage, _ := Open("./testdata/")

	signal := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			storage.Put([]byte(fmt.Sprintf("foo-%d", i)), []byte(fmt.Sprintf("bar-%d", i)))
		}
		wg.Done()
		signal <- struct{}{}
	}()

	go func() {
		<-signal
		for i := 0; i < 10; i++ {
			entity, _ := storage.Get([]byte(fmt.Sprintf("foo-%d", i)))
			t.Log(string(entity.Value))
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestRemove(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	storage, _ := Open("./testdata/")

	signal := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			storage.Put([]byte(fmt.Sprintf("foo-%d", i)), []byte(fmt.Sprintf("bar-%d", i)))
		}
		wg.Done()
		signal <- struct{}{}
	}()

	go func() {
		<-signal
		for i := 0; i < 10; i++ {
			storage.Remove([]byte(fmt.Sprintf("foo-%d", i)))
		}
		wg.Done()
	}()
	wg.Wait()
	t.Log("bottle storage engine data clear successful:", len(storage.index) == 0)
}

func TestBitOperation(t *testing.T) {
	t.Log(defaultMaxFileSize)
}

func TestFileSeek(t *testing.T) {
	file, _ := openOnlyReadFile("./testdata/test.md")
	file.Seek(3, 0)
	buf := make([]byte, 3)
	file.Read(buf)
	t.Log(string(buf))
}

func TestCRC32(t *testing.T) {
	data := []byte("HELLO")
	ieee := crc32.ChecksumIEEE(data)
	t.Log("ChecksumIEEE:", ieee)
	table := crc32.MakeTable(ieee)
	t.Log("MakeTable:", table)
	t.Log(crc32.Checksum(data, table))
}

// recoveryDir clear temporary testing data
func recoveryDir() {
	os.RemoveAll("./testdata/")
	os.MkdirAll("./testdata/", perm)
}
