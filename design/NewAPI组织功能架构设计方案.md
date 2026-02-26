# New API 组织功能架构设计方案

## 文档信息
- **版本**: v1.0
- **创建日期**: 2026-01-28
- **状态**: 设计阶段

---

## 一、需求概述

### 1.1 核心需求
1. **组织概念引入**: 在系统中添加组织(Organization)实体，作为用户和资源管理的新层级
2. **组织创建与管理**: 用户可以创建组织，创建者自动成为组织管理员
3. **多组织归属**: 一个用户可以同时属于多个组织，支持组织切换
4. **组织级令牌**: 用户在选定组织下创建的令牌属于组织，组织成员共享可见
5. **组织级配额**: 用户在选定组织下充值的金额归属于组织配额池
6. **组织级限流**: 组织管理员可以控制组织内用户的限流策略

### 1.2 设计目标
- **多租户隔离**: 组织之间资源完全隔离，配额、令牌、日志独立管理
- **灵活权限**: 支持组织内角色分配（管理员、成员、财务等）
- **平滑迁移**: 兼容现有用户系统，支持个人账户与组织账户并存
- **易扩展性**: 为未来企业级功能（SSO、审计日志、成本中心）预留接口

---

## 二、系统架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    用户层 (User Layer)                        │
│  用户可以：                                                    │
│  1. 创建组织 (Owner)                                          │
│  2. 加入组织 (Member)                                         │
│  3. 在多个组织间切换                                          │
│  4. 使用个人配额 或 组织配额                                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   组织层 (Organization Layer)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Organization│  │Organization  │  │Organization  │      │
│  │      A       │  │      B       │  │      C       │      │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │      │
│  │  │ Quota  │  │  │  │ Quota  │  │  │  │ Quota  │  │      │
│  │  │ Tokens │  │  │  │ Tokens │  │  │  │ Tokens │  │      │
│  │  │ Members│  │  │  │ Members│  │  │  │ Members│  │      │
│  │  │ Logs   │  │  │  │ Logs   │  │  │  │ Logs   │  │      │
│  │  │ Limits │  │  │  │ Limits │  │  │  │ Limits │  │      │
│  │  └────────┘  │  │  └────────┘  │  │  └────────┘  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   资源层 (Resource Layer)                     │
│  • 令牌 (Token) - 归属组织或个人                              │
│  • 配额 (Quota) - 组织配额池 + 个人配额                       │
│  • 日志 (Log) - 按组织维度记录                                │
│  • 限流 (RateLimit) - 组织级 + 用户级                        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心实体关系

```
User (用户)
  ├─── 1:N ─── OrganizationMember (组织成员关系)
  │                  │
  │                  └─── N:1 ─── Organization (组织)
  │
  ├─── 1:N ─── Token (个人令牌)
  └─── 1:N ─── TopUp (个人充值)

Organization (组织)
  ├─── 1:N ─── OrganizationMember (成员)
  ├─── 1:N ─── OrganizationToken (组织令牌)
  ├─── 1:N ─── OrganizationTopUp (组织充值)
  ├─── 1:N ─── OrganizationRateLimit (组织限流策略)
  └─── 1:N ─── Log (API调用日志)
```

---

## 三、数据库设计

### 3.1 Organization (组织表)

```sql
CREATE TABLE organizations (
    id              INT AUTO_INCREMENT PRIMARY KEY COMMENT '组织ID',
    name            VARCHAR(100) NOT NULL COMMENT '组织名称',
    display_name    VARCHAR(200) COMMENT '组织显示名称',
    description     TEXT COMMENT '组织描述',
    owner_id        INT NOT NULL COMMENT '所有者用户ID',

    -- 配额管理
    quota           BIGINT DEFAULT 0 COMMENT '组织配额池',
    used_quota      BIGINT DEFAULT 0 COMMENT '已使用配额',
    request_count   INT DEFAULT 0 COMMENT '请求计数',

    -- 分组和设置
    `group`         VARCHAR(50) DEFAULT 'default' COMMENT '组织分组',
    settings        TEXT COMMENT '组织设置(JSON)',

    -- 状态管理
    status          TINYINT DEFAULT 1 COMMENT '状态: 1=启用, 2=禁用, 3=冻结',

    -- 元数据
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP NULL COMMENT '软删除时间',

    INDEX idx_owner_id (owner_id),
    INDEX idx_name (name),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织表';
```

**OrganizationSettings JSON结构**:
```json
{
  "maxMembers": 100,                    // 最大成员数
  "allowMemberCreateToken": true,       // 允许成员创建令牌
  "allowMemberViewLogs": false,         // 允许成员查看日志
  "quotaWarningThreshold": 0.1,         // 配额预警阈值(10%)
  "webhookUrl": "",                     // Webhook通知地址
  "notificationEmail": "",              // 通知邮箱
  "ssoEnabled": false,                  // SSO单点登录
  "ipWhitelist": []                     // IP白名单
}
```

### 3.2 OrganizationMember (组织成员关系表)

```sql
CREATE TABLE organization_members (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT NOT NULL COMMENT '组织ID',
    user_id         INT NOT NULL COMMENT '用户ID',

    -- 角色和权限
    role            VARCHAR(20) DEFAULT 'member' COMMENT '角色: owner/admin/finance/member',
    permissions     TEXT COMMENT '权限列表(JSON)',

    -- 个人配额
    quota_limit     BIGINT DEFAULT 0 COMMENT '该成员的配额限制(0=不限制)',
    used_quota      BIGINT DEFAULT 0 COMMENT '该成员已使用配额',

    -- 状态
    status          TINYINT DEFAULT 1 COMMENT '状态: 1=正常, 2=禁用',
    invited_by      INT COMMENT '邀请人用户ID',
    joined_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 元数据
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_org_user (organization_id, user_id),
    INDEX idx_user_id (user_id),
    INDEX idx_organization_id (organization_id),
    INDEX idx_role (role),

    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织成员表';
```

**角色定义**:
- `owner`: 组织所有者（创建者，唯一，可转让）
- `admin`: 组织管理员（可管理成员、令牌、配额）
- `finance`: 财务角色（可充值、查看账单）
- `member`: 普通成员（可使用令牌、查看自己的日志）

**Permissions JSON结构**:
```json
{
  "tokens": {
    "create": true,
    "view": true,
    "edit": false,
    "delete": false
  },
  "members": {
    "invite": false,
    "remove": false,
    "editRole": false
  },
  "quota": {
    "view": true,
    "topup": false
  },
  "logs": {
    "view": false,
    "export": false
  },
  "settings": {
    "view": false,
    "edit": false
  }
}
```

### 3.3 OrganizationToken (组织令牌表)

**方案一：扩展现有Token表**（推荐）
```sql
ALTER TABLE tokens ADD COLUMN organization_id INT DEFAULT NULL COMMENT '组织ID(NULL=个人令牌)';
ALTER TABLE tokens ADD INDEX idx_organization_id (organization_id);
ALTER TABLE tokens ADD CONSTRAINT fk_tokens_org
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
```

**方案二：新建独立表**
```sql
CREATE TABLE organization_tokens (
    id                  INT AUTO_INCREMENT PRIMARY KEY,
    organization_id     INT NOT NULL COMMENT '组织ID',
    created_by          INT NOT NULL COMMENT '创建者用户ID',

    -- 令牌信息（与tokens表结构一致）
    `key`               VARCHAR(48) NOT NULL UNIQUE COMMENT '令牌密钥',
    status              TINYINT DEFAULT 1,
    name                VARCHAR(100),
    expired_time        BIGINT DEFAULT -1,
    remain_quota        BIGINT DEFAULT 0,
    unlimited_quota     BOOLEAN DEFAULT FALSE,
    used_quota          BIGINT DEFAULT 0,

    -- 限制
    model_limits_enabled BOOLEAN DEFAULT FALSE,
    model_limits         TEXT,
    allow_ips            TEXT,
    `group`              VARCHAR(50),

    -- 元数据
    created_time        BIGINT,
    accessed_time       BIGINT,
    deleted_at          TIMESTAMP NULL,

    INDEX idx_organization_id (organization_id),
    INDEX idx_created_by (created_by),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**推荐方案一**，原因：
- 复用现有令牌管理逻辑
- 统一的令牌验证流程
- 减少代码重复

### 3.4 OrganizationTopUp (组织充值记录表)

**方案：扩展TopUp表**
```sql
ALTER TABLE topups ADD COLUMN organization_id INT DEFAULT NULL COMMENT '组织ID(NULL=个人充值)';
ALTER TABLE topups ADD INDEX idx_organization_id (organization_id);
```

### 3.5 OrganizationRateLimit (组织限流策略表)

```sql
CREATE TABLE organization_rate_limits (
    id                  INT AUTO_INCREMENT PRIMARY KEY,
    organization_id     INT NOT NULL COMMENT '组织ID',

    -- 组织级限流
    org_enabled         BOOLEAN DEFAULT TRUE COMMENT '组织级限流启用',
    org_duration_minutes INT DEFAULT 1 COMMENT '时间窗口(分钟)',
    org_max_requests    INT DEFAULT 10000 COMMENT '组织总请求上限',
    org_max_success     INT DEFAULT 8000 COMMENT '组织成功请求上限',

    -- 成员级限流
    member_enabled      BOOLEAN DEFAULT TRUE COMMENT '成员级限流启用',
    member_duration_minutes INT DEFAULT 1,
    member_max_requests INT DEFAULT 1000 COMMENT '单个成员请求上限',
    member_max_success  INT DEFAULT 800,

    -- 模型级限流(JSON)
    model_limits        TEXT COMMENT '按模型限流配置',

    -- 元数据
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_org_id (organization_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织限流策略表';
```

**ModelLimits JSON结构**:
```json
{
  "gpt-4": {
    "enabled": true,
    "maxRequests": 100,
    "durationMinutes": 1
  },
  "claude-3-5-sonnet": {
    "enabled": true,
    "maxRequests": 200,
    "durationMinutes": 1
  }
}
```

### 3.6 OrganizationInvite (组织邀请表) - 可选

```sql
CREATE TABLE organization_invites (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT NOT NULL,
    inviter_id      INT NOT NULL COMMENT '邀请人',

    -- 邀请方式
    invite_code     VARCHAR(32) UNIQUE COMMENT '邀请码',
    invite_email    VARCHAR(255) COMMENT '邀请邮箱',

    -- 角色和权限
    role            VARCHAR(20) DEFAULT 'member',
    permissions     TEXT,

    -- 状态
    status          TINYINT DEFAULT 1 COMMENT '1=待接受, 2=已接受, 3=已过期, 4=已拒绝',
    accepted_by     INT COMMENT '接受者用户ID',
    accepted_at     TIMESTAMP NULL,
    expired_at      TIMESTAMP NULL,

    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_organization_id (organization_id),
    INDEX idx_invite_code (invite_code),
    INDEX idx_invite_email (invite_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3.7 用户表扩展

```sql
ALTER TABLE users ADD COLUMN current_organization_id INT DEFAULT NULL COMMENT '当前选中的组织ID';
ALTER TABLE users ADD INDEX idx_current_organization_id (current_organization_id);
```

### 3.8 日志表扩展

```sql
ALTER TABLE logs ADD COLUMN organization_id INT DEFAULT NULL COMMENT '组织ID(NULL=个人使用)';
ALTER TABLE logs ADD INDEX idx_organization_id (organization_id);
```

---

## 四、API设计

### 4.1 组织管理 API

#### 4.1.1 创建组织
```
POST /api/organizations
Content-Type: application/json
Authorization: Bearer {user_token}

Request:
{
  "name": "acme-corp",
  "displayName": "Acme Corporation",
  "description": "我们公司的AI平台组织",
  "group": "default",
  "settings": {
    "maxMembers": 50
  }
}

Response:
{
  "success": true,
  "message": "组织创建成功",
  "data": {
    "id": 123,
    "name": "acme-corp",
    "displayName": "Acme Corporation",
    "ownerId": 456,
    "role": "owner",
    "createdAt": "2026-01-28T10:00:00Z"
  }
}
```

#### 4.1.2 获取用户的组织列表
```
GET /api/organizations
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 123,
      "name": "acme-corp",
      "displayName": "Acme Corporation",
      "role": "owner",
      "quota": 1000000,
      "usedQuota": 250000,
      "memberCount": 5,
      "status": 1
    },
    {
      "id": 124,
      "name": "startup-ai",
      "displayName": "Startup AI Lab",
      "role": "member",
      "quota": 500000,
      "usedQuota": 100000,
      "memberCount": 3,
      "status": 1
    }
  ]
}
```

#### 4.1.3 获取组织详情
```
GET /api/organizations/:id
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": {
    "id": 123,
    "name": "acme-corp",
    "displayName": "Acme Corporation",
    "description": "...",
    "ownerId": 456,
    "ownerUsername": "john_doe",
    "group": "default",
    "quota": 1000000,
    "usedQuota": 250000,
    "requestCount": 5000,
    "status": 1,
    "settings": {...},
    "memberCount": 5,
    "tokenCount": 10,
    "createdAt": "2026-01-28T10:00:00Z"
  }
}
```

#### 4.1.4 更新组织信息
```
PUT /api/organizations/:id
Authorization: Bearer {user_token}
Permission: owner or admin

Request:
{
  "displayName": "Acme Corp (Updated)",
  "description": "新的描述",
  "settings": {
    "maxMembers": 100
  }
}

Response:
{
  "success": true,
  "message": "组织信息已更新"
}
```

#### 4.1.5 删除组织
```
DELETE /api/organizations/:id
Authorization: Bearer {user_token}
Permission: owner only

Response:
{
  "success": true,
  "message": "组织已删除"
}
```

#### 4.1.6 切换当前组织
```
POST /api/user/switch-organization
Authorization: Bearer {user_token}

Request:
{
  "organizationId": 123  // 0 or null = 切换到个人模式
}

Response:
{
  "success": true,
  "message": "已切换到组织: Acme Corporation",
  "data": {
    "currentOrganizationId": 123,
    "organizationName": "acme-corp",
    "role": "admin"
  }
}
```

### 4.2 成员管理 API

#### 4.2.1 获取组织成员列表
```
GET /api/organizations/:id/members
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 1,
      "userId": 456,
      "username": "john_doe",
      "displayName": "John Doe",
      "email": "john@example.com",
      "role": "owner",
      "quotaLimit": 0,
      "usedQuota": 50000,
      "status": 1,
      "joinedAt": "2026-01-28T10:00:00Z"
    },
    {
      "id": 2,
      "userId": 457,
      "username": "jane_smith",
      "displayName": "Jane Smith",
      "email": "jane@example.com",
      "role": "admin",
      "quotaLimit": 200000,
      "usedQuota": 30000,
      "status": 1,
      "joinedAt": "2026-01-29T11:00:00Z"
    }
  ]
}
```

#### 4.2.2 邀请成员
```
POST /api/organizations/:id/members/invite
Authorization: Bearer {user_token}
Permission: owner, admin (if allowed by settings)

Request:
{
  "email": "newuser@example.com",  // 或 "username": "newuser"
  "role": "member",
  "quotaLimit": 100000,
  "permissions": {...}
}

Response:
{
  "success": true,
  "message": "邀请已发送",
  "data": {
    "inviteCode": "abc123def456",
    "inviteLink": "https://api.example.com/invite/abc123def456",
    "expiredAt": "2026-02-04T10:00:00Z"
  }
}
```

#### 4.2.3 接受邀请
```
POST /api/organizations/accept-invite
Authorization: Bearer {user_token}

Request:
{
  "inviteCode": "abc123def456"
}

Response:
{
  "success": true,
  "message": "已加入组织",
  "data": {
    "organizationId": 123,
    "organizationName": "acme-corp",
    "role": "member"
  }
}
```

#### 4.2.4 更新成员角色/权限
```
PUT /api/organizations/:orgId/members/:memberId
Authorization: Bearer {user_token}
Permission: owner, admin

Request:
{
  "role": "admin",
  "quotaLimit": 500000,
  "permissions": {...}
}

Response:
{
  "success": true,
  "message": "成员信息已更新"
}
```

#### 4.2.5 移除成员
```
DELETE /api/organizations/:orgId/members/:memberId
Authorization: Bearer {user_token}
Permission: owner, admin

Response:
{
  "success": true,
  "message": "成员已移除"
}
```

### 4.3 组织令牌 API

#### 4.3.1 创建组织令牌
```
POST /api/organizations/:id/tokens
Authorization: Bearer {user_token}
Permission: member (if allowMemberCreateToken=true), admin, owner

Request:
{
  "name": "Production API Key",
  "expiredTime": 1735689600,  // Unix timestamp
  "remainQuota": 100000,
  "unlimitedQuota": false,
  "modelLimitsEnabled": true,
  "modelLimits": "gpt-4,claude-3-5-sonnet"
}

Response:
{
  "success": true,
  "message": "组织令牌已创建",
  "data": {
    "id": 789,
    "key": "sk-org-abc123def456...",
    "name": "Production API Key",
    "organizationId": 123,
    "createdBy": 456,
    "createdTime": 1706428800
  }
}
```

#### 4.3.2 获取组织令牌列表
```
GET /api/organizations/:id/tokens
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 789,
      "key": "sk-org-abc123...",  // 部分隐藏
      "name": "Production API Key",
      "status": 1,
      "remainQuota": 80000,
      "usedQuota": 20000,
      "createdBy": 456,
      "createdByUsername": "john_doe",
      "createdTime": 1706428800,
      "accessedTime": 1706515200
    }
  ]
}
```

#### 4.3.3 更新/删除组织令牌
```
PUT /api/organizations/:orgId/tokens/:tokenId
DELETE /api/organizations/:orgId/tokens/:tokenId
```

### 4.4 组织充值 API

#### 4.4.1 获取组织充值记录
```
GET /api/organizations/:id/topups
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": [
    {
      "id": 1001,
      "organizationId": 123,
      "userId": 456,
      "username": "john_doe",
      "amount": 10000000,  // 配额
      "money": 20.0,       // 美元
      "tradeNo": "trade_123456",
      "paymentMethod": "stripe",
      "status": "success",
      "createTime": 1706428800,
      "completeTime": 1706428900
    }
  ]
}
```

#### 4.4.2 组织充值
```
POST /api/organizations/:id/topup
Authorization: Bearer {user_token}
Permission: finance, admin, owner

Request:
{
  "amount": 10000000,      // 配额数量
  "paymentMethod": "stripe"
}

Response:
{
  "success": true,
  "message": "支付订单已创建",
  "data": {
    "tradeNo": "trade_123456",
    "paymentUrl": "https://checkout.stripe.com/...",
    "amount": 10000000,
    "money": 20.0
  }
}
```

### 4.5 组织限流 API

#### 4.5.1 获取组织限流策略
```
GET /api/organizations/:id/rate-limits
Authorization: Bearer {user_token}
Permission: admin, owner

Response:
{
  "success": true,
  "data": {
    "organizationId": 123,
    "orgEnabled": true,
    "orgDurationMinutes": 1,
    "orgMaxRequests": 10000,
    "orgMaxSuccess": 8000,
    "memberEnabled": true,
    "memberDurationMinutes": 1,
    "memberMaxRequests": 1000,
    "memberMaxSuccess": 800,
    "modelLimits": {
      "gpt-4": {
        "enabled": true,
        "maxRequests": 100,
        "durationMinutes": 1
      }
    }
  }
}
```

#### 4.5.2 更新组织限流策略
```
PUT /api/organizations/:id/rate-limits
Authorization: Bearer {user_token}
Permission: admin, owner

Request:
{
  "orgEnabled": true,
  "orgMaxRequests": 20000,
  "memberMaxRequests": 2000,
  "modelLimits": {
    "gpt-4": {
      "enabled": true,
      "maxRequests": 200,
      "durationMinutes": 1
    }
  }
}

Response:
{
  "success": true,
  "message": "限流策略已更新"
}
```

### 4.6 组织统计 API

#### 4.6.1 组织使用统计
```
GET /api/organizations/:id/statistics
Authorization: Bearer {user_token}

Query Parameters:
- start_time: Unix timestamp
- end_time: Unix timestamp
- group_by: day|hour|model|user

Response:
{
  "success": true,
  "data": {
    "totalRequests": 50000,
    "totalQuota": 1000000,
    "totalUsedQuota": 250000,
    "successRate": 0.98,
    "avgResponseTime": 1200,
    "topModels": [
      {"model": "gpt-4", "requests": 20000, "quota": 150000},
      {"model": "claude-3-5-sonnet", "requests": 15000, "quota": 80000}
    ],
    "topUsers": [
      {"userId": 456, "username": "john_doe", "requests": 10000, "quota": 60000},
      {"userId": 457, "username": "jane_smith", "requests": 8000, "quota": 50000}
    ]
  }
}
```

---

## 五、权限控制设计

### 5.1 权限矩阵

| 操作 | Owner | Admin | Finance | Member |
|------|-------|-------|---------|--------|
| **组织管理** |
| 查看组织信息 | ✅ | ✅ | ✅ | ✅ |
| 编辑组织信息 | ✅ | ✅ | ❌ | ❌ |
| 删除组织 | ✅ | ❌ | ❌ | ❌ |
| 转让所有权 | ✅ | ❌ | ❌ | ❌ |
| **成员管理** |
| 查看成员列表 | ✅ | ✅ | ✅ | ⚠️* |
| 邀请成员 | ✅ | ✅ | ❌ | ❌ |
| 移除成员 | ✅ | ✅ | ❌ | ❌ |
| 修改成员角色 | ✅ | ⚠️** | ❌ | ❌ |
| **令牌管理** |
| 查看所有令牌 | ✅ | ✅ | ❌ | ❌ |
| 创建令牌 | ✅ | ✅ | ❌ | ⚠️*** |
| 编辑/删除自己创建的令牌 | ✅ | ✅ | ❌ | ✅ |
| 编辑/删除他人创建的令牌 | ✅ | ✅ | ❌ | ❌ |
| **配额管理** |
| 查看配额 | ✅ | ✅ | ✅ | ✅ |
| 充值 | ✅ | ✅ | ✅ | ❌ |
| 查看充值记录 | ✅ | ✅ | ✅ | ❌ |
| 分配成员配额限制 | ✅ | ✅ | ❌ | ❌ |
| **日志和统计** |
| 查看组织日志 | ✅ | ✅ | ❌ | ⚠️**** |
| 查看统计数据 | ✅ | ✅ | ✅ | ⚠️**** |
| 导出报表 | ✅ | ✅ | ✅ | ❌ |
| **限流管理** |
| 查看限流策略 | ✅ | ✅ | ❌ | ❌ |
| 修改限流策略 | ✅ | ✅ | ❌ | ❌ |

**注释**:
- ⚠️* : 仅能查看基本信息
- ⚠️** : Admin只能修改Member角色，不能修改Owner/Admin
- ⚠️*** : 取决于组织设置 `allowMemberCreateToken`
- ⚠️**** : 仅能查看自己的数据

### 5.2 中间件实现

#### 5.2.1 组织上下文中间件
```go
// middleware/organization.go
func OrganizationContext() gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.GetInt("id")

        // 从用户表获取当前选中的组织
        user, err := model.GetUserById(userId, false)
        if err != nil {
            c.Next()
            return
        }

        if user.CurrentOrganizationId > 0 {
            // 验证用户是否为该组织成员
            member, err := model.GetOrganizationMember(
                user.CurrentOrganizationId,
                userId,
            )
            if err == nil && member.Status == 1 {
                c.Set("organization_id", user.CurrentOrganizationId)
                c.Set("organization_role", member.Role)
                c.Set("organization_permissions", member.Permissions)
            }
        }

        c.Next()
    }
}
```

#### 5.2.2 组织权限验证中间件
```go
// middleware/organization.go
func OrganizationAuth(requiredRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        orgId, exists := c.Get("organization_id")
        if !exists {
            c.JSON(403, gin.H{
                "success": false,
                "message": "需要选择组织",
            })
            c.Abort()
            return
        }

        role, _ := c.GetString("organization_role")

        // 检查角色权限
        hasPermission := false
        for _, requiredRole := range requiredRoles {
            if role == requiredRole {
                hasPermission = true
                break
            }
        }

        // Owner和Admin有高级权限
        if role == "owner" || (role == "admin" && len(requiredRoles) > 0) {
            hasPermission = true
        }

        if !hasPermission {
            c.JSON(403, gin.H{
                "success": false,
                "message": "权限不足",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 5.2.3 使用示例
```go
// router/api-router.go
orgGroup := apiGroup.Group("/organizations")
orgGroup.Use(middleware.UserAuth())
orgGroup.Use(middleware.OrganizationContext())
{
    // 所有用户可访问
    orgGroup.GET("", controller.GetUserOrganizations)
    orgGroup.POST("", controller.CreateOrganization)

    // 需要组织权限
    orgGroup.GET("/:id", controller.GetOrganization)

    // 需要admin或owner权限
    orgGroup.PUT("/:id",
        middleware.OrganizationAuth("admin", "owner"),
        controller.UpdateOrganization,
    )

    // 仅owner可访问
    orgGroup.DELETE("/:id",
        middleware.OrganizationAuth("owner"),
        controller.DeleteOrganization,
    )

    // 成员管理
    memberGroup := orgGroup.Group("/:id/members")
    {
        memberGroup.GET("", controller.GetOrganizationMembers)
        memberGroup.POST("/invite",
            middleware.OrganizationAuth("admin", "owner"),
            controller.InviteOrganizationMember,
        )
        memberGroup.DELETE("/:memberId",
            middleware.OrganizationAuth("admin", "owner"),
            controller.RemoveOrganizationMember,
        )
    }
}
```

### 5.3 细粒度权限检查

```go
// service/organization_permission.go
type OrganizationPermission struct {
    UserId         int
    OrganizationId int
    Role           string
    Permissions    map[string]interface{}
}

func (op *OrganizationPermission) Can(action string) bool {
    // Owner拥有所有权限
    if op.Role == "owner" {
        return true
    }

    // 检查细粒度权限
    parts := strings.Split(action, ".")
    if len(parts) < 2 {
        return false
    }

    resource := parts[0]  // "tokens", "members", "quota"
    operation := parts[1] // "create", "view", "edit", "delete"

    if resourcePerms, ok := op.Permissions[resource].(map[string]interface{}); ok {
        if allowed, ok := resourcePerms[operation].(bool); ok {
            return allowed
        }
    }

    // Admin的默认权限
    if op.Role == "admin" {
        adminDefaultPerms := map[string][]string{
            "tokens":  {"create", "view", "edit", "delete"},
            "members": {"invite", "view"},
            "quota":   {"view"},
            "logs":    {"view"},
        }

        if ops, ok := adminDefaultPerms[resource]; ok {
            for _, op := range ops {
                if op == operation {
                    return true
                }
            }
        }
    }

    return false
}

// 使用示例
func CreateOrganizationToken(c *gin.Context) {
    orgId := c.GetInt("organization_id")
    userId := c.GetInt("id")
    role := c.GetString("organization_role")
    permissions := c.GetString("organization_permissions")

    perm := &OrganizationPermission{
        UserId:         userId,
        OrganizationId: orgId,
        Role:           role,
        Permissions:    parsePermissions(permissions),
    }

    if !perm.Can("tokens.create") {
        c.JSON(403, gin.H{"message": "无权创建令牌"})
        return
    }

    // 创建令牌逻辑...
}
```

---

## 六、业务逻辑实现

### 6.1 组织创建流程

```go
// service/organization.go
func CreateOrganization(userId int, req *dto.CreateOrganizationRequest) (*model.Organization, error) {
    // 1. 验证用户权限（可选：限制普通用户创建组织数量）
    userOrgCount, _ := model.GetUserOrganizationCount(userId)
    if userOrgCount >= setting.MaxOrganizationsPerUser {
        return nil, errors.New("超出组织创建数量限制")
    }

    // 2. 验证组织名称唯一性
    exists, _ := model.IsOrganizationNameTaken(req.Name)
    if exists {
        return nil, errors.New("组织名称已被占用")
    }

    // 3. 创建组织
    org := &model.Organization{
        Name:         req.Name,
        DisplayName:  req.DisplayName,
        Description:  req.Description,
        OwnerId:      userId,
        Group:        req.Group,
        Quota:        0,
        UsedQuota:    0,
        Status:       1,
        Settings:     req.Settings,
    }

    err := org.Insert()
    if err != nil {
        return nil, err
    }

    // 4. 创建者自动成为Owner成员
    member := &model.OrganizationMember{
        OrganizationId: org.Id,
        UserId:         userId,
        Role:           "owner",
        Status:         1,
        Permissions:    getDefaultOwnerPermissions(),
    }

    err = member.Insert()
    if err != nil {
        // 回滚组织创建
        org.Delete()
        return nil, err
    }

    // 5. 初始化组织限流策略
    rateLimit := &model.OrganizationRateLimit{
        OrganizationId:         org.Id,
        OrgEnabled:             true,
        OrgDurationMinutes:     1,
        OrgMaxRequests:         10000,
        OrgMaxSuccess:          8000,
        MemberEnabled:          true,
        MemberDurationMinutes:  1,
        MemberMaxRequests:      1000,
        MemberMaxSuccess:       800,
    }
    rateLimit.Insert()

    // 6. 记录操作日志
    common.LogInfo(fmt.Sprintf("User %d created organization %d", userId, org.Id))

    return org, nil
}
```

### 6.2 组织切换流程

```go
// service/organization.go
func SwitchOrganization(userId int, orgId int) error {
    // 1. 如果orgId=0，切换到个人模式
    if orgId == 0 {
        return model.UpdateUserCurrentOrganization(userId, 0)
    }

    // 2. 验证用户是否为该组织成员
    member, err := model.GetOrganizationMember(orgId, userId)
    if err != nil {
        return errors.New("您不是该组织的成员")
    }

    if member.Status != 1 {
        return errors.New("您在该组织中的账户已被禁用")
    }

    // 3. 验证组织状态
    org, err := model.GetOrganizationById(orgId)
    if err != nil {
        return errors.New("组织不存在")
    }

    if org.Status != 1 {
        return errors.New("该组织已被禁用")
    }

    // 4. 更新用户当前组织
    err = model.UpdateUserCurrentOrganization(userId, orgId)
    if err != nil {
        return err
    }

    // 5. 清除用户缓存（包含currentOrganizationId）
    model.InvalidateUserCache(userId)

    // 6. 记录切换日志
    common.LogInfo(fmt.Sprintf("User %d switched to organization %d", userId, orgId))

    return nil
}
```

### 6.3 组织令牌创建流程

```go
// controller/token.go (扩展)
func CreateOrganizationToken(c *gin.Context) {
    userId := c.GetInt("id")
    orgId := c.GetInt("organization_id")

    if orgId == 0 {
        c.JSON(400, gin.H{"message": "请先选择组织"})
        return
    }

    var req dto.CreateTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"message": err.Error()})
        return
    }

    // 1. 权限检查（通过中间件已完成）

    // 2. 验证配额限制
    org, err := model.GetOrganizationById(orgId)
    if err != nil {
        c.JSON(500, gin.H{"message": "获取组织信息失败"})
        return
    }

    if req.RemainQuota > org.Quota {
        c.JSON(400, gin.H{"message": "令牌配额超出组织可用配额"})
        return
    }

    // 3. 创建令牌（扩展Token模型）
    token := &model.Token{
        UserId:             userId,
        OrganizationId:     &orgId,  // 设置组织ID
        Key:                generateTokenKey("sk-org-"),
        Name:               req.Name,
        Status:             1,
        ExpiredTime:        req.ExpiredTime,
        RemainQuota:        req.RemainQuota,
        UnlimitedQuota:     req.UnlimitedQuota,
        ModelLimitsEnabled: req.ModelLimitsEnabled,
        ModelLimits:        req.ModelLimits,
        Group:              org.Group,
    }

    err = token.Insert()
    if err != nil {
        c.JSON(500, gin.H{"message": "令牌创建失败"})
        return
    }

    // 4. 记录日志
    model.RecordLog(userId, model.LogTypeCreateToken, fmt.Sprintf(
        "Created organization token: %s for org %d",
        token.Name,
        orgId,
    ))

    c.JSON(200, gin.H{
        "success": true,
        "message": "组织令牌创建成功",
        "data":    token,
    })
}
```

### 6.4 组织充值流程

```go
// controller/topup.go (扩展)
func OrganizationTopup(c *gin.Context) {
    userId := c.GetInt("id")
    orgId := c.GetInt("organization_id")

    if orgId == 0 {
        c.JSON(400, gin.H{"message": "请先选择组织"})
        return
    }

    // 1. 权限检查（finance/admin/owner）
    role := c.GetString("organization_role")
    if role != "owner" && role != "admin" && role != "finance" {
        c.JSON(403, gin.H{"message": "权限不足"})
        return
    }

    var req dto.TopupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"message": err.Error()})
        return
    }

    // 2. 获取组织信息
    org, err := model.GetOrganizationById(orgId)
    if err != nil {
        c.JSON(500, gin.H{"message": "获取组织信息失败"})
        return
    }

    // 3. 计算支付金额（应用组织分组倍率）
    groupRatio := setting.GetGroupRatio(org.Group)
    money := getPayMoney(req.Amount, groupRatio)

    // 4. 创建充值订单
    topup := &model.TopUp{
        UserId:            userId,
        OrganizationId:    &orgId,  // 设置组织ID
        Amount:            req.Amount,
        Money:             money,
        PaymentMethod:     req.PaymentMethod,
        Status:            "pending",
        CreateTime:        time.Now().Unix(),
    }

    err = topup.Insert()
    if err != nil {
        c.JSON(500, gin.H{"message": "订单创建失败"})
        return
    }

    // 5. 生成支付链接（根据支付方式）
    var paymentUrl string
    switch req.PaymentMethod {
    case "stripe":
        paymentUrl, err = createStripeCheckout(topup, org)
    case "epay":
        paymentUrl, err = createEpayOrder(topup, org)
    default:
        c.JSON(400, gin.H{"message": "不支持的支付方式"})
        return
    }

    if err != nil {
        c.JSON(500, gin.H{"message": "支付订单创建失败"})
        return
    }

    c.JSON(200, gin.H{
        "success": true,
        "data": gin.H{
            "tradeNo":    topup.TradeNo,
            "paymentUrl": paymentUrl,
            "amount":     req.Amount,
            "money":      money,
        },
    })
}

// 支付回调处理（扩展）
func handleOrganizationTopupCallback(topup *model.TopUp) error {
    if topup.OrganizationId == nil {
        // 个人充值，使用原有逻辑
        return handlePersonalTopup(topup)
    }

    // 组织充值
    orgId := *topup.OrganizationId

    // 1. 计算配额
    quota := int(topup.Money * common.QuotaPerUnit)

    // 2. 增加组织配额
    err := model.IncreaseOrganizationQuota(orgId, quota)
    if err != nil {
        return err
    }

    // 3. 更新订单状态
    topup.Status = "success"
    topup.CompleteTime = time.Now().Unix()
    err = topup.Update()
    if err != nil {
        return err
    }

    // 4. 发送通知（Webhook/邮件）
    org, _ := model.GetOrganizationById(orgId)
    sendTopupNotification(org, topup, quota)

    // 5. 记录日志
    model.RecordLog(topup.UserId, model.LogTypeTopup, fmt.Sprintf(
        "Organization %d topup: $%.2f, quota: %d",
        orgId,
        topup.Money,
        quota,
    ))

    return nil
}
```

### 6.5 组织限流检查流程

```go
// middleware/organization_rate_limit.go
func OrganizationRateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        orgId, exists := c.Get("organization_id")
        if !exists {
            // 个人模式，使用原有限流逻辑
            c.Next()
            return
        }

        userId := c.GetInt("id")

        // 1. 获取组织限流策略
        rateLimit, err := model.GetOrganizationRateLimit(orgId.(int))
        if err != nil {
            c.Next()
            return
        }

        // 2. 组织级限流检查
        if rateLimit.OrgEnabled {
            key := fmt.Sprintf("org_rate_limit:%d", orgId)

            allowed, err := checkRateLimit(
                key,
                rateLimit.OrgMaxRequests,
                rateLimit.OrgDurationMinutes * 60,
            )

            if !allowed {
                c.JSON(429, gin.H{
                    "success": false,
                    "message": "组织请求频率超限",
                })
                c.Abort()
                return
            }
        }

        // 3. 成员级限流检查
        if rateLimit.MemberEnabled {
            key := fmt.Sprintf("org_member_rate_limit:%d:%d", orgId, userId)

            allowed, err := checkRateLimit(
                key,
                rateLimit.MemberMaxRequests,
                rateLimit.MemberDurationMinutes * 60,
            )

            if !allowed {
                c.JSON(429, gin.H{
                    "success": false,
                    "message": "个人请求频率超限",
                })
                c.Abort()
                return
            }
        }

        // 4. 模型级限流检查
        modelName := c.GetString("model")
        if modelName != "" && rateLimit.ModelLimits != "" {
            modelLimits := parseModelLimits(rateLimit.ModelLimits)
            if limit, ok := modelLimits[modelName]; ok && limit.Enabled {
                key := fmt.Sprintf("org_model_rate_limit:%d:%s", orgId, modelName)

                allowed, err := checkRateLimit(
                    key,
                    limit.MaxRequests,
                    limit.DurationMinutes * 60,
                )

                if !allowed {
                    c.JSON(429, gin.H{
                        "success": false,
                        "message": fmt.Sprintf("模型 %s 请求频率超限", modelName),
                    })
                    c.Abort()
                    return
                }
            }
        }

        c.Next()
    }
}
```

### 6.6 组织配额扣减流程

```go
// service/quota.go (扩展)
func PreConsumeQuotaWithOrganization(quota int, relayInfo *RelayInfo) error {
    userId := relayInfo.UserId
    orgId := relayInfo.OrganizationId

    if orgId == 0 {
        // 个人模式，使用原有逻辑
        return PreConsumeQuota(quota, relayInfo)
    }

    // 1. 获取组织配额
    org, err := model.GetOrganizationById(orgId)
    if err != nil {
        return errors.New("获取组织信息失败")
    }

    if org.Quota < quota {
        return errors.New("组织配额不足")
    }

    // 2. 检查成员配额限制
    member, err := model.GetOrganizationMember(orgId, userId)
    if err != nil {
        return errors.New("获取成员信息失败")
    }

    if member.QuotaLimit > 0 {
        if member.QuotaLimit - member.UsedQuota < quota {
            return errors.New("个人配额限制已达上限")
        }
    }

    // 3. 扣减组织配额
    err = model.DecreaseOrganizationQuota(orgId, quota)
    if err != nil {
        return err
    }

    // 4. 扣减令牌配额（如果有）
    if relayInfo.TokenId > 0 {
        token, _ := model.GetTokenById(relayInfo.TokenId)
        if token != nil && !token.UnlimitedQuota {
            err = model.DecreaseTokenQuota(relayInfo.TokenId, token.Key, quota)
            if err != nil {
                // 回滚组织配额
                model.IncreaseOrganizationQuota(orgId, quota)
                return err
            }
        }
    }

    // 5. 记录成员使用量
    err = model.IncreaseOrganizationMemberUsedQuota(orgId, userId, quota)
    if err != nil {
        // 不阻断流程，仅记录日志
        common.LogError(fmt.Sprintf("Failed to update member used quota: %v", err))
    }

    relayInfo.FinalPreConsumedQuota = quota

    return nil
}

// 后扣费结算（组织模式）
func PostConsumeQuotaWithOrganization(quotaDelta int, relayInfo *RelayInfo) error {
    if quotaDelta == 0 {
        return nil
    }

    orgId := relayInfo.OrganizationId
    userId := relayInfo.UserId

    if quotaDelta > 0 {
        // 补扣配额
        err := model.DecreaseOrganizationQuota(orgId, quotaDelta)
        if err != nil {
            return err
        }

        model.IncreaseOrganizationMemberUsedQuota(orgId, userId, quotaDelta)
        model.IncreaseOrganizationUsedQuota(orgId, quotaDelta)

    } else {
        // 返还配额
        refundQuota := -quotaDelta
        model.IncreaseOrganizationQuota(orgId, refundQuota)
        model.DecreaseOrganizationMemberUsedQuota(orgId, userId, refundQuota)
        model.DecreaseOrganizationUsedQuota(orgId, refundQuota)
    }

    return nil
}
```

---

## 七、前端设计

### 7.1 组织切换器组件

```jsx
// components/OrganizationSwitcher.jsx
import React, { useState, useEffect } from 'react';
import { Select, Avatar, Spin } from '@douyinfe/semi-ui';
import { API } from '../helpers';

const OrganizationSwitcher = () => {
  const [organizations, setOrganizations] = useState([]);
  const [currentOrgId, setCurrentOrgId] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadOrganizations();
  }, []);

  const loadOrganizations = async () => {
    try {
      const res = await API.get('/api/organizations');
      if (res.data.success) {
        setOrganizations(res.data.data);

        // 获取当前组织
        const user = await API.get('/api/user/self');
        setCurrentOrgId(user.data.currentOrganizationId || 0);
      }
    } catch (error) {
      console.error('加载组织列表失败', error);
    }
  };

  const handleSwitch = async (value) => {
    setLoading(true);
    try {
      const res = await API.post('/api/user/switch-organization', {
        organizationId: value === 'personal' ? 0 : parseInt(value)
      });

      if (res.data.success) {
        setCurrentOrgId(value === 'personal' ? 0 : parseInt(value));
        window.location.reload(); // 刷新页面以更新上下文
      }
    } catch (error) {
      console.error('切换组织失败', error);
    } finally {
      setLoading(false);
    }
  };

  const options = [
    {
      value: 'personal',
      label: (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar size="small" style={{ marginRight: 8 }}>个</Avatar>
          <span>个人账户</span>
        </div>
      )
    },
    ...organizations.map(org => ({
      value: org.id.toString(),
      label: (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar size="small" style={{ marginRight: 8 }}>
            {org.displayName.charAt(0)}
          </Avatar>
          <div>
            <div>{org.displayName}</div>
            <div style={{ fontSize: 12, color: '#999' }}>
              {org.role === 'owner' ? '所有者' :
               org.role === 'admin' ? '管理员' : '成员'}
            </div>
          </div>
        </div>
      )
    }))
  ];

  return (
    <Select
      value={currentOrgId === 0 ? 'personal' : currentOrgId.toString()}
      onChange={handleSwitch}
      style={{ width: 240 }}
      disabled={loading}
      prefix={loading ? <Spin size="small" /> : null}
    >
      {options.map(opt => (
        <Select.Option key={opt.value} value={opt.value}>
          {opt.label}
        </Select.Option>
      ))}
    </Select>
  );
};

export default OrganizationSwitcher;
```

### 7.2 组织管理页面

```jsx
// pages/Organization/OrganizationList.jsx
import React, { useState, useEffect } from 'react';
import { Table, Button, Card, Tag, Modal, Form } from '@douyinfe/semi-ui';
import { API, showSuccess, showError } from '../../helpers';

const OrganizationList = () => {
  const [organizations, setOrganizations] = useState([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);

  const columns = [
    {
      title: '组织名称',
      dataIndex: 'displayName',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 'bold' }}>{text}</div>
          <div style={{ fontSize: 12, color: '#999' }}>@{record.name}</div>
        </div>
      )
    },
    {
      title: '角色',
      dataIndex: 'role',
      render: (role) => {
        const colorMap = {
          owner: 'blue',
          admin: 'green',
          finance: 'orange',
          member: 'grey'
        };
        const textMap = {
          owner: '所有者',
          admin: '管理员',
          finance: '财务',
          member: '成员'
        };
        return <Tag color={colorMap[role]}>{textMap[role]}</Tag>;
      }
    },
    {
      title: '配额',
      dataIndex: 'quota',
      render: (quota, record) => (
        <div>
          <div>剩余: {(quota / 500000).toFixed(2)} 美元</div>
          <div style={{ fontSize: 12, color: '#999' }}>
            已用: {(record.usedQuota / 500000).toFixed(2)} 美元
          </div>
        </div>
      )
    },
    {
      title: '成员数',
      dataIndex: 'memberCount'
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? '正常' : '禁用'}
        </Tag>
      )
    },
    {
      title: '操作',
      render: (record) => (
        <div>
          <Button
            size="small"
            onClick={() => window.location.href = `/organization/${record.id}`}
          >
            管理
          </Button>
          {record.role === 'owner' && (
            <Button
              size="small"
              type="danger"
              style={{ marginLeft: 8 }}
              onClick={() => handleDelete(record.id)}
            >
              删除
            </Button>
          )}
        </div>
      )
    }
  ];

  useEffect(() => {
    loadOrganizations();
  }, []);

  const loadOrganizations = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/organizations');
      if (res.data.success) {
        setOrganizations(res.data.data);
      }
    } catch (error) {
      showError('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (values) => {
    try {
      const res = await API.post('/api/organizations', values);
      if (res.data.success) {
        showSuccess('组织创建成功');
        setCreateModalVisible(false);
        loadOrganizations();
      }
    } catch (error) {
      showError('创建失败');
    }
  };

  return (
    <div>
      <Card
        title="我的组织"
        headerExtraContent={
          <Button onClick={() => setCreateModalVisible(true)}>
            创建组织
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={organizations}
          loading={loading}
          pagination={false}
        />
      </Card>

      <Modal
        title="创建组织"
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
      >
        <Form onSubmit={handleCreate}>
          <Form.Input
            field="name"
            label="组织标识"
            placeholder="acme-corp"
            rules={[{ required: true, message: '请输入组织标识' }]}
          />
          <Form.Input
            field="displayName"
            label="显示名称"
            placeholder="Acme Corporation"
            rules={[{ required: true, message: '请输入显示名称' }]}
          />
          <Form.TextArea
            field="description"
            label="描述"
            placeholder="组织描述"
          />
          <Button htmlType="submit" type="primary" block>
            创建
          </Button>
        </Form>
      </Modal>
    </div>
  );
};

export default OrganizationList;
```

### 7.3 组织详情页面

```jsx
// pages/Organization/OrganizationDetail.jsx
import React, { useState, useEffect } from 'react';
import { Tabs, Card, Button, Descriptions, Tag } from '@douyinfe/semi-ui';
import { useParams } from 'react-router-dom';
import { API } from '../../helpers';

// 导入子组件
import OrganizationMembers from './components/OrganizationMembers';
import OrganizationTokens from './components/OrganizationTokens';
import OrganizationTopups from './components/OrganizationTopups';
import OrganizationRateLimits from './components/OrganizationRateLimits';
import OrganizationSettings from './components/OrganizationSettings';

const OrganizationDetail = () => {
  const { id } = useParams();
  const [organization, setOrganization] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadOrganization();
  }, [id]);

  const loadOrganization = async () => {
    try {
      const res = await API.get(`/api/organizations/${id}`);
      if (res.data.success) {
        setOrganization(res.data.data);
      }
    } catch (error) {
      console.error('加载组织失败', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading || !organization) {
    return <div>加载中...</div>;
  }

  return (
    <div>
      {/* 组织概览 */}
      <Card style={{ marginBottom: 16 }}>
        <Descriptions row>
          <Descriptions.Item itemKey="组织名称">
            {organization.displayName}
          </Descriptions.Item>
          <Descriptions.Item itemKey="标识">
            @{organization.name}
          </Descriptions.Item>
          <Descriptions.Item itemKey="所有者">
            {organization.ownerUsername}
          </Descriptions.Item>
          <Descriptions.Item itemKey="配额">
            剩余: {(organization.quota / 500000).toFixed(2)} 美元 /
            已用: {(organization.usedQuota / 500000).toFixed(2)} 美元
          </Descriptions.Item>
          <Descriptions.Item itemKey="成员数">
            {organization.memberCount}
          </Descriptions.Item>
          <Descriptions.Item itemKey="令牌数">
            {organization.tokenCount}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 选项卡 */}
      <Tabs type="line">
        <Tabs.TabPane tab="成员" itemKey="members">
          <OrganizationMembers organizationId={id} />
        </Tabs.TabPane>

        <Tabs.TabPane tab="令牌" itemKey="tokens">
          <OrganizationTokens organizationId={id} />
        </Tabs.TabPane>

        <Tabs.TabPane tab="充值记录" itemKey="topups">
          <OrganizationTopups organizationId={id} />
        </Tabs.TabPane>

        <Tabs.TabPane tab="限流策略" itemKey="ratelimits">
          <OrganizationRateLimits organizationId={id} />
        </Tabs.TabPane>

        <Tabs.TabPane tab="设置" itemKey="settings">
          <OrganizationSettings organizationId={id} organization={organization} />
        </Tabs.TabPane>
      </Tabs>
    </div>
  );
};

export default OrganizationDetail;
```

### 7.4 路由配置

```jsx
// App.jsx
import OrganizationList from './pages/Organization/OrganizationList';
import OrganizationDetail from './pages/Organization/OrganizationDetail';

// 路由配置
<Route path="/organizations" element={<OrganizationList />} />
<Route path="/organization/:id" element={<OrganizationDetail />} />
```

---

## 八、迁移方案

### 8.1 数据库迁移脚本

```sql
-- migration_v1.0_organizations.sql

-- 1. 创建组织表
CREATE TABLE IF NOT EXISTS organizations (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    display_name    VARCHAR(200),
    description     TEXT,
    owner_id        INT NOT NULL,
    quota           BIGINT DEFAULT 0,
    used_quota      BIGINT DEFAULT 0,
    request_count   INT DEFAULT 0,
    `group`         VARCHAR(50) DEFAULT 'default',
    settings        TEXT,
    status          TINYINT DEFAULT 1,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP NULL,
    INDEX idx_owner_id (owner_id),
    INDEX idx_name (name),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. 创建组织成员表
CREATE TABLE IF NOT EXISTS organization_members (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT NOT NULL,
    user_id         INT NOT NULL,
    role            VARCHAR(20) DEFAULT 'member',
    permissions     TEXT,
    quota_limit     BIGINT DEFAULT 0,
    used_quota      BIGINT DEFAULT 0,
    status          TINYINT DEFAULT 1,
    invited_by      INT,
    joined_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_org_user (organization_id, user_id),
    INDEX idx_user_id (user_id),
    INDEX idx_organization_id (organization_id),
    INDEX idx_role (role),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 3. 创建组织限流策略表
CREATE TABLE IF NOT EXISTS organization_rate_limits (
    id                      INT AUTO_INCREMENT PRIMARY KEY,
    organization_id         INT NOT NULL,
    org_enabled             BOOLEAN DEFAULT TRUE,
    org_duration_minutes    INT DEFAULT 1,
    org_max_requests        INT DEFAULT 10000,
    org_max_success         INT DEFAULT 8000,
    member_enabled          BOOLEAN DEFAULT TRUE,
    member_duration_minutes INT DEFAULT 1,
    member_max_requests     INT DEFAULT 1000,
    member_max_success      INT DEFAULT 800,
    model_limits            TEXT,
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_org_id (organization_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 4. 创建组织邀请表
CREATE TABLE IF NOT EXISTS organization_invites (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT NOT NULL,
    inviter_id      INT NOT NULL,
    invite_code     VARCHAR(32) UNIQUE,
    invite_email    VARCHAR(255),
    role            VARCHAR(20) DEFAULT 'member',
    permissions     TEXT,
    status          TINYINT DEFAULT 1,
    accepted_by     INT,
    accepted_at     TIMESTAMP NULL,
    expired_at      TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_organization_id (organization_id),
    INDEX idx_invite_code (invite_code),
    INDEX idx_invite_email (invite_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 5. 扩展users表
ALTER TABLE users
ADD COLUMN current_organization_id INT DEFAULT NULL AFTER `group`,
ADD INDEX idx_current_organization_id (current_organization_id);

-- 6. 扩展tokens表
ALTER TABLE tokens
ADD COLUMN organization_id INT DEFAULT NULL AFTER user_id,
ADD INDEX idx_organization_id (organization_id);

-- 7. 扩展topups表
ALTER TABLE topups
ADD COLUMN organization_id INT DEFAULT NULL AFTER user_id,
ADD INDEX idx_organization_id (organization_id);

-- 8. 扩展logs表
ALTER TABLE logs
ADD COLUMN organization_id INT DEFAULT NULL AFTER user_id,
ADD INDEX idx_organization_id (organization_id);

-- 9. 创建组织配额历史表（可选，用于审计）
CREATE TABLE IF NOT EXISTS organization_quota_history (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT NOT NULL,
    user_id         INT,
    action          VARCHAR(20) NOT NULL COMMENT 'topup/consume/refund/adjust',
    quota_delta     BIGINT NOT NULL,
    quota_before    BIGINT NOT NULL,
    quota_after     BIGINT NOT NULL,
    reason          VARCHAR(255),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_organization_id (organization_id),
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 8.2 数据兼容性处理

**原则**:
- 现有个人账户不受影响，继续使用个人配额和令牌
- `organization_id = NULL` 表示个人模式
- 用户可以同时拥有个人账户和组织账户

**代码兼容**:
```go
// 在所有配额/令牌/日志操作中添加组织判断
func getQuotaSource(c *gin.Context) (isOrg bool, id int) {
    orgId, exists := c.Get("organization_id")
    if exists && orgId.(int) > 0 {
        return true, orgId.(int)
    }

    userId := c.GetInt("id")
    return false, userId
}

// 使用示例
isOrg, id := getQuotaSource(c)
if isOrg {
    // 使用组织配额
    quota, _ = model.GetOrganizationQuota(id)
} else {
    // 使用个人配额
    quota, _ = model.GetUserQuota(id, false)
}
```

### 8.3 渐进式上线策略

**阶段一：Beta测试（1-2周）**
- 仅对管理员账户开放组织功能
- 限制每个用户最多创建1个组织
- 限制每个组织最多5个成员
- 收集反馈，修复问题

**阶段二：部分开放（2-4周）**
- 对付费用户开放组织功能
- 放宽限制：每用户3个组织，每组织20个成员
- 监控性能和资源使用

**阶段三：全面开放**
- 对所有用户开放
- 根据用户等级设置组织限制
- 提供企业级增值服务（更多成员、更高配额等）

---

## 九、监控和日志

### 9.1 关键指标监控

```go
// common/metrics.go
type OrganizationMetrics struct {
    TotalOrganizations      int64
    ActiveOrganizations     int64
    TotalMembers            int64
    AvgMembersPerOrg        float64
    TotalOrgQuota           int64
    TotalOrgUsedQuota       int64
    OrgRequestCount         int64
    OrgErrorRate            float64
}

func CollectOrganizationMetrics() *OrganizationMetrics {
    metrics := &OrganizationMetrics{}

    // 统计组织数量
    DB.Model(&model.Organization{}).
        Where("status = ?", 1).
        Count(&metrics.TotalOrganizations)

    // 统计活跃组织（近30天有请求）
    DB.Model(&model.Organization{}).
        Where("status = ? AND request_count > 0", 1).
        Where("updated_at > ?", time.Now().AddDate(0, 0, -30)).
        Count(&metrics.ActiveOrganizations)

    // 统计总成员数
    DB.Model(&model.OrganizationMember{}).
        Where("status = ?", 1).
        Count(&metrics.TotalMembers)

    // 计算平均成员数
    if metrics.TotalOrganizations > 0 {
        metrics.AvgMembersPerOrg = float64(metrics.TotalMembers) / float64(metrics.TotalOrganizations)
    }

    // 统计配额
    DB.Model(&model.Organization{}).
        Select("SUM(quota) as total_quota, SUM(used_quota) as total_used_quota").
        Scan(&metrics)

    return metrics
}
```

### 9.2 审计日志

```go
// model/organization_audit_log.go
type OrganizationAuditLog struct {
    Id              int
    OrganizationId  int
    UserId          int
    Action          string  // create_token, invite_member, topup, etc.
    Resource        string  // token, member, quota
    ResourceId      int
    Details         string  // JSON details
    IpAddress       string
    UserAgent       string
    CreatedAt       time.Time
}

func RecordOrganizationAudit(orgId, userId int, action, resource string, resourceId int, details interface{}) {
    detailsJson, _ := json.Marshal(details)

    log := &OrganizationAuditLog{
        OrganizationId: orgId,
        UserId:         userId,
        Action:         action,
        Resource:       resource,
        ResourceId:     resourceId,
        Details:        string(detailsJson),
        IpAddress:      getClientIp(),
        UserAgent:      getUserAgent(),
    }

    DB.Create(log)
}

// 使用示例
RecordOrganizationAudit(
    orgId,
    userId,
    "create_token",
    "token",
    token.Id,
    map[string]interface{}{
        "token_name": token.Name,
        "quota":      token.RemainQuota,
    },
)
```

---

## 十、安全考虑

### 10.1 权限安全

1. **最小权限原则**: 新成员默认为`member`角色，权限最小
2. **角色继承**: `owner > admin > finance > member`
3. **敏感操作**: 删除组织、转让所有权需要二次确认
4. **IP白名单**: 组织级IP限制，防止未授权访问

### 10.2 配额安全

1. **预消费机制**: 防止恶意使用导致欠费
2. **成员配额限制**: 防止单个成员耗尽组织配额
3. **配额预警**: 配额低于阈值时通知管理员
4. **配额锁**: 使用数据库行级锁防止并发扣减错误

### 10.3 数据隔离

1. **组织数据隔离**: 日志、令牌、配额严格按组织过滤
2. **软删除**: 组织删除采用软删除，数据可恢复
3. **级联删除**: 组织删除时级联删除成员关系、邀请记录
4. **令牌保护**: 组织令牌仅组织成员可查看

---

## 十一、性能优化

### 11.1 缓存策略

```go
// 组织信息缓存
func GetOrganizationCache(orgId int) (*Organization, error) {
    cacheKey := fmt.Sprintf("org:%d", orgId)

    // 1. 尝试从Redis获取
    if common.RedisEnabled {
        cached, err := common.RedisGet(cacheKey)
        if err == nil && cached != "" {
            var org Organization
            json.Unmarshal([]byte(cached), &org)
            return &org, nil
        }
    }

    // 2. 从数据库获取
    org, err := GetOrganizationById(orgId)
    if err != nil {
        return nil, err
    }

    // 3. 异步更新缓存
    gopool.Go(func() {
        orgJson, _ := json.Marshal(org)
        common.RedisSet(cacheKey, string(orgJson), 300) // 5分钟
    })

    return org, nil
}

// 成员关系缓存
func GetOrganizationMemberCache(orgId, userId int) (*OrganizationMember, error) {
    cacheKey := fmt.Sprintf("org_member:%d:%d", orgId, userId)

    // 类似逻辑...
}
```

### 11.2 批量操作优化

```go
// 批量获取组织成员
func BatchGetOrganizationMembers(orgIds []int) (map[int][]*OrganizationMember, error) {
    var members []*OrganizationMember
    err := DB.Where("organization_id IN ?", orgIds).
        Where("status = ?", 1).
        Find(&members).Error

    // 按组织ID分组
    result := make(map[int][]*OrganizationMember)
    for _, member := range members {
        result[member.OrganizationId] = append(result[member.OrganizationId], member)
    }

    return result, err
}
```

### 11.3 索引优化

```sql
-- 组织成员查询优化
CREATE INDEX idx_org_user_status ON organization_members(organization_id, user_id, status);

-- 组织日志查询优化
CREATE INDEX idx_org_created ON logs(organization_id, created_at DESC);

-- 组织充值查询优化
CREATE INDEX idx_org_status_created ON topups(organization_id, status, create_time DESC);
```

---

## 十二、扩展规划

### 12.1 短期扩展（1-3个月）

1. **SSO集成**: 支持企业单点登录（SAML, OAuth2）
2. **审计日志导出**: 支持CSV/JSON格式导出
3. **Webhook通知**: 配额预警、成员变更等事件通知
4. **API密钥分组**: 为令牌添加标签和分组功能
5. **成本中心**: 按项目/部门分配配额和成本

### 12.2 中期扩展（3-6个月）

1. **组织层级**: 支持父子组织结构（集团 > 子公司 > 部门）
2. **配额池**: 多个组织共享配额池
3. **使用报表**: 可视化图表展示使用趋势
4. **配额转移**: 组织间转移配额
5. **自定义角色**: 支持自定义角色和权限

### 12.3 长期扩展（6-12个月）

1. **计费系统**: 按量付费、包月套餐
2. **发票管理**: 自动生成发票
3. **合规性**: GDPR、SOC2等合规支持
4. **多地域**: 数据本地化存储
5. **API配额市场**: 组织间交易配额

---

## 十三、测试计划

### 13.1 单元测试

```go
// model/organization_test.go
func TestCreateOrganization(t *testing.T) {
    org := &Organization{
        Name:        "test-org",
        DisplayName: "Test Organization",
        OwnerId:     1,
        Status:      1,
    }

    err := org.Insert()
    assert.Nil(t, err)
    assert.Greater(t, org.Id, 0)

    // 清理
    org.Delete()
}

func TestOrganizationMemberPermissions(t *testing.T) {
    member := &OrganizationMember{
        OrganizationId: 1,
        UserId:         2,
        Role:           "member",
        Permissions:    `{"tokens": {"create": true}}`,
    }

    perm := parsePermissions(member.Permissions)
    assert.True(t, perm.Can("tokens.create"))
    assert.False(t, perm.Can("members.invite"))
}
```

### 13.2 集成测试

```go
// controller/organization_test.go
func TestOrganizationAPI(t *testing.T) {
    // 1. 创建组织
    createReq := `{"name": "test", "displayName": "Test Org"}`
    resp := performRequest("POST", "/api/organizations", createReq, userToken)
    assert.Equal(t, 200, resp.Code)

    orgId := parseOrgIdFromResponse(resp.Body)

    // 2. 邀请成员
    inviteReq := `{"email": "user@test.com", "role": "member"}`
    resp = performRequest("POST", fmt.Sprintf("/api/organizations/%d/members/invite", orgId), inviteReq, userToken)
    assert.Equal(t, 200, resp.Code)

    // 3. 切换组织
    switchReq := fmt.Sprintf(`{"organizationId": %d}`, orgId)
    resp = performRequest("POST", "/api/user/switch-organization", switchReq, userToken)
    assert.Equal(t, 200, resp.Code)

    // 4. 创建组织令牌
    tokenReq := `{"name": "Test Token", "remainQuota": 100000}`
    resp = performRequest("POST", fmt.Sprintf("/api/organizations/%d/tokens", orgId), tokenReq, userToken)
    assert.Equal(t, 200, resp.Code)
}
```

### 13.3 性能测试

使用Apache Bench或k6进行压力测试：

```bash
# 测试组织列表接口
ab -n 10000 -c 100 -H "Authorization: Bearer TOKEN" \
   https://api.example.com/api/organizations

# 测试组织切换接口
k6 run --vus 100 --duration 30s organization_switch_test.js
```

---

## 十四、文档和培训

### 14.1 用户文档

**内容大纲**:
1. 什么是组织？
2. 如何创建组织
3. 邀请和管理成员
4. 组织令牌的使用
5. 组织充值和配额管理
6. 限流策略配置
7. 常见问题FAQ

### 14.2 开发者文档

**内容大纲**:
1. 组织API完整参考
2. 权限模型说明
3. Webhook事件列表
4. 代码示例（Go, Python, Node.js）
5. 最佳实践

### 14.3 管理员手册

**内容大纲**:
1. 组织功能系统架构
2. 数据库结构说明
3. 配置参数说明
4. 监控和告警
5. 故障排查指南

---

## 十五、总结

本方案为 New API 系统设计了完整的组织功能，核心特性包括：

### 15.1 核心亮点

1. **多租户隔离**: 组织间资源完全隔离，配额、令牌、日志独立管理
2. **灵活权限**: 4级角色（Owner/Admin/Finance/Member）+ 细粒度权限控制
3. **平滑迁移**: 兼容现有用户系统，个人账户与组织账户并存
4. **易扩展性**: 预留SSO、审计日志、成本中心等企业级功能接口

### 15.2 技术方案

- **数据库**: 9张表设计，完整的关系模型
- **API**: 30+ RESTful接口，覆盖所有组织管理场景
- **权限**: 中间件 + 细粒度权限检查，确保安全
- **性能**: Redis缓存 + 索引优化 + 批量操作
- **前端**: React组件化设计，Semi UI风格统一

### 15.3 实施建议

1. **分阶段上线**: Beta测试 → 部分开放 → 全面开放
2. **监控先行**: 关键指标监控 + 审计日志
3. **文档完善**: 用户文档 + 开发者文档 + 管理员手册
4. **测试覆盖**: 单元测试 + 集成测试 + 性能测试

### 15.4 下一步行动

1. 评审本方案，确定功能范围
2. 数据库迁移脚本编写和测试
3. 后端API开发（预计2-3周）
4. 前端UI开发（预计2周）
5. 集成测试和Beta测试（预计1周）
6. 文档编写和培训

---

**方案编写人**: Claude AI Assistant
**审核人**: [待填写]
**批准人**: [待填写]
**版本**: v1.0
**最后更新**: 2026-01-28