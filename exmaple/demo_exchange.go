// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/28 - 10:14 下午 - UTC/GMT+08:00

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/auula/bottle"
)

func init() {
	if err := bottle.Open(bottle.Option{
		Directory:       "./data",
		DataFileMaxSize: 10240,
	}); err != nil {
		panic(err)
	}
}

type Userinfo struct {
	Name  string
	Age   uint8
	Skill []string
}

func main() {

	var u Userinfo

	for i := 0; i < 999; i++ {
		if err := bottle.Put([]byte(fmt.Sprintf("user-%d", i)), bottle.Bson(&Userinfo{
			Name:  fmt.Sprintf("user-%d", i),
			Age:   22,
			Skill: []string{"Java", "Go", "Rust"},
		})); err != nil {
			panic(err)
		}
	}

	// 通过Unwrap解析出结构体
	bottle.Get([]byte("user-998")).Unwrap(&u)
	// 打印取值
	fmt.Println(u)

	bottle.Close()
}
