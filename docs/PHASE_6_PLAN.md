# Phase 6 - Domain Model & Data Layer Separation Plan

## 1. 目标
将业务逻辑与缓存机制解耦，使 ggcache 成为通用组件。

## 2. 目录结构

### 2.1 Domain Layer
```markdown
internal/domain/student/
├── student.go          # 业务实体定义
├── student_validator.go # 数据验证
└── student_service.go   # 业务逻辑
```

### 2.2 Data Layer
```markdown
internal/data/student/
├── dao.go             # 数据访问对象
├── dao_test.go        # 数据访问测试
└── mock_dao.go        # 测试用 mock
```

### 2.3 Cache Core
```markdown
internal/cache/core/
├── student_retriever.go # 数据源实现
└── student_cache.go     # 缓存策略
```

### 2.4 Cache Transport
```markdown
internal/cache/transport/grpc/
└── student_grpc.go    # gRPC 实现
```

### 2.5 Application Layer
```markdown
internal/app/
└── student_app.go     # 业务流程组合
```

### 2.6 Command Layer
```markdown
cmd/ggcache/
└── student_handler.go # 请求处理
```

## 3. 实现步骤

1. 创建 Domain Layer
   - 定义 Student 实体
   - 实现业务验证
   - 实现业务服务

2. 创建 Data Layer
   - 实现 DAO 接口
   - 编写测试用例
   - 添加 mock 支持

3. 更新 Cache Core
   - 实现 StudentRetriever
   - 定义缓存策略

4. 更新 Cache Transport
   - 实现 gRPC 交互
   - 添加序列化逻辑

5. 更新 Application
   - 组合业务服务
   - 配置缓存

6. 更新 Command
   - 调用 StudentApp

## 4. 验证步骤

1. 单元测试
   - Domain 层
   - Data 层
   - Cache 层

2. 集成测试
   - 业务流程
   - 缓存交互
   - 网络通信

3. 性能测试
   - 缓存命中率
   - 请求响应时间

## 5. 文档更新

1. 更新架构文档
2. 更新开发指南
3. 更新部署文档
4. 更新测试文档
