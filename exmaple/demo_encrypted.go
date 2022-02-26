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

// Bottle Use demo file

// go:build ignore
//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/auula/bottle"
)

var bt *bottle.Storage

func init() {

	// open bottle storage engine
	bt, _ = bottle.Open(bottle.Options{
		Path:        "testdata",
		FileMaxSize: bottle.DefaultMaxFileSize,
	})

	bottle.SetEncryptor(bottle.AESEncryptor{}, bottle.DefaultSecret)
}

func main() {

	bt.Put([]byte("K"), []byte("V"))

	item, _ := bt.Get([]byte("K"))

	fmt.Println("K:", string(item.Value))

	bt.Close()
}
