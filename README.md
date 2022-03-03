<img align="right" src="gopher-bottle.svg" alt="bottle-kv-storage" width="180" height="180" />

# Bottle

Bottle is a lightweight kv storage engine based on a log structured Hash Table.

---

### 特 性

- 嵌入的存储引擎
- 数据可以加密存储
- 可以自定义实现存储加密器
- 即使数据文件被拷贝，也保证存储数据的安全
- 未来索引数据结构也可以支持自定义实现

---

### 介 绍

首先要说明的是`Bottle`是一款`KV`嵌入式存储引擎，并非是一款`KV`数据库，我知道很多人看到了`KV`认为是数据库，当然不是了，很多人会把这些搞混淆掉，`KV`
存储可以用来存储很多东西，而并非是数据库这一领域。可以这么理解数据库是一台汽车，那么`Bottle`是一台车的发动机。可以简单理解`Bottle`是一个对操作系统文件系统的`KV`抽象化封装，可以基于`Bottle`
做为存储层，在`Bottle`层之上封装一些数据结构和对外服务的存储协议就可以实现一个数据库。

![层次架构图](https://tva1.sinaimg.cn/large/e6c9d24egy1gzfrmt7qo4j21c20u0tai.jpg)


---

### 安装Bottle

你只需要在你的项目中安装`Bottle`模块即可使用：

```shell
go get -u github.com/auula/bottle
```

### 基本API

如何操作一个`bottle`实例代码:

```go
package main

import (
	"fmt"
	"github.com/auula/bottle"
)

func init() {
    // 打开存储引擎实例
	err := bottle.Open(bottle.Option{
		Directory:       "./data",
		DataFileMaxSize: 10240,
	})
	if err != nil {
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
	bottle.Put([]byte("foo"), []byte("66.6"))

	// 如果转成string那么就是字符串
	fmt.Println(bottle.Get([]byte("foo")).String())

	// 如果不存在默认值就是0
	fmt.Println(bottle.Get([]byte("foo")).Int())

	// 如果不成功就是false
	fmt.Println(bottle.Get([]byte("foo")).Bool())

	// 如果不成功就是0.0
	fmt.Println(bottle.Get([]byte("foo")).Float())

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

	// 删除一个key
	bottle.Remove([]byte("foo"))

	// 关闭处理一下可能发生的错误
	if err := bottle.Close(); err != nil {
		fmt.Println(err)
	}
}
```
