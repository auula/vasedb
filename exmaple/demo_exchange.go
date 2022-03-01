// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/28 - 10:14 下午 - UTC/GMT+08:00

package main

import (
	"fmt"
	"github.com/auula/bottle"
)

func init() {
	if err := bottle.Open(bottle.Option{
		Directory:       "./data",
		DataFileMaxSize: 10240 / 2,
	}); err != nil {
		panic(err)
	}

}

func main() {

	type Userinfo struct {
		Name  string
		Age   uint8
		Skill []string
	}

	for i := 0; i < 1000; i++ {
		bottle.Put([]byte(fmt.Sprintf("user-%d", i)), bottle.Bson(&Userinfo{
			Name:  fmt.Sprintf("user-%d", i),
			Age:   22,
			Skill: []string{"Java", "Go", "Rust"},
		}))
	}

	var u Userinfo
	// 通过Unwrap解析出结构体
	key := fmt.Sprintf("user-%d", 888)

	fmt.Println("FIND KEY IS:", key)

	bottle.Get([]byte(key)).Unwrap(&u)
	// 打印取值
	fmt.Println("FIND KEY VALUE IS:", u)

	bottle.Close()
}
