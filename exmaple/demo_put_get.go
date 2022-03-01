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

type Userinfo struct {
	Name  string
	Age   uint8
	Skill []string
}

func main() {

	//// PUT Data
	err := bottle.Put([]byte("foo"), []byte("test"))
	if err != nil {
		fmt.Println(err)
	}

	// 如果转成string那么就是字符串
	fmt.Println(bottle.Get([]byte("foo")).Value)

	// 如果不存在默认值就是0
	fmt.Println(bottle.Get([]byte("foo")).Int())

	// 如果不成功就是false
	fmt.Println(bottle.Get([]byte("foo")).Bool())

	// 如果不成功就是0.0
	fmt.Println(bottle.Get([]byte("foo")).Float())

	bottle.Close()
}
