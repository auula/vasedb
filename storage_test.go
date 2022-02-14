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
	"encoding/json"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {

	defer recoveryDir()

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
			if err := Open(tt.args.path); (err != nil) != tt.wantErr {
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

func TestPut(t *testing.T) {
	t.Log(Open("./testdata/"))
	t.Log(Put([]byte("foo"), []byte("bar"), TTL(1800)))
	user := &UserInfo{
		UserName: "Leon Ding",
		Email:    "ding@ibyte.me",
		Age:      21,
	}
	bytes, err := json.Marshal(user)
	if err != nil {
		return
	}
	t.Log(Put([]byte("userinfo"), bytes))
}

// recoveryDir clear temporary testing data
func recoveryDir() {
	os.RemoveAll("./testdata/")
	os.MkdirAll("./testdata/", perm)
}
