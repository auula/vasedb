## Bottle

Implement KV storage engine based on Golang.

### 介 绍

首先要说明的是`Bottle`是一款`KV`嵌入式存储引擎，并非是一款`KV`数据库，可以这么理解数据库是一台汽车，那么`Bottle`是一台车的发动机。可以基于`Bottle`做为存储层，在`Bottle`层之上封装一些数据结构和对外服务的存储协议就可以实现一个数据库。

![层次架构图](https://tva1.sinaimg.cn/large/e6c9d24egy1gzfrmt7qo4j21c20u0tai.jpg)

为什么叫`Bottle`？我觉得这个项目的主要的功能就是存储，存储？存储不就是可以装东西，存储数据，所以我就联想到了`魔瓶`这个词，于是我想到了`Klein Bottle`，所以我就选择`Bottle`作为存储引擎的名字。

### 特 性

- 嵌入的存储引擎
- 数据可以加密存储
- 可以自定义实现存储加密器
- 即使数据文件被拷贝，也保证存储数据的安全
- 未来索引数据结构也可以支持自定义实现

