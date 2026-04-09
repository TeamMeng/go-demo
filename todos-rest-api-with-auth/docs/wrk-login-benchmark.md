# wrk 登录接口压测记录

## 脚本说明

`scripts/wrk/login.lua` 当前支持两类写法：

1. 固定单用户压测

```lua
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body = '{"email":"TeamAlice@163.com","password":"123456"}'
```

这种方式可以使用，但只适合固定请求体的简单场景，例如一直压同一个有效账号。

它的限制也很明显：

- 只能发送一个固定账号
- 不能混入无效用户
- 不能控制有效和无效请求比例
- 不能按请求动态切换 body

2. 动态用户压测

当前仓库里的 `scripts/wrk/login.lua` 使用 `request()` 动态构造请求体，适合下面这些场景：

- 单个有效用户
- 多个有效用户轮询
- 有效用户和无效用户混合
- 按比例控制有效请求和无效请求

如果你只是想快速压一个账号，固定 `wrk.body` 的写法已经够用。
如果你要模拟真实登录流量，尤其是需要混入错误密码或不存在用户，就必须使用动态 `request()` 的写法。

## 单有效用户命令

```bash
wrk -t4 -c50 -d30s -s scripts/wrk/login.lua http://localhost:80 -- TeamAlice@163.com 123456
```

## 单有效用户结果

```text
Running 30s test @ http://localhost:80
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   559.91ms  269.59ms   1.78s    71.87%
    Req/Sec    22.13     12.05    60.00     54.93%
  2579 requests in 30.05s, 0.96MB read
Requests/sec:     85.84
Transfer/sec:     32.78KB
```

## 有效与无效用户混合命令

```bash
wrk -t4 -c50 -d30s -s scripts/wrk/login.lua http://localhost:80 -- @scripts/wrk/valid-users.txt @scripts/wrk/invalid-users.txt 0.8
```

其中：

- `valid-users.txt` 存放有效账号，格式为 `email,password`
- `invalid-users.txt` 存放无效账号或错误密码，格式同样为 `email,password`
- `0.8` 表示 80% 请求使用有效用户，20% 请求使用无效用户

示例文件位于：

- `scripts/wrk/valid-users.txt`
- `scripts/wrk/invalid-users.txt`

## 有效与无效用户混合结果

```text
Running 30s test @ http://localhost:80
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   492.33ms  299.89ms   1.58s    69.06%
    Req/Sec    25.64     15.62   100.00     61.64%
  2958 requests in 30.04s, 0.98MB read
  Non-2xx or 3xx responses: 590
Requests/sec:     98.46
Transfer/sec:     33.24KB
```

## 分析

- 单有效用户压测结果为 `85.84 req/s`，平均延迟 `559.91ms`
- 有效与无效用户混合压测结果为 `98.46 req/s`，平均延迟 `492.33ms`
- 混合压测中出现 `590` 个非 `2xx/3xx` 响应，占比约 `19.9%`，与 `0.8` 有效用户比例对应的 `20%` 无效请求基本一致
- 混合压测的吞吐高于单有效用户压测，说明部分无效请求比成功登录请求更快返回
- 从两组结果看，成功登录请求的处理成本明显高于失败登录请求，登录接口的主要开销更可能在密码校验和令牌签发，而不是简单的路由转发

---

# wrk 获取单条 Todo 压测记录

## 背景

`GET /todos/:id` 是受保护接口，请求需要携带有效 JWT。

这个项目的 JWT 不仅校验签名和过期时间，还额外校验当前请求的 `User-Agent` 是否与登录签发 Token 时一致。也就是说：

- 登录时会把 `ctx.Request.UserAgent()` 写入 JWT Claims
- 访问受保护接口时会再次读取当前请求头里的 `User-Agent`
- 如果两者不一致，即使 Token 本身未过期，也会返回 `401`

因此，使用 `wrk` 压测该接口时，不能只带 `Authorization`，还必须保证 `User-Agent` 与签发 Token 时使用的客户端一致。

## 首次压测现象

压测命令：

```bash
wrk -t4 -c50 -d30s -s scripts/wrk/get_todo.lua http://live.webook.com/todos/2
```

结果：

```text
Running 30s test @ http://live.webook.com/todos/2
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.66ms    3.75ms  71.94ms   91.00%
    Req/Sec     3.86k   578.79     5.47k    69.25%
  461262 requests in 30.05s, 72.58MB read
  Non-2xx or 3xx responses: 461262
Requests/sec:  15349.47
Transfer/sec:      2.42MB
```

问题特征：

- 所有请求都落入 `Non-2xx or 3xx responses`
- 吞吐看起来很高，但没有业务意义
- 这类结果通常说明服务端在快速返回错误，而不是正常执行业务逻辑

本次排查结论是：`wrk` 请求里的 `User-Agent` 与签发 Token 时使用的客户端不一致，导致所有请求都被鉴权中间件拒绝。

## 修正方式

`scripts/wrk/get_todo.lua` 中除了 `Authorization` 外，还需要显式设置与登录时一致的 `User-Agent`。

示例：

```lua
wrk.method = "GET"
wrk.headers["Accept"] = "application/json"
wrk.headers["Authorization"] = "Bearer " .. token
wrk.headers["User-Agent"] = "vscode-restclient"
```

如果 Token 是通过其他客户端签发的，这里的 `User-Agent` 也需要同步改成对应值。

## 修正后压测结果

压测命令：

```bash
wrk -t4 -c50 -d30s -s scripts/wrk/get_todo.lua http://live.webook.com/todos/2
```

结果：

```text
Running 30s test @ http://live.webook.com/todos/2
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     7.95ms   16.42ms 287.80ms   95.52%
    Req/Sec     2.34k   667.15     3.73k    77.92%
  279889 requests in 30.08s, 79.81MB read
Requests/sec:   9303.43
Transfer/sec:      2.65MB
```

## 分析

- 修正后 `Non-2xx or 3xx responses` 消失，说明请求已经通过鉴权并进入正常业务路径
- 吞吐从约 `15349 req/s` 下降到约 `9303 req/s`，这是合理现象
- 第一次压测本质上是在高并发下快速返回 `401`，成本很低，所以看起来更快
- 修正后请求需要经过完整鉴权、查询缓存或数据库、JSON 序列化与响应写回，真实业务成本更高，因此吞吐下降、平均延迟上升
- 当前这组结果才能代表 `GET /todos/:id` 的有效性能数据

## 后续建议

为了更准确评估读接口性能，建议后续拆成以下几组对比：

- 鉴权失败路径：验证 `401` 的极限吞吐
- 未命中缓存路径：验证数据库查询场景下的真实延迟
- 命中 Redis 缓存路径：验证读缓存后的吞吐提升

后续如果接入了 Redis 缓存，可继续在本文档追加“未使用缓存”和“命中缓存”两组结果，便于横向比较。

## Redis 接入后的新一轮结果

压测命令：

```bash
wrk -t4 -c50 -d30s -s scripts/wrk/get_todo.lua http://live.webook.com/todos/2
```

结果：

```text
Running 30s test @ http://live.webook.com/todos/2
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.05ms   10.21ms 176.30ms   94.68%
    Req/Sec     2.80k   667.38     4.42k    73.06%
  334729 requests in 30.07s, 95.45MB read
Requests/sec:  11133.11
Transfer/sec:      3.17MB
```

## 与上一轮成功压测对比

上一轮成功压测结果：

- `9303.43 req/s`
- 平均延迟 `7.95ms`
- 最大延迟 `287.80ms`
- 传输速率 `2.65 MB/s`

本轮结果：

- `11133.11 req/s`
- 平均延迟 `6.05ms`
- 最大延迟 `176.30ms`
- 传输速率 `3.17 MB/s`

变化情况：

- 吞吐提升约 `19.7%`
- 平均延迟下降约 `23.9%`
- 最大延迟明显下降
- 传输速率随成功请求数上升而提高

## 结论

- 从结果上看，当前版本相比上一轮成功压测已经有明显提升
- 如果这一轮对应的是接入 Redis 缓存后的实现，那么数据表现符合预期
- 这说明 `GET /todos/:id` 的热点读取路径已经比上一轮更轻，缓存大概率已经在发挥作用

不过，仅凭这一组 `wrk` 数据，还不能严格证明请求全部命中了 Redis。更稳妥的验证方式仍然是结合服务端日志、Redis 命中统计，或分别构造“冷缓存首轮请求”和“热缓存重复请求”两组压测做对照。
