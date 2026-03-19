# Redis String

`String` 是 Redis 最常用的数据类型，适合存储以下内容：

- 普通字符串，如用户名、状态值
- 数字计数器，如浏览量、点赞数
- JSON 字符串，如缓存对象、配置信息
- 带过期时间的缓存值

## 1. 基础读写

设置一个字符串值：

```redis
SET user:1001:name Alice
```

返回：

```text
OK
```

读取这个值：

```redis
GET user:1001:name
```

返回：

```text
Alice
```

删除后再次读取：

```redis
DEL user:1001:name
GET user:1001:name
```

返回：

```text
(integer) 1
(nil)
```

重新写入一个新值：

```redis
SET user:1001:name Bob
GET user:1001:name
```

返回：

```text
OK
Bob
```

## 2. 存储数字

Redis 的 `String` 也可以存数字，但底层仍然是字符串表示：

```redis
SET counter 100
GET counter
```

返回：

```text
OK
"100"
```

## 3. 存储 JSON 字符串

业务中常用 `String` 直接缓存一段 JSON：

```redis
SET tokeninfo:1:0x123 "{\"name\":\"USDT\",\"symbol\":\"USDT\",\"decimals\":18}"
GET tokeninfo:1:0x123
```

返回：

```text
OK
"{\"name\":\"USDT\",\"symbol\":\"USDT\",\"decimals\":18}"
```

## 4. 设置过期时间

可以用 `SETEX` 一次性写入并设置过期时间：

```redis
SETEX cache:user:1001 10 cached_data
GET cache:user:1001
```

返回：

```text
OK
"cached_data"
```

10 秒后再次读取：

```redis
GET cache:user:1001
```

返回：

```text
(nil)
```

## 5. 仅在不存在时写入

`SETNX` 常用于简单加锁或防重复写入：

```redis
SETNX lock:order:1001 locked
SETNX lock:order:1001 locked
DEL lock:order:1001
```

返回：

```text
(integer) 1
(integer) 0
(integer) 1
```

更推荐直接使用 `SET key value NX EX seconds`，一次完成“仅在不存在时设置 + 过期时间”：

```redis
SET tokeninfo:1:0x789 token_data NX EX 30
GET tokeninfo:1:0x789
SET tokeninfo:1:0x789 token_data NX EX 30
```

返回：

```text
OK
"token_data"
(nil)
```

## 6. 计数器操作

`INCR` / `DECR` 适合实现浏览量、库存变化、调用次数统计等场景：

```redis
INCR counter:page:views
INCR counter:page:views
INCR counter:page:views
INCRBY counter:page:views 5
DECR counter:page:views
DECRBY counter:page:views 5
```

返回：

```text
(integer) 1
(integer) 2
(integer) 3
(integer) 8
(integer) 7
(integer) 2
```

## 7. 小结

`String` 适合做以下几类数据：

- 单个文本值
- 数字计数器
- JSON 缓存
- 分布式锁的简单标记
- 带 TTL 的临时缓存

## 8. String 的底层原理

从使用上看，Redis 是 `key -> value` 的存储；从实现上看，`String` 是最基础的一种值类型。

可以先把一条数据理解成这样：

```text
user:1001:name -> "Alice"
counter:page:views -> 100
```

其中：

- 左边是 `key`
- 右边是 `value`
- `value` 的类型是 Redis 的 `String`

### 8.1 为什么 String 很快

Redis 把大部分数据放在内存里操作，`GET` / `SET` 不需要像传统数据库那样先查磁盘页，所以单次读写非常快。

同时，Redis 的单线程命令执行模型避免了很多复杂的锁竞争。对 `String` 这类简单操作来说，一条命令通常就是：

1. 根据 `key` 找到对应的数据位置
2. 读取或覆盖 `value`
3. 返回结果

这也是为什么 `GET`、`SET`、`INCR` 这类命令通常都很高效。

### 8.2 String 存的并不只是“文本”

虽然名字叫 `String`，但它本质上是一段二进制安全的数据，也就是说它可以存：

- 普通文本
- 数字
- JSON
- 序列化后的对象
- 二进制内容

所以 Redis 的 `String` 更准确地说，是“最通用的值类型”。

### 8.3 数字为什么可以直接 INCR

当一个 `String` 的内容是整数时，Redis 可以把它按整数语义处理，所以像下面这样：

```redis
SET counter 100
INCR counter
```

Redis 会直接做数值加一，再把结果更新回去，因此 `INCR`、`DECR`、`INCRBY` 很适合实现计数器。

### 8.4 大字符串也能存，但不一定合适

Redis 的 `String` 单个 value 可以很大，但工程上通常不建议把超大的文本或大对象直接塞进去。

原因主要有两个：

- 单个 key 过大，会增加网络传输和内存占用成本
- 更新大对象时，通常需要整块读写，不如拆分结构灵活

所以：

- 小型文本、数字、JSON 缓存，很适合 `String`
- 字段很多的对象，通常更适合 `Hash`
- 特别大的内容，通常更适合文件存储或对象存储

## 9. Redis 的 KV 存储原理

Redis 常被称为 KV 数据库，意思是它最核心的模型就是：

```text
key -> value
```

例如：

```text
name -> "Alice"
counter -> 100
tokeninfo:1:0x123 -> "{\"name\":\"USDT\"}"
```

### 9.1 key 是怎么找到 value 的

Redis 在内部维护了一张哈希表，你可以把它理解成一个大字典：

```text
{
  "user:1001:name": "Alice",
  "counter:page:views": 8,
  "tokeninfo:1:0x123": "{\"name\":\"USDT\"}"
}
```

当执行：

```redis
GET user:1001:name
```

Redis 会先对 `key` 做哈希计算，再快速定位到对应的位置，然后拿到 `value`。这也是 Redis 按 key 查询很快的核心原因。

### 9.2 value 不只是字符串字面量

KV 里的 `V` 不一定只是普通文本，而是“某一种 Redis 数据类型的值”。

例如：

- `String` 用来存文本、数字、JSON
- `Hash` 用来存对象字段
- `List` 用来存列表
- `Set` 用来存无序集合
- `ZSet` 用来存带分数的有序集合

所以 Redis 虽然叫 KV，但它其实是“key 对应一个具备数据结构特性的 value”。

### 9.3 过期时间也是 KV 系统的一部分

如果一个 key 设置了 TTL，Redis 除了保存 `key -> value` 关系外，还会额外记录这个 key 的过期时间。

例如：

```redis
SETEX cache:user:1001 10 cached_data
```

除了存下：

```text
cache:user:1001 -> "cached_data"
```

还会记录这个 key 在 10 秒后过期。到期后，Redis 会通过惰性删除或定期清理，把它从内存中移除。

### 9.4 为什么 Redis 的 KV 适合做缓存

Redis 的 KV 模型非常适合缓存，核心原因是：

- 按 key 读取非常快
- value 可以直接存业务结果
- 支持过期时间
- 支持原子操作，比如 `INCR`、`SET NX EX`

因此它特别适合：

- 用户信息缓存
- 热点数据缓存
- 页面计数器
- 登录态或验证码临时存储
- 分布式锁的简单场景

## 10. 最后总结

可以把 Redis `String` 理解成 Redis 里最基础、最通用的 value 类型；可以把 Redis 本身理解成一个高性能的内存 KV 系统。

最常见的一条链路就是：

1. 业务代码传入一个 `key`
2. Redis 在内部字典里找到这个 `key`
3. 取出对应的 `value`
4. 如果这个 `value` 是 `String`，就按字符串、数字或二进制数据来处理

所以学习 Redis，最先掌握的就是两件事：

- `String` 适合存什么
- Redis 是如何通过 `key` 快速找到 `value` 的

如果数据需要更复杂的结构，再考虑 `Hash`、`List`、`Set`、`ZSet` 等类型。
