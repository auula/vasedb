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
	//bottle.Put([]byte("foo"), []byte("66.6"))
	//
	//fmt.Println(bottle.Get([]byte("foo")).String())
	//fmt.Println(bottle.Get([]byte("foo")).Int())
	//fmt.Println(bottle.Get([]byte("foo")).Bool())
	//fmt.Println(bottle.Get([]byte("foo")).Float())

	user := Userinfo{
		Name:  "Leon Ding",
		Age:   22,
		Skill: []string{"Java", "Go", "Rust"},
	}

	var u Userinfo

	// 通过Bson保存数据对象,并且设置超时时间为5秒
	bottle.Put([]byte("user"), bottle.Bson(&user), bottle.TTL(5))

	// 通过Unwrap解析出结构体
	bottle.Get([]byte("user")).Unwrap(&u)
	// 打印取值
	fmt.Println(u)

	bottle.Close()
}
