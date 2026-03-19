# Redis Hash

`Hash` 适合存储“一个对象的多个字段”，比如用户信息、商品信息、订单信息。

如果说 `String` 适合存单个值，那么 `Hash` 更适合存一组有结构的字段。

例如，一个用户对象可以这样表示：

```text
user:1001 -> {
  name: "Alice",
  age: "28",
  email: "alice@example.com"
}
```

## 1. Hash 是什么

可以把 Redis `Hash` 理解成：

```text
key -> field -> value
```

也就是说：

- 最外层还是一个 Redis key
- key 对应的 value 不是单个字符串
- 而是一组字段和值的映射关系

例如：

```text
user:1001 -> {
  name: "Alice",
  age: "28",
  city: "Shanghai"
}
```

这里：

- `user:1001` 是 key
- `name`、`age`、`city` 是 field
- `"Alice"`、`"28"`、`"Shanghai"` 是 field 对应的 value

## 2. 基础写入和读取

写入一个字段：

```redis
HSET user:1001 name Alice
```

返回：

```text
(integer) 1
```

读取这个字段：

```redis
HGET user:1001 name
```

返回：

```text
"Alice"
```

## 3. 一次写入多个字段

可以一次性给一个对象写多个字段：

```redis
HSET user:1002 name Bob age 30 email bob@example.com
```

返回：

```text
(integer) 3
```

分别读取：

```redis
HGET user:1002 name
HGET user:1002 age
HGET user:1002 email
```

返回：

```text
"Bob"
"30"
"bob@example.com"
```

## 4. 读取全部字段

如果想一次取出整个对象，可以使用 `HGETALL`：

```redis
HGETALL user:1002
```

返回：

```text
1) "name"
2) "Bob"
3) "age"
4) "30"
5) "email"
6) "bob@example.com"
```

如果只想看有哪些字段，可以用：

```redis
HKEYS user:1002
```

返回：

```text
1) "name"
2) "age"
3) "email"
```

如果只想看所有值，可以用：

```redis
HVALS user:1002
```

返回：

```text
1) "Bob"
2) "30"
3) "bob@example.com"
```

## 5. 判断字段是否存在

```redis
HEXISTS user:1002 age
HEXISTS user:1002 phone
```

返回：

```text
(integer) 1
(integer) 0
```

这适合用来判断某个对象字段是否已经写入。

## 6. 删除字段

删除某个字段：

```redis
HDEL user:1002 email
HGET user:1002 email
```

返回：

```text
(integer) 1
(nil)
```

如果把所有 field 都删掉，这个 hash 对应的 key 最终也会消失。

## 7. 获取字段数量

```redis
HLEN user:1002
```

返回：

```text
(integer) 2
```

这个命令适合查看一个对象当前有多少个字段。

## 8. 对字段做数值累加

如果某个 field 存的是数字，可以直接做自增：

```redis
HSET article:1001 stats:views 100
HINCRBY article:1001 stats:views 1
HINCRBY article:1001 stats:views 20
HGET article:1001 stats:views
```

返回：

```text
(integer) 1
(integer) 101
(integer) 121
"121"
```

这很适合统计对象内部某个字段，比如：

- 文章阅读量
- 用户积分
- 商品库存

## 9. Hash 和 String 怎么选

假设要存一个用户信息，有两种写法。

用 `String`：

```text
user:1001 -> "{\"name\":\"Alice\",\"age\":28,\"email\":\"alice@example.com\"}"
```

用 `Hash`：

```text
user:1001 -> {
  name: "Alice",
  age: "28",
  email: "alice@example.com"
}
```

一般来说：

- 如果你只是整块缓存一个对象，不关心单个字段，`String` 很方便
- 如果你经常只改某几个字段，`Hash` 更合适
- 如果对象字段比较多，`Hash` 通常比把整段 JSON 反复读写更自然

## 10. Hash 的底层原理

从使用角度看，`Hash` 就像 Redis 里的一个小字典。

```text
user:1001 -> {
  name: "Alice",
  age: "28",
  email: "alice@example.com"
}
```

它和 `String` 的区别在于：

- `String` 是 `key -> 一个值`
- `Hash` 是 `key -> 多个 field/value`

所以 `Hash` 特别适合表示对象。

### 10.1 为什么 Hash 适合对象存储

例如用户对象里只有 `age` 变了：

- 用 `String` 存 JSON，通常需要先取出整段 JSON，再修改，再整体写回
- 用 `Hash` 时，可以直接执行 `HSET user:1001 age 29`

也就是说，`Hash` 支持对对象的局部字段直接更新。

### 10.2 Hash 底层也是键值映射

Redis 整体是一个大 KV 系统，而 `Hash` 可以理解成 value 里面又维护了一层更小的映射关系：

```text
Redis:
  key -> value

Hash value:
  field -> value
```

所以访问 `HGET user:1001 name` 时，本质上做了两步：

1. 先通过 Redis 的全局 key 找到 `user:1001`
2. 再在这个 hash 里找到 field `name`

### 10.3 小对象很适合 Hash

当对象字段不多，而且每个字段值都不大时，`Hash` 会非常合适。

典型场景包括：

- 用户资料
- 商品属性
- 订单摘要
- 配置项集合

但如果：

- 对象特别深
- 需要嵌套很多层结构
- 经常整体序列化传输

那有时用 JSON + `String` 反而更直接。

## 11. Hash 的工程实践建议

- key 命名尽量带业务前缀，例如 `user:1001`、`order:20240301:88`
- field 名保持稳定，避免同一类对象字段风格不一致
- 不要把一个 hash 做得过大，否则读取和维护成本都会上升
- 如果整个对象需要统一过期，过期时间是设置在 key 上，不是 field 上

例如：

```redis
EXPIRE user:1002 600
TTL user:1002
```

这里过期的是整个 `user:1002`，不是其中某一个字段。

## 12. 最后总结

可以把 Redis `Hash` 理解成“用一个 key 存一个对象”。

最适合它的场景是：

- 一个对象有多个字段
- 经常按字段读取
- 经常只更新其中几个字段

如果你只需要存一个简单值，用 `String` 就够了；如果你需要表达一个结构化对象，`Hash` 往往更自然。
