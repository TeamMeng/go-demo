# wrk 登录接口压测记录

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
