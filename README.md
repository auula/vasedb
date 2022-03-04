<img align="right" src="gopher-bottle.svg" alt="bottle-kv-storage" width="180" height="180" />

# Bottle

Bottle is a lightweight kv storage engine based on a log structured Hash Table.

[![Go Report Card](https://goreportcard.com/badge/github.com/auula/bottle)](https://goreportcard.com/report/github.com/auula/bottle)
[![Release](https://img.shields.io/github/v/release/auula/bottle.svg?style=flat-square)](https://github.com/auula/bottle)
[![License](https://img.shields.io/badge/license-MIT-db5149.svg)](https://github.com/auula/bottle/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/auula/bottle.svg)](https://pkg.go.dev/github.com/auula/bottle)
[![codecov](https://codecov.io/gh/auula/bottle/branch/dev/graph/badge.svg?token=btbed5BUUZ)](https://codecov.io/gh/auula/bottle)

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
做为存储层，在`Bottle`层之上封装一些数据结构和对外服务的协议就可以实现一个数据库。

![层次架构图](https://tva1.sinaimg.cn/large/e6c9d24egy1gzfrmt7qo4j21c20u0tai.jpg)

本项目功能实现完全基于 [`bitcask`](https://blog.ibyte.me/post/bitcask-kvbase/) 论文所实现，另外本项目所用到一些知识和`卡内基梅隆大学`的`CMU 15-445: Database Systems
`课程内容很接近，这门课由数据库领域的大牛`Andy Pavlo`讲授，有感兴趣的朋友可以去看看这套课，如果觉得不错你可以给我按一颗小星`⭐`谢谢。

---

### 安装Bottle

你只需要在你的项目中安装`Bottle`模块即可使用：

```shell
go get -u github.com/auula/bottle
```

### 基本API

如何操作一个`Bottle`实例代码:

```go
package main

import (
	"fmt"
	"github.com/auula/bottle"
)

func init() {
	// 通过默认配置打开一个存储实例
	err := bottle.Open(bottle.DefaultOption)
	// 并且处理一下可能发生的错误
	if err != nil {
		panic(err)
	}
}

// Userinfo 测试数据结构
type Userinfo struct {
	Name  string
	Age   uint8
	Skill []string
}

func main() {

	// PUT Data
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

	// 通过Bson保存数据对象，并且设置超时时间为5秒，TTL超时可以不设置看需求
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

### 加密器

下面例子是通过`bottle.SetEncryptor(encryptor Encryptor, secret []byte)`函数去设置数据加密器并且配置16位的数据加密秘钥。

```go
func init() {
    bottle.SetEncryptor(bottle.AES(), []byte("1234567890123456"))
}
```

你也可以自定义去实现数据加密器的接口：

```go
// SourceData for encryption and decryption
type SourceData struct {
    Data   []byte
    Secret []byte
}

// Encryptor used for data encryption and decryption operation
type Encryptor interface {
    Encode(sd *SourceData) error
    Decode(sd *SourceData) error
}
```

下面代码就是内置`AES`加密器的实现代码，实现`Encryptor`接口即可，数据源为`SourceData`结构体字段：

```go
// AESEncryptor Implement the Encryptor interface
type AESEncryptor struct{}

// Encode source data encode
func (AESEncryptor) Encode(sd *SourceData) error {
    sd.Data = aesEncrypt(sd.Data, sd.Secret)
    return nil
}

// Decode source data decode
func (AESEncryptor) Decode(sd *SourceData) error {
    sd.Data = aesDecrypt(sd.Data, sd.Secret)
    return nil
}
```
具体的加密器实现代码可以查看[`encrypted.go`](./encrypted.go)


### 散列函数

如果你需要自定义实现散列函数，实现`Hashed`接口即可：

```go
type Hashed interface {
    Sum64([]byte) uint64
}
```

然后通过内置的[`bottle.SetHashFunc(hash Hashed)`](./option.go)设置即可完成你的散列函数配置。

### 索引大小

索引预设置的大小很大程度上会影响你的程序存取和读取数据的速度，如果在初始化的时候能够预计出程序运行时需要的索引大小，并且在初始化的时候配置好，可以减小程序在运行过程中带来的运行数据迁移和扩容带来的性能问题。

```go
func init() {
	// 设置索引大小 
	bottle.SetIndexSize(1000)
}
```

### 配置信息

你也可以不使用默认配置，你可以使用内置的`bottle.Option`的结构体初始化你存储引擎，配置实例如下：

```go
func init() {
    // 自定义配置信息
    option := bottle.Option{
        // 工作目录
        Directory:       "./data",
        // 算法开启加密
        Enable:          true,
        // 自定义秘钥，可以使用内置的秘钥
        Secret:          bottle.Secret,
        // 自定义数据大小，存储单位是kb
        DataFileMaxSize: 1048576,
    }
    // 通过自定义配置信息
    bottle.Open(option)
}
```
当然也可以使用内置的[`bottle.Load(path string)`](./bottle.go)函数加载配置文件启动`Bottle`，配置文件格式为`yaml`，可配置项如下：

```yaml
# Bottle config options
Enable: TRUE
Secret: "1234567890123456"
Directory: "./testdata"
DataFileMaxSize: 536870912
```
需要注意的是内置的加密器实现的秘钥必须是`16`位，如果你是自定义实现的加密器可通过[`bottle.SetEncryptor(Encryptor,[]byte)`](./option.go)设置你自定义的加密器，那这个秘钥位数将不受限制。

