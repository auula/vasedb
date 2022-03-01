// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/28 - 10:14 下午 - UTC/GMT+08:00

// go:build ignore
//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/auula/bottle"
	"sync"
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

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for i := 0; i < 1000; i++ {
			if err := bottle.Put([]byte(fmt.Sprintf("user-%d", i)), bottle.Bson(&Userinfo{
				Name:  fmt.Sprintf("user-%d", i),
				Age:   22,
				Skill: []string{"Java", "Go", "Rust"},
			})); err != nil {
				panic(err)
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			var u Userinfo
			// 通过Unwrap解析出结构体
			bottle.Get([]byte(fmt.Sprintf("user-%d", i))).Unwrap(&u)
			// 打印取值
			fmt.Println(u)
		}
		wg.Done()
	}()

	wg.Wait()

	var u Userinfo
	// 通过Unwrap解析出结构体
	key := fmt.Sprintf("user-%d", 888)

	fmt.Println("FIND KEY IS:", key)

	bottle.Get([]byte(key)).Unwrap(&u)
	// 打印取值
	fmt.Println("FIND KEY VALUE IS:", u)

	fmt.Println(bottle.Close())
}
