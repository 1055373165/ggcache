# GGCache 性能分析指南

## 可用的性能分析端点

- http://localhost:6060/debug/pprof/ - 查看所有可用的 profile
- http://localhost:6060/debug/pprof/heap - 内存分配情况
- http://localhost:6060/debug/pprof/goroutine - goroutine 情况
- http://localhost:6060/debug/pprof/profile - CPU profile
- http://localhost:6060/debug/pprof/mutex - 锁竞争情况

## 一、交互式命令行分析

### 1. 收集性能数据
```bash
# CPU profile（默认30秒）
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 锁分析
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### 2. 常用交互式命令

- `top`: 显示资源使用最多的函数
  - flat: 函数自身的执行时间/内存使用
  - flat%: 自身占用百分比
  - sum%: 累计百分比
  - cum: 包含所有调用的函数的时间/内存
  - cum%: 累计占用百分比

- `list [函数名]`: 查看函数源码及性能数据
- `traces`: 查看所有调用栈信息
- `peek [函数名]`: 查看指定函数的调用方

## 二、图形化界面分析

### 1. 启动 Web 界面
```bash
go tool pprof -http=:8080 [profile文件]
```

### 2. 可用视图
- Graph: 函数调用关系图
  - 节点大小：执行时间/内存使用
  - 箭头粗细：调用频率
  - 颜色深浅：资源占用程度

- Flame Graph: 火焰图
- Top: 最耗资源函数列表
- Source: 带性能数据的源码视图

### 3. 安装依赖
```bash
# 查看图形界面需要安装 graphviz
brew install graphviz
```

## 三、性能文件生成与分析

### 1. 生成性能文件
```bash
# CPU profile
go tool pprof -output=$(pwd)/deploy/pprof/cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof -output=$(pwd)/deploy/pprof/heap.prof http://localhost:6060/debug/pprof/heap

# Goroutine 分析
go tool pprof -output=$(pwd)/deploy/pprof/goroutine.prof http://localhost:6060/debug/pprof/goroutine
```

### 2. 分析已生成的文件
```bash
# 交互式分析
go tool pprof [profile文件]

# 图形化分析
go tool pprof -http=:8080 [profile文件]
```

## 性能优化建议

1. 优先关注 `top` 命令中的高占用函数
2. 使用火焰图直观定位性能瓶颈
3. 注意 `cum%` 较高的函数调用链
4. 分析锁竞争情况，优化并发性能

## 注意事项

1. 性能分析会影响程序运行，建议在测试环境使用
2. 可通过 URL 参数调整采样时间：`?seconds=60`
3. 内存分析会增加内存开销
4. 生产环境使用时需要限制访问权限

## 热点函数分析最佳实践

### 1. 发现热点函数

1. 使用 top 命令初步定位
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
(pprof) top10 -cum    # 按照累计时间排序
```

2. 使用火焰图可视化
```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
# 在浏览器中查看 Flame Graph
```

### 2. 分析函数调用链

1. 查看函数调用关系
```bash
(pprof) peek <hot-function>    # 查看谁调用了热点函数
(pprof) list <hot-function>    # 查看函数源码及性能数据
```

2. 生成调用图
```bash
(pprof) web    # 在浏览器中查看调用图
# 关注红色节点和粗线条
```

### 3. 常见优化策略

1. CPU 密集型优化
   - 优化算法复杂度
   - 减少不必要的内存分配
   - 使用 sync.Pool 复用对象
   - 考虑使用并发处理

2. 内存分配优化
   - 减少临时对象创建
   - 使用对象池
   - 预分配内存
   - 使用合适的数据结构

3. 并发优化
   - 减少锁的粒度
   - 使用原子操作代替互斥锁
   - 合理设置 goroutine 数量
   - 优化 channel 使用

4. I/O 优化
   - 使用缓冲 I/O
   - 批量处理
   - 异步处理
   - 使用连接池

### 4. 优化步骤示例

1. 发现 CPU 使用率高的函数
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
(pprof) top
```

2. 分析函数代码
```bash
(pprof) list <function>
# 查看具体耗时的代码行
```

3. 确定优化方向
- 如果是算法问题：优化算法复杂度
- 如果是内存问题：减少分配，使用对象池
- 如果是并发问题：优化锁使用，调整并发模型
- 如果是 I/O 问题：使用缓冲，批量处理

4. 实施优化
- 修改代码
- 编写基准测试
- 对比优化前后的性能

5. 验证优化效果
```bash
# 收集优化后的性能数据
go tool pprof http://localhost:6060/debug/pprof/profile

# 对比优化前后的 profile 文件
go tool pprof -http=:8080 -diff_base=before.prof after.prof
```

### 5. 注意事项

1. 优化建议
   - 先优化最耗时的热点函数
   - 关注 CPU 和内存使用异常的代码
   - 注意并发安全和死锁问题
   - 保持代码可读性和可维护性

2. 避免过早优化
   - 先确保代码正确性
   - 有数据支撑再优化
   - 保持简单性
   - 权衡优化成本和收益

3. 持续监控
   - 定期收集性能数据
   - 监控关键指标变化
   - 及时发现性能退化
   - 建立性能基准