// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/28 - 7:07 下午 - UTC/GMT+08:00

package main

import (
	"fmt"
	"github.com/auula/bottle"
)

func init() {
	if err := bottle.Open(bottle.DefaultOption); err != nil {
		panic(err)
	}
}

func main() {

	// PUT Data
	bottle.Put([]byte("foo"), []byte("bar"))

	// GET Data
	fmt.Println(bottle.Index[15902901984413996407])

	fmt.Println(bottle.Get([]byte("foo")))

	bottle.Close()
}
