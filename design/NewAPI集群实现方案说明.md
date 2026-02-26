# NewAPI100集群实现方案说明

## 1. 总体架构概览

### 1.1 集群架构设计

NewAPI100采用多层次的集群架构设计，支持水平扩展和高可用部署：

```
┌─────────────────────────────────────────────────────────────┐
│                        负载均衡层                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   Nginx/HAProxy │  │   API Gateway   │  │   CDN节点    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────┐
│                       应用集群层                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   Master节点    │  │   Slave节点     │  │  Slave节点   │  │
│  │ - 数据库迁移     │  │ - 只读服务      │  │ - 只读服务   │  │
│  │ - 定时任务       │  │ - API处理       │  │ - API处理    │  │
│  │ - 配置管理       │  │ - 缓存同步      │  │ - 缓存同步   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────┐
│                      数据存储层                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   主数据库      │  │   Redis集群     │  │   日志数据库 │  │
│  │ - MySQL/Postgre │  │ - 分布式缓存    │  │ - 审计日志    │  │
│  │ - 业务数据      │  │ - 分布式锁      │  │ - 分析数据    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 核心集群概念

#### 1.2.1 Master-Slave模式
- **Master节点**: 负责数据写入、定时任务、数据库迁移
- **Slave节点**: 负责API请求处理、缓存同步、只读操作
- **配置方式**: 通过环境变量 `NODE_TYPE` 控制

#### 1.2.2 分布式缓存 (Redis)
- **作用**: 跨节点数据共享、分布式锁、限流控制
- **特性**: 支持集群模式、主从复制、持久化

#### 1.2.3 数据库分离
- **主数据库**: 业务数据读写
- **日志数据库**: 审计日志、分析数据
- **支持类型**: MySQL、PostgreSQL、SQLite

#### 1.2.4 分布式限流
- **实现方式**: Redis Lua脚本
- **算法**: 令牌桶算法
- **特性**: 跨节点精确控制

## 2. Master-Slave集群架构详解

### 2.1 节点角色控制机制

#### 2.1.1 角色配置时序图

```mermaid
sequenceDiagram
    participant Env as 环境变量
    participant Init as 初始化过程
    participant Node as 节点实例
    participant DB as 数据库

    Env->>Init: 读取NODE_TYPE
    Init->>Init: 解析节点类型

    alt NODE_TYPE == "slave"
        Init->>Node: 设置为Slave节点
        Node->>Node: 禁用数据库迁移
        Node->>Node: 禁用定时任务
        Node->>DB: 使用只读连接
    else 其他情况 (默认Master)
        Init->>Node: 设置为Master节点
        Node->>Node: 启用数据库迁移
        Node->>Node: 启用定时任务
        Node->>DB: 使用读写连接
    end

    Node->>Node: 初始化完成
    Node-->>Node: 节点准备就绪
```

#### 2.1.2 节点角色判断代码

```go
// common/init.go
IsMasterNode = os.Getenv("NODE_TYPE") != "slave"

// model/main.go - 数据库初始化时的角色控制
if !common.IsMasterNode {
    return nil  // Slave节点跳过数据库迁移
}
```

### 2.2 Master节点职责

#### 2.2.1 数据库迁移与维护

**Master节点独占操作**:
```go
// model/main.go
func InitDB() (err error) {
    // ... 数据库连接初始化 ...

    if !common.IsMasterNode {
        return nil  // Slave节点不执行迁移
    }

    // Master节点执行数据库迁移
    common.SysLog("database migration started")
    err = migrateDB()
    return err
}
```

#### 2.2.2 定时任务执行

**定时任务控制**:
```go
// main.go
if common.IsMasterNode && constant.UpdateTask {
    // Master节点执行定时任务
    service.InitTaskSystem()
}
```

### 2.3 Slave节点职责

#### 2.3.1 只读服务处理

**Slave节点API处理**:
```go
// 所有API请求都可以由Slave节点处理
// Master-Slave模式下，业务逻辑相同，只是禁用写操作
func handleAPIRequest(c *gin.Context) {
    // 读操作：所有节点都可以处理
    // 写操作：会通过数据库约束或业务逻辑控制
}
```

#### 2.3.2 缓存同步机制

**跨节点缓存同步**:
```mermaid
sequenceDiagram
    participant Master as Master节点
    participant Slave1 as Slave节点1
    participant Slave2 as Slave节点2
    participant Redis as Redis集群

    Master->>Redis: 更新缓存数据
    Redis->>Redis: 广播缓存变更

    Redis->>Slave1: 推送缓存更新
    Redis->>Slave2: 推送缓存更新

    Slave1->>Slave1: 同步本地缓存
    Slave2->>Slave2: 同步本地缓存

    Slave1-->>Slave1: 缓存同步完成
    Slave2-->>Slave2: 缓存同步完成
```

## 3. 分布式缓存与锁机制

### 3.1 Redis集群架构

#### 3.1.1 Redis连接配置

```go
// common/redis.go
func InitRedisClient() (err error) {
    if os.Getenv("REDIS_CONN_STRING") == "" {
        RedisEnabled = false
        SysLog("REDIS_CONN_STRING not set, Redis is not enabled")
        return nil
    }

    opt, err := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))
    if err != nil {
        FatalLog("failed to parse Redis connection string: " + err.Error())
    }

    opt.PoolSize = GetEnvOrDefault("REDIS_POOL_SIZE", 10)
    RDB = redis.NewClient(opt)

    // 连接测试
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = RDB.Ping(ctx).Result()
    if err != nil {
        FatalLog("Redis ping test failed: " + err.Error())
    }
}
```

#### 3.1.2 Redis操作封装

**原子操作支持**:
```go
// Redis原子增减操作
func RedisIncr(key string, delta int64) error {
    // 检查TTL并保持TTL一致性
    ttlCmd := RDB.TTL(context.Background(), key)
    ttl, err := ttlCmd.Result()
    if err != nil && !errors.Is(err, redis.Nil) {
        return fmt.Errorf("failed to get TTL: %w", err)
    }

    if ttl > 0 {
        // 使用事务保持TTL
        ctx := context.Background()
        txn := RDB.TxPipeline()

        decrCmd := txn.IncrBy(ctx, key, delta)
        txn.Expire(ctx, key, ttl)

        _, err = txn.Exec(ctx)
        return err
    }
    return nil
}
```

### 3.2 分布式锁实现

#### 3.2.1 Redis分布式锁时序图

```mermaid
sequenceDiagram
    participant Node as 请求节点
    participant Redis as Redis集群
    participant LockMgr as 锁管理器

    Node->>LockMgr: 请求分布式锁
    LockMgr->>Redis: SET key value NX PX ttl

    alt 锁获取成功
        Redis-->>LockMgr: OK
        LockMgr-->>Node: 锁获取成功
        Node->>Node: 执行临界区代码
        Node->>LockMgr: 释放锁
        LockMgr->>Redis: DEL key
    else 锁获取失败
        Redis-->>LockMgr: nil
        LockMgr-->>Node: 锁获取失败
        Node->>LockMgr: 等待重试
    end
```

#### 3.2.2 配额操作分布式锁

**防止并发扣费**:
```go
// service/quota.go - 原子配额操作
func atomicQuotaOperation(userId int, operation func() error) error {
    lockKey := fmt.Sprintf("quota_lock:%d", userId)

    // 获取分布式锁（带超时和重试）
    lock := acquireDistributedLock(lockKey, 30*time.Second, 3)
    if lock == nil {
        return errors.New("获取配额操作锁失败，请重试")
    }
    defer releaseDistributedLock(lock)

    // 执行原子操作
    return operation()
}
```

## 4. 分布式限流系统

### 4.1 令牌桶算法实现

#### 4.1.1 分布式限流架构

```mermaid
graph TD
    A[API请求] --> B{本地缓存检查}
    B -->|命中| C[快速允许]
    B -->|未命中| D[Redis分布式检查]

    D --> E{Lua脚本执行}
    E -->|允许| F[更新令牌桶]
    E -->|拒绝| G[返回429错误]

    F --> H[设置过期时间]
    H --> I[返回成功]

    C --> J[记录限流事件]
    G --> J
    I --> J
```

#### 4.1.2 Redis Lua脚本实现

**原子性令牌桶算法**:
```lua
-- common/limiter/lua/rate_limit.lua
local key = KEYS[1]
local requested = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])

-- 获取当前时间戳
local now = redis.call('TIME')
local nowInSeconds = tonumber(now[1])

-- 获取桶状态
local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local tokens = tonumber(bucket[1])
local last_time = tonumber(bucket[2])

-- 初始化或补充令牌
if not tokens or not last_time then
    tokens = capacity
    last_time = nowInSeconds
else
    local elapsed = nowInSeconds - last_time
    local add_tokens = elapsed * rate
    tokens = math.min(capacity, tokens + add_tokens)
    last_time = nowInSeconds
end

-- 检查是否允许请求
local allowed = false
if tokens >= requested then
    tokens = tokens - requested
    allowed = true
end

-- 更新桶状态
redis.call('HMSET', key, 'tokens', tokens, 'last_time', last_time)

return allowed and 1 or 0
```

#### 4.1.3 限流器使用示例

**用户级限流**:
```go
// 创建限流器实例
limiter, _ := limiter.New(context.Background(), common.RDB)

// 用户请求限流：每分钟最多10个请求
allowed, err := limiter.Allow(ctx, fmt.Sprintf("user:%d", userId),
    limiter.WithCapacity(10),    // 容量：10
    limiter.WithRate(1.0/60),    // 速率：每秒1/60个
    limiter.WithRequested(1),    // 请求1个
)

if !allowed {
    return errors.New("请求过于频繁，请稍后再试")
}
```

### 4.2 限流规则配置

#### 4.2.1 多维度限流配置

```go
// 限流配置结构
type RateLimitConfig struct {
    // 用户级限流
    UserLimits struct {
        RequestsPerMinute int `json:"requests_per_minute"`
        RequestsPerHour   int `json:"requests_per_hour"`
        RequestsPerDay    int `json:"requests_per_day"`
    } `json:"user_limits"`

    // 模型级限流
    ModelLimits map[string]struct {
        RequestsPerMinute int `json:"requests_per_minute"`
        ConcurrentLimit   int `json:"concurrent_limit"`
    } `json:"model_limits"`

    // 渠道级限流
    ChannelLimits map[int]struct {
        RequestsPerMinute int `json:"requests_per_minute"`
        ErrorRateLimit    float64 `json:"error_rate_limit"`
    } `json:"channel_limits"`
}
```

#### 4.2.2 动态限流调整

**自适应限流算法**:
```go
type AdaptiveRateLimiter struct {
    baseRate       int
    currentRate    int
    adjustmentStep int
    monitorWindow  time.Duration
    cpuThreshold   float64
    memThreshold   float64
}

func (arl *AdaptiveRateLimiter) Allow() bool {
    arl.adjustRate()  // 动态调整限流速率

    // 使用当前限流速率进行检查
    return arl.tokenBucket.Allow()
}

func (arl *AdaptiveRateLimiter) adjustRate() {
    if time.Since(arl.lastAdjustment) < arl.monitorWindow {
        return
    }

    // 获取系统负载
    cpuUsage := getCPUUsage()
    memUsage := getMemoryUsage()

    // 根据负载调整限流
    if cpuUsage > arl.cpuThreshold || memUsage > arl.memThreshold {
        // 高负载：降低限流阈值
        arl.currentRate = max(arl.baseRate/2, arl.currentRate - arl.adjustmentStep)
        common.SysLog(fmt.Sprintf("高负载检测，降低限流速率至: %d", arl.currentRate))
    } else if cpuUsage < arl.cpuThreshold*0.8 {
        // 低负载：逐步恢复限流
        arl.currentRate = min(arl.baseRate, arl.currentRate + arl.adjustmentStep/2)
        common.SysLog(fmt.Sprintf("负载正常，恢复限流速率至: %d", arl.currentRate))
    }

    arl.lastAdjustment = time.Now()
}
```

## 5. 数据库分离与读写分离

### 5.1 数据库架构设计

#### 5.1.1 双数据库架构

```mermaid
graph TD
    subgraph "主数据库 (Master DB)"
        A1[业务数据表]
        A2[用户数据]
        A3[配置数据]
        A4[交易数据]
    end

    subgraph "日志数据库 (Log DB)"
        B1[使用日志]
        B2[审计日志]
        B3[错误日志]
        B4[分析数据]
    end

    subgraph "应用集群"
        C1[Master节点]
        C2[Slave节点]
    end

    C1 --> A1
    C1 --> A2
    C1 --> A3
    C1 --> A4

    C1 --> B1
    C1 --> B2
    C2 --> B1
    C2 --> B2

    A1 -.-> C2
    A2 -.-> C2
```

#### 5.1.2 数据库初始化流程

**双数据库连接管理**:
```go
// 主数据库初始化
func InitDB() (err error) {
    db, err := chooseDB("SQL_DSN", false)
    if err != nil {
        return err
    }

    if common.UsingMySQL {
        // MySQL中文支持检查
        if err := checkMySQLChineseSupport(db); err != nil {
            return err
        }
    }

    // 连接池配置
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 100))
    sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 1000))
    sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 60)))

    DB = db

    // Master节点执行数据库迁移
    if common.IsMasterNode {
        err = migrateDB()
    }
    return err
}

// 日志数据库初始化
func InitLogDB() (err error) {
    if os.Getenv("LOG_SQL_DSN") == "" {
        LOG_DB = DB  // 使用主数据库
        return
    }

    db, err := chooseDB("LOG_SQL_DSN", true)
    if err != nil {
        return err
    }

    LOG_DB = db
    return nil
}
```

### 5.2 数据库支持类型

#### 5.2.1 多数据库适配

**数据库选择逻辑**:
```go
func chooseDB(envName string, isLog bool) (*gorm.DB, error) {
    dsn := os.Getenv(envName)
    if dsn == "" {
        // 默认使用SQLite
        common.SysLog("SQL_DSN not set, using SQLite as database")
        common.UsingSQLite = true
        return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
            PrepareStmt: true,
        })
    }

    // PostgreSQL支持
    if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
        common.SysLog("using PostgreSQL as database")
        if !isLog {
            common.UsingPostgreSQL = true
        }
        return gorm.Open(postgres.New(postgres.Config{
            DSN:                  dsn,
            PreferSimpleProtocol: true,
        }), &gorm.Config{
            PrepareStmt: true,
        })
    }

    // MySQL支持（默认）
    common.SysLog("using MySQL as database")
    if strings.Contains(dsn, "parseTime") {
        // 确保MySQL连接支持时间解析
        if !strings.Contains(dsn, "parseTime=true") {
            dsn += "&parseTime=true"
        }
    }

    if !isLog {
        common.UsingMySQL = true
    }
    return gorm.Open(mysql.Open(dsn), &gorm.Config{
        PrepareStmt: true,
    })
}
```

## 6. 批量更新与异步处理

### 6.1 批量更新队列架构

#### 6.1.1 批量更新机制时序图

```mermaid
sequenceDiagram
    participant Request as API请求
    participant Queue as 更新队列
    participant Worker as 批量处理协程
    participant DB as 数据库
    participant Cache as Redis缓存

    Request->>Queue: 提交更新任务
    Queue->>Queue: 累积批量任务

    alt 达到批量大小
        Queue->>Worker: 触发批量处理
    else 定时器触发
        Queue->>Worker: 定时批量处理
    end

    Worker->>DB: 批量数据库更新
    DB-->>Worker: 更新确认

    Worker->>Cache: 批量缓存更新
    Cache-->>Worker: 缓存确认

    Worker->>Queue: 处理完成
    Worker-->>Request: 更新成功
```

#### 6.1.2 批量更新实现

**队列处理机制**:
```go
// 批量更新记录结构
type BatchUpdateRecord struct {
    Type  int // 更新类型 (配额、使用量等)
    Id    int // 用户ID
    Value int // 更新值
}

var batchUpdateQueue = make(chan BatchUpdateRecord, 1000)

// 批量处理协程
func batchUpdateWorker() {
    const batchSize = 100
    const flushInterval = time.Second * 5

    ticker := time.NewTicker(flushInterval)
    defer ticker.Stop()

    batch := make([]BatchUpdateRecord, 0, batchSize)

    for {
        select {
        case record := <-batchUpdateQueue:
            batch = append(batch, record)

            // 达到批量大小时立即处理
            if len(batch) >= batchSize {
                processBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            // 定时处理剩余记录
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0]
            }
        }
    }
}

func processBatch(batch []BatchUpdateRecord) {
    // 按用户ID分组
    userUpdates := make(map[int]int)
    for _, record := range batch {
        userUpdates[record.Id] += record.Value
    }

    // 批量更新数据库
    for userId, totalValue := range userUpdates {
        updateUserQuotaBatch(userId, totalValue)
        // 异步更新缓存
        gopool.Go(func() {
            cacheIncrUserQuota(userId, int64(totalValue))
        })
    }
}
```

### 6.2 异步任务处理系统

#### 6.2.1 任务队列架构

```mermaid
graph TD
    A[任务提交] --> B{任务类型}
    B -->|即时任务| C[直接执行]
    B -->|异步任务| D[推送到队列]

    D --> E[任务调度器]
    E --> F{队列长度}
    F -->|正常| G[等待处理]
    F -->|拥堵| H[流量控制]

    G --> I[工作协程池]
    I --> J[任务执行器]
    J --> K[结果处理]

    K --> L{执行结果}
    L -->|成功| M[完成回调]
    L -->|失败| N[重试机制]

    N --> O{重试次数}
    O -->|未超限| P[重新入队]
    O -->|超限| Q[失败处理]
```

#### 6.2.2 协程池管理

**高性能协程池**:
```go
// 使用gopkg协程池
import "github.com/bytedance/gopkg/util/gopool"

// 异步任务执行
func asyncTaskExecution(task Task) {
    gopool.Go(func() {
        defer func() {
            if r := recover(); r != nil {
                // 异常处理和日志记录
                common.SysError(fmt.Sprintf("Task execution panic: %v", r))
            }
        }()

        // 执行具体任务
        result := executeTask(task)

        // 结果回调处理
        if task.Callback != nil {
            task.Callback(result)
        }
    })
}
```

## 7. 集群监控与健康检查

### 7.1 集群健康监控

#### 7.1.1 健康检查指标

```yaml
# 集群健康监控指标
cluster_health_metrics:
  # 节点状态
  - name: "node_status"
    type: "gauge"
    description: "节点运行状态"
    labels: ["node_id", "node_type", "node_role"]

  # 数据库连接
  - name: "db_connection_status"
    type: "gauge"
    description: "数据库连接状态"
    labels: ["db_type", "db_role"]

  # Redis连接
  - name: "redis_connection_status"
    type: "gauge"
    description: "Redis连接状态"

  # 队列积压
  - name: "queue_backlog"
    type: "gauge"
    description: "队列积压数量"
    labels: ["queue_name"]

  # 响应时间
  - name: "node_response_time"
    type: "histogram"
    description: "节点响应时间分布"
    labels: ["node_id", "endpoint"]
```

#### 7.1.2 健康检查实现

**多维度健康检查**:
```go
// 集群健康检查器
type ClusterHealthChecker struct {
    nodes       []NodeInfo
    checkInterval time.Duration
    alertManager *AlertManager
}

func (chc *ClusterHealthChecker) StartHealthCheck() {
    ticker := time.NewTicker(chc.checkInterval)
    for range ticker.C {
        chc.performHealthChecks()
    }
}

func (chc *ClusterHealthChecker) performHealthChecks() {
    for _, node := range chc.nodes {
        // 数据库连接检查
        if !chc.checkDatabaseConnection(node) {
            chc.alertManager.SendAlert(Alert{
                Type:    AlertTypeDatabase,
                Level:   AlertLevelCritical,
                Message: fmt.Sprintf("Node %s database connection failed", node.ID),
            })
        }

        // Redis连接检查
        if !chc.checkRedisConnection(node) {
            chc.alertManager.SendAlert(Alert{
                Type:    AlertTypeRedis,
                Level:   AlertLevelWarning,
                Message: fmt.Sprintf("Node %s Redis connection failed", node.ID),
            })
        }

        // 队列积压检查
        backlog := chc.checkQueueBacklog(node)
        if backlog > 1000 {
            chc.alertManager.SendAlert(Alert{
                Type:    AlertTypeQueue,
                Level:   AlertLevelWarning,
                Message: fmt.Sprintf("Node %s queue backlog: %d", node.ID, backlog),
            })
        }
    }
}
```

### 7.2 负载均衡监控

#### 7.2.1 负载分布监控

```mermaid
graph TD
    A[负载均衡器] --> B[请求分发]
    B --> C[节点选择]

    C --> D{负载检查}
    D -->|正常| E[转发请求]
    D -->|过载| F[重定向到其他节点]

    E --> G[请求处理]
    G --> H[响应返回]

    F --> I[负载重新评估]
    I --> J{其他节点可用?}
    J -->|是| K[选择备用节点]
    J -->|否| L[返回503错误]

    K --> G
    L --> M[负载均衡失败]
```

#### 7.2.2 负载均衡算法

**加权轮询负载均衡**:
```go
type LoadBalancer struct {
    nodes       []*Node
    currentIndex int64
    mutex       sync.Mutex
}

type Node struct {
    ID       string
    Address  string
    Weight   int    // 节点权重
    CurrentWeight int // 当前权重
}

func (lb *LoadBalancer) SelectNode() *Node {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()

    if len(lb.nodes) == 0 {
        return nil
    }

    totalWeight := 0
    var selectedNode *Node

    // 加权轮询算法
    for _, node := range lb.nodes {
        node.CurrentWeight += node.Weight
        totalWeight += node.Weight

        if selectedNode == nil || node.CurrentWeight > selectedNode.CurrentWeight {
            selectedNode = node
        }
    }

    if selectedNode != nil {
        selectedNode.CurrentWeight -= totalWeight
    }

    return selectedNode
}
```

## 8. 集群部署配置

### 8.1 Docker集群部署

#### 8.1.1 Docker Compose配置

**集群部署配置**:
```yaml
# docker-compose.cluster.yml
version: '3.8'

services:
  # Master节点
  newapi-master:
    image: quantumnous/new-api:latest
    environment:
      - NODE_TYPE=master
      - SQL_DSN=mysql://user:password@mysql:3306/newapi
      - LOG_SQL_DSN=mysql://user:password@mysql:3306/newapi_logs
      - REDIS_CONN_STRING=redis://redis:6379
    ports:
      - "3000:3000"
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/app/logs

  # Slave节点 (多个)
  newapi-slave-1:
    image: quantumnous/new-api:latest
    environment:
      - NODE_TYPE=slave
      - SQL_DSN=mysql://user:password@mysql:3306/newapi
      - LOG_SQL_DSN=mysql://user:password@mysql:3306/newapi_logs
      - REDIS_CONN_STRING=redis://redis:6379
    ports:
      - "3001:3000"
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/app/logs

  newapi-slave-2:
    image: quantumnous/new-api:latest
    environment:
      - NODE_TYPE=slave
      - SQL_DSN=mysql://user:password@mysql:3306/newapi
      - LOG_SQL_DSN=mysql://user:password@mysql:3306/newapi_logs
      - REDIS_CONN_STRING=redis://redis:6379
    ports:
      - "3002:3000"
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/app/logs

  # MySQL主从集群
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
      - MYSQL_DATABASE=newapi
    volumes:
      - mysql_data:/var/lib/mysql
      - ./mysql-init:/docker-entrypoint-initdb.d

  # Redis集群
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  # Nginx负载均衡
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - newapi-master
      - newapi-slave-1
      - newapi-slave-2

volumes:
  mysql_data:
  redis_data:
```

#### 8.1.2 Nginx负载均衡配置

**Nginx配置**:
```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream newapi_cluster {
        # Master节点权重更高
        server newapi-master:3000 weight=3;
        server newapi-slave-1:3000 weight=2;
        server newapi-slave-2:3000 weight=2;

        # 健康检查
        check interval=3000 rise=2 fall=3 timeout=1000;
    }

    server {
        listen 80;

        location / {
            proxy_pass http://newapi_cluster;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # 超时配置
            proxy_connect_timeout 10s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # 健康检查端点
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

### 8.2 Kubernetes集群部署

#### 8.2.1 Kubernetes部署配置

**Deployment配置**:
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: newapi
  labels:
    app: newapi
spec:
  replicas: 3
  selector:
    matchLabels:
      app: newapi
  template:
    metadata:
      labels:
        app: newapi
    spec:
      containers:
      - name: newapi
        image: quantumnous/new-api:latest
        ports:
        - containerPort: 3000
        env:
        - name: NODE_TYPE
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: SQL_DSN
          value: "mysql://user:password@mysql-service:3306/newapi"
        - name: REDIS_CONN_STRING
          value: "redis://redis-service:6379"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

#### 8.2.2 Service和Ingress配置

**Service配置**:
```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: newapi-service
spec:
  selector:
    app: newapi
  ports:
  - port: 80
    targetPort: 3000
    protocol: TCP
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: newapi-ingress
spec:
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: newapi-service
            port:
              number: 80
```

## 9. 集群性能优化

### 9.1 缓存优化策略

#### 9.1.1 多级缓存架构

```mermaid
graph TD
    A[用户请求] --> B{本地缓存}
    B -->|命中| C[直接返回]
    B -->|未命中| D{分布式缓存}

    D -->|命中| E[更新本地缓存]
    D -->|未命中| F[数据库查询]

    E --> C
    F --> G[更新缓存]
    G --> C

    C --> H[返回结果]
    G --> H
```

#### 9.1.2 缓存一致性策略

**缓存同步机制**:
```go
// 缓存预热和一致性保证
func ensureCacheConsistency() {
    // 1. 启动缓存同步协程
    go func() {
        ticker := time.NewTicker(time.Minute * 5)
        for range ticker.C {
            syncCacheFromDatabase()
        }
    }()

    // 2. 写操作时主动更新缓存
    func updateWithCacheSync(key string, value interface{}) error {
        // 更新数据库
        if err := updateDatabase(key, value); err != nil {
            return err
        }

        // 更新分布式缓存
        if err := updateRedisCache(key, value); err != nil {
            // 记录错误但不影响主流程
            common.SysLog(fmt.Sprintf("Cache update failed: %v", err))
        }

        // 广播缓存失效通知
        broadcastCacheInvalidation(key)

        return nil
    }
}
```

### 9.2 连接池优化

#### 9.2.1 数据库连接池配置

**连接池调优**:
```go
func optimizeDatabaseConnectionPool() {
    sqlDB, err := DB.DB()
    if err != nil {
        return
    }

    // 根据集群规模调整连接池
    nodeCount := getClusterNodeCount()
    baseConnections := 10

    // 动态计算连接池大小
    maxIdleConns := baseConnections * nodeCount
    maxOpenConns := maxIdleConns * 5

    sqlDB.SetMaxIdleConns(maxIdleConns)
    sqlDB.SetMaxOpenConns(maxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Hour)

    common.SysLog(fmt.Sprintf("Database connection pool optimized: MaxIdle=%d, MaxOpen=%d",
        maxIdleConns, maxOpenConns))
}
```

#### 9.2.2 Redis连接池配置

**Redis连接池优化**:
```go
func optimizeRedisConnectionPool() {
    opt, _ := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))

    // 集群环境下增加连接池大小
    clusterSize := getRedisClusterSize()
    opt.PoolSize = 10 * clusterSize     // 连接池大小
    opt.MinIdleConns = 2 * clusterSize  // 最小空闲连接
    opt.PoolTimeout = time.Second * 30  // 连接池超时

    // 启用管道模式提升性能
    opt.PipelineTimeout = time.Second * 5

    common.SysLog(fmt.Sprintf("Redis connection pool optimized for cluster: PoolSize=%d, MinIdle=%d",
        opt.PoolSize, opt.MinIdleConns))
}
```

## 10. 总结

NewAPI的集群实现方案是一个企业级的、高可用的分布式系统架构，具有以下核心特点：

### 架构优势

1. **Master-Slave模式**: 清晰的职责分离，支持水平扩展
2. **分布式缓存**: Redis集群提供高性能数据共享
3. **分布式锁**: 防止并发冲突，保证数据一致性
4. **分布式限流**: 跨节点精确流量控制
5. **数据库分离**: 读写分离提升性能和可用性

### 技术亮点

1. **原子操作**: Redis Lua脚本保证分布式操作的原子性
2. **异步处理**: 协程池和批量更新提升并发性能
3. **自适应调节**: 基于负载的动态限流和资源调整
4. **健康监控**: 多维度监控和自动故障恢复
5. **容器化部署**: 支持Docker和Kubernetes集群部署

### 运维便捷性

1. **配置统一**: 环境变量驱动的配置管理
2. **监控完善**: Prometheus指标和告警系统
3. **扩容灵活**: 支持动态节点增减
4. **故障自愈**: 自动故障检测和恢复机制

该集群架构不仅满足了NewAPI的高并发、大流量场景需求，还为未来的业务扩展和技术升级提供了坚实的技术基础，确保了系统在分布式环境下的稳定运行和高效性能。