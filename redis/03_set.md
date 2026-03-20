# Redis Set

`Set` 是 Redis 里“无序且元素唯一”的集合类型，适合存储标签、去重结果、用户关系、权限集合这类数据。

如果说：

- `String` 适合存一个值
- `Hash` 适合存一个对象的多个字段

那么 `Set` 更适合存“一组不重复的值”。

例如，一个用户关注的话题可以这样表示：

```text
user:1001:tags -> {"redis", "golang", "docker"}
```

## 1. Set 是什么

可以把 Redis `Set` 理解成：

```text
key -> {member1, member2, member3}
```

也就是说：

- 最外层还是一个 Redis key
- value 不是单个字符串，也不是 field-value 结构
- 而是一组“不重复”的 member

例如：

```text
user:1001:likes -> {"movie", "music", "travel"}
```

这里：

- `user:1001:likes` 是 key
- `"movie"`、`"music"`、`"travel"` 是集合里的 member
- member 不能重复

## 2. 添加元素

向集合里添加一个元素：

```redis
SADD user:1001:tags redis
```

返回：

```text
(integer) 1
```

再添加一个新元素：

```redis
SADD user:1001:tags golang
```

返回：

```text
(integer) 1
```

如果重复添加已有元素：

```redis
SADD user:1001:tags redis
```

返回：

```text
(integer) 0
```

这说明 `Set` 天然支持去重。

## 3. 查看所有元素

读取集合里的所有成员：

```redis
SMEMBERS user:1001:tags
```

返回结果类似：

```text
1) "redis"
2) "golang"
```

需要注意：

- `Set` 是无序的
- 返回顺序不保证固定

所以不要依赖 `SMEMBERS` 的返回顺序做业务逻辑。

## 4. 判断元素是否存在

```redis
SISMEMBER user:1001:tags redis
SISMEMBER user:1001:tags mysql
```

返回：

```text
(integer) 1
(integer) 0
```

这很适合做：

- 用户是否拥有某个标签
- 某个权限是否已经授予
- 某个用户是否已经点赞

## 5. 删除元素

从集合中删除一个成员：

```redis
SREM user:1001:tags golang
SMEMBERS user:1001:tags
```

返回结果类似：

```text
(integer) 1
1) "redis"
```

如果删除一个不存在的成员：

```redis
SREM user:1001:tags mysql
```

返回：

```text
(integer) 0
```

## 6. 获取集合大小

```redis
SCARD user:1001:tags
```

返回：

```text
(integer) 1
```

`SCARD` 常用于统计：

- 某用户拥有多少个标签
- 某篇文章被多少人点赞
- 某活动有多少去重后的参与用户

## 7. 随机获取或弹出元素

随机获取一个元素但不删除：

```redis
SRANDMEMBER user:1001:tags
```

返回结果可能是：

```text
"redis"
```

随机弹出一个元素并删除：

```redis
SPOP user:1001:tags
```

返回结果可能是：

```text
"redis"
```

这类命令适合做：

- 随机抽样
- 从候选池中随机取一个任务

## 8. 集合运算

`Set` 一个很强的点是支持交集、并集、差集。

假设有两个集合：

```redis
SADD user:1001:skills redis golang docker
SADD user:1002:skills redis mysql docker
```

### 8.1 交集

找出两个人共同拥有的技能：

```redis
SINTER user:1001:skills user:1002:skills
```

返回结果类似：

```text
1) "redis"
2) "docker"
```

### 8.2 并集

合并两个集合里的全部技能：

```redis
SUNION user:1001:skills user:1002:skills
```

返回结果类似：

```text
1) "redis"
2) "golang"
3) "docker"
4) "mysql"
```

### 8.3 差集

找出 `user:1001` 有、但 `user:1002` 没有的技能：

```redis
SDIFF user:1001:skills user:1002:skills
```

返回结果类似：

```text
1) "golang"
```

这类集合运算非常适合：

- 共同好友
- 共同兴趣
- 权限差异对比
- 标签推荐

## 9. Set 的典型使用场景

### 9.1 标签系统

```text
article:1001:tags -> {"redis", "database", "backend"}
```

适合按标签做去重存储。

### 9.2 点赞用户集合

```text
post:2001:likes -> {"user:1", "user:8", "user:15"}
```

优点是：

- 一个用户不会重复点赞两次
- 可以快速判断某用户是否点过赞
- 可以快速统计点赞人数

### 9.3 权限集合

```text
user:1001:permissions -> {"read", "write", "export"}
```

适合做权限判断和权限比对。

### 9.4 去重

比如记录今天访问过活动页的用户：

```text
activity:20260320:visitors -> {"user:1", "user:2", "user:3"}
```

同一个用户重复访问，也只会保留一份。

## 10. Set 和 List 怎么选

如果你关心的是“元素是否唯一”，通常选 `Set`。

如果你关心的是“元素顺序”或“允许重复”，通常选 `List`。

例如：

- 用户标签、权限、点赞用户集合，适合 `Set`
- 消息队列、时间顺序评论列表，适合 `List`

## 11. Set 的底层原理

从使用角度看，`Set` 就像一个自动去重的集合：

```text
user:1001:tags -> {"redis", "golang", "docker"}
```

它最核心的两个特点是：

- 元素唯一
- 判断某个元素是否存在通常很快

### 11.1 为什么 Set 适合做去重

因为 Redis 在 `Set` 里不会保存重复 member。

比如：

```redis
SADD user:1001:tags redis
SADD user:1001:tags redis
SADD user:1001:tags redis
```

虽然执行了三次，但集合里最终还是只有一个 `"redis"`。

所以很多“去重统计”“已处理用户集合”“已领取用户集合”都适合用 `Set`。

### 11.2 为什么判断成员是否存在很快

`Set` 的设计目标之一，就是高效判断一个 member 在不在集合中。

像下面这种命令：

```redis
SISMEMBER user:1001:tags redis
```

Redis 不需要像遍历数组那样一个个比较全部元素，因此成员判断通常非常高效。

这也是 `Set` 很适合做权限判断、点赞判断、是否处理过判断的原因。

## 12. 小结

`Set` 最适合这几类问题：

- 存一组不重复的数据
- 判断某个值是否存在
- 对数据做去重
- 统计集合大小
- 做交集、并集、差集运算

如果你的业务里出现“不能重复”“只关心在不在”“需要集合运算”这些关键词，优先考虑 Redis `Set`。
