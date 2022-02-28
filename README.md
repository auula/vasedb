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

首先要说明的是`Bottle`是一款`KV`嵌入式存储引擎，并非是一款`KV`数据库，我知道很多人看到了`KV`认为是数据库，当然不是了，很多人会把这些搞混淆掉，`KV`存储可以用来存储很多东西，而并非是数据库这一领域。可以这么理解数据库是一台汽车，那么`Bottle`是一台车的发动机。可以简单理解`Bottle`是一个对操作系统文件系统的`KV`抽象化封装，可以基于`Bottle`做为存储层，在`Bottle`层之上封装一些数据结构和对外服务的存储协议就可以实现一个数据库。

![层次架构图](https://tva1.sinaimg.cn/large/e6c9d24egy1gzfrmt7qo4j21c20u0tai.jpg)

<!-- 为什么叫`Bottle`？我觉得这个项目的主要的功能就是存储，存储？存储不就是可以装东西，存储数据，所以我就联想到了`魔瓶`这个词，于是我想到了`Klein Bottle`，所以我就选择`Bottle`作为存储引擎的名字。

为什么有这个项目，事情是这样传统`CRUD`我已经写不再想写了，之前写`Java`是天天`CRUD`。由于我接触`Go`语言很久了，想用`Go`去做一些`Lab`，本人目前也对分布式存储这块感兴趣，然后看了一些论文，所以动机就有了，就花了一个星期左右把这个存储引擎用`Go`语言实现出来了，如果读者对这个项目感兴趣，本项目大部分理论知识和`CMU 15-445: Database Systems`这套课很接近，这门课由数据库领域的大牛`Andy Pavlo`讲授，感兴趣可以自己去看看这方面的资料吧。 -->

---
