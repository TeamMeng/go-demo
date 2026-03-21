# Go Experiments

`experiments/` 是这个仓库的知识实验区。这里没有业务逻辑，只有一组按主题拆开的 Go 最小示例：每个文件讲一个点，每个例子都尽量短，每个主题都可以单独运行。

如果把它当成学习材料来看，这个目录的价值不在“完整”，而在“可验证”：

- 代码短，读起来快
- 每个主题边界清楚
- 大部分例子可以直接通过 `go test` 观察行为
- 学习路径从语法一路推进到并发

## 运行方式

项目当前使用 Go `1.26.0`。

在仓库根目录执行：

```bash
go test -v ./experiments/...
go test -run TestSlice -v ./experiments
go test -run TestContext -v ./experiments
```

如果本地对默认 Go 缓存目录有限制，可以临时指定缓存位置：

```bash
GOCACHE=$(pwd)/.gocache go test -v ./experiments/...
```

## 阅读顺序

推荐按下面顺序看：

1. 语法基础
2. 数据结构与内存语义
3. 方法、接口与组合
4. 并发原语与协作模型

对应到文件，就是从 `1_variable_test.go` 一直看到 `26_block_test.go`。

## 知识地图

### 一、语法基础

| 文件 | 主题 | 你会看到什么 |
| --- | --- | --- |
| [1_variable_test.go](1_variable_test.go) | 变量 | `var`、`:=`、类型推断 |
| [2_constants_test.go](2_constants_test.go) | 常量 | `const`、显式类型与推断类型 |
| [3_interfaces_test.go](3_interfaces_test.go) | 类型预热 | `struct` 和 `interface` 的基本语法形态 |
| [4_func_test.go](4_func_test.go) | 函数 | 函数声明、返回值、简单断言 |
| [5_multiple_return_values_test.go](5_multiple_return_values_test.go) | 多返回值 | Go 原生支持多个返回值 |
| [6_if_else_test.go](6_if_else_test.go) | 条件分支 | `if / else` 的基本写法 |
| [7_switch_test.go](7_switch_test.go) | `switch` | 值匹配 `switch` 与无表达式 `switch` |
| [8_for_test.go](8_for_test.go) | 循环 | 传统循环、条件循环、`range`、无限循环 |

这一段的目标是先把 Go 的最小语法骨架搭起来。读完之后，至少应该对“怎么声明、怎么分支、怎么写函数、怎么循环”形成稳定直觉。

### 二、数据结构与内存语义

| 文件 | 主题 | 你会看到什么 |
| --- | --- | --- |
| [9_array_test.go](9_array_test.go) | 数组 | 定长数组、长度属于类型、`len` |
| [10_slice_test.go](10_slice_test.go) | 切片 | `nil` 切片、`make`、`len/cap`、`append`、切片表达式 |
| [11_map_test.go](11_map_test.go) | 映射 | `make(map)`、读写、删除、清空、零值读取 |
| [12_struct_test.go](12_struct_test.go) | 结构体 | 按字段名初始化、字段赋值 |
| [13_pointer_test.go](13_pointer_test.go) | 指针 | 传值与传指针的区别 |

这一段最关键的不是“会写语法”，而是开始理解 Go 的数据行为：

- 哪些值会拷贝
- 哪些操作会影响原对象
- 为什么切片、map、指针在使用感受上和普通值不同

### 三、方法、接口与组合

| 文件 | 主题 | 你会看到什么 |
| --- | --- | --- |
| [14_struct_method_test.go](14_struct_method_test.go) | 方法 | 给类型绑定方法，值接收者与指针接收者并存 |
| [15_receiver_test.go](15_receiver_test.go) | receiver 语义 | 值接收者修改副本，指针接收者修改原对象 |
| [16_interfaces_test.go](16_interfaces_test.go) | 接口 | 用方法集表达抽象能力 |
| [17_embed_test.go](17_embed_test.go) | 嵌入 | 通过组合复用字段和方法 |

这一段对应 Go 类型系统里最重要的几个思想：

- Go 没有传统类体系，但有方法
- 行为抽象靠接口，不靠继承
- 代码复用更偏向组合，而不是父类子类层级

其中有两个点最值得反复看：

- [15_receiver_test.go](15_receiver_test.go)：它直接解释了为什么很多方法会使用指针接收者
- [17_embed_test.go](17_embed_test.go)：它体现了 Go 风格里的“组合优于继承”

### 四、并发模型

| 文件 | 主题 | 你会看到什么 |
| --- | --- | --- |
| [18_goroutine_test.go](18_goroutine_test.go) | goroutine | 并发启动任务，命名函数与匿名函数都可用 |
| [19_channel_test.go](19_channel_test.go) | channel | 无缓冲与有缓冲 channel 的区别 |
| [20_select_test.go](20_select_test.go) | `select` | 同时等待多个 channel 分支 |
| [21_mutex_test.go](21_mutex_test.go) | `Mutex` | 用锁保护共享 map |
| [22_wait_group_test.go](22_wait_group_test.go) | `WaitGroup` | 等待一组 goroutine 执行结束 |
| [23_context_test.go](23_context_test.go) | `context` | 取消信号、超时模型的起点 |
| [24_chan_buffer_test.go](24_chan_buffer_test.go) | buffered channel | 缓冲区容量、阻塞时机、关闭后的读取行为 |
| [25_shared_array_lock_test.go](25_shared_array_lock_test.go) | 共享切片与加锁 | 观察并发 `append` 的风险，以及 `Mutex` 如何保护共享切片 |
| [26_block_test.go](26_block_test.go) | channel 阻塞 | 用最小示例观察缓冲区写满后发送方会卡住 |

这几组实验连起来，基本就是 Go 并发编程的入门框架：

- `goroutine` 负责启动任务
- `channel` 负责通信
- `select` 负责多路等待
- `Mutex` 负责保护共享状态
- `WaitGroup` 负责收尾同步
- `context` 负责取消与超时控制
- buffered channel 帮你理解“通信队列”和“背压”是怎么来的
- `Mutex` 也可以直接保护切片这类共享内存，不只是 map
- channel 在缓冲区写满后的阻塞行为，也可以先用最小示例直观看到

这里有几个值得注意的小点：

- [20_select_test.go](20_select_test.go) 使用了 `default`，所以是非阻塞轮询，更像概念演示
- [21_mutex_test.go](21_mutex_test.go) 现在用更准确的命名表达 `Mutex` 示例
- [22_wait_group_test.go](22_wait_group_test.go) 很适合继续延伸讨论循环变量捕获
- [23_context_test.go](23_context_test.go) 已经留下了 `WithTimeout` 的扩展空间
- [24_chan_buffer_test.go](24_chan_buffer_test.go) 把“有缓冲 channel”单独拆开，补了 `len/cap`、缓冲区满时的阻塞、`close` 后的 drain，以及 runtime 里环形缓冲区的直觉
- [25_shared_array_lock_test.go](25_shared_array_lock_test.go) 把“共享内存需要同步”讲得更直接：对共享切片做并发 `append` 会产生数据竞争，加锁后才有稳定语义
- [26_block_test.go](26_block_test.go) 保持成最小演示：先写满缓冲区，再观察下一次发送卡住以及后续日志不再出现

## 这一组代码背后的 Go 思维

从这组实验里，能看出 Go 的几条主线。

第一，语法追求直接。变量、循环、函数、返回值都尽量保持简单，不鼓励复杂表达。

第二，类型系统强调实用。结构体承担数据组织，方法承担行为，接口承担抽象，嵌入承担组合复用。

第三，并发是语言的一等公民。不是外部框架补上的能力，而是从 `go`、`chan`、`select` 到 `context` 都直接进入语言日常用法。

第四，Go 的并发原语虽然语法简单，但底层模型很具体。比如 buffered channel 并不是“自动异步”，而是一个固定容量的队列；队列满了就会产生背压，空了就会让接收方等待。

第五，Go 并发并不意味着“多个 goroutine 同时改同一份数据也没问题”。只要开始共享内存，就必须明确同步策略。`map` 的并发写很容易触发明显错误，切片的并发 `append` 则更隐蔽，经常要借助 `go test -race` 才能把问题暴露得更直接。

## 如果继续扩展，下一步建议补什么

这套实验已经把基础骨架搭出来了。往下扩展时，建议优先补这些主题：

1. `error` 处理
2. `defer` / `panic` / `recover`
3. table-driven tests
4. benchmark
5. 泛型
6. channel close
7. `for range ch`
8. channel 单向类型 `chan<-` / `<-chan`
9. `context.WithTimeout`
10. `go test -race`
11. `RWMutex`
12. 原子操作 `sync/atomic`

## 总结

`experiments/` 最适合被当成一个 Go 入门实验室来维护。好的方向不是把单个例子写复杂，而是继续保持两个原则：

- 每个文件只解释一个明确主题
- README 负责把这些主题组织成一条可阅读、可运行、可扩展的学习路径

如果沿着当前目录继续补，最自然的一条线就是把“并发通信”和“共享内存同步”这两类问题再拆细一点：一边继续补 channel 关闭、单向 channel、超时控制；另一边继续补 race detector、`RWMutex`、原子操作。
