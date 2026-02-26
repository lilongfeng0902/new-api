# New API 数据库表结构文档

> 生成时间: 2026-01-16
> 数据库版本: v0.10.5
> 总表数: 19

## 目录

1. [abilities - 模型能力表](#1-abilities---模型能力表)
2. [channels - 渠道表](#2-channels---渠道表)
3. [checkins - 签到记录表](#3-checkins---签到记录表)
4. [logs - 日志表](#4-logs---日志表)
5. [midjourneys - Midjourney任务记录表](#5-midjourneys---midjourney任务记录表)
6. [models - 模型元数据表](#6-models---模型元数据表)
7. [options - 系统选项配置表](#7-options---系统选项配置表)
8. [passkey_credentials - Passkey凭证表](#8-passkey_credentials---passkey凭证表)
9. [prefill_groups - 预填充组表](#9-prefill_groups---预填充组表)
10. [quota_data - 额度数据统计表](#10-quota_data---额度数据统计表)
11. [redemptions - 兑换码表](#11-redemptions---兑换码表)
12. [setups - 系统初始化记录表](#12-setups---系统初始化记录表)
13. [tasks - 异步任务表](#13-tasks---异步任务表)
14. [tokens - API令牌表](#14-tokens---api令牌表)
15. [top_ups - 充值记录表](#15-top_ups---充值记录表)
16. [two_fa_backup_codes - 双因素认证备用码表](#16-two_fa_backup_codes---双因素认证备用码表)
17. [two_fas - 双因素认证设置表](#17-two_fas---双因素认证设置表)
18. [users - 用户表](#18-users---用户表)
19. [vendors - 供应商表](#19-vendors---供应商表)

---

## 1. abilities - 模型能力表

**描述**: 记录每个渠道支持的AI模型及其调度配置（优先级、权重等）
**用途**: 用于智能路由和负载均衡，实现模型请求的多渠道分发

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| group | varchar(64) | NO | - | 用户组名称（如default、vip等） |
| model | varchar(255) | NO | - | 模型名称（如gpt-4、claude-3-opus等） |
| channel_id | bigint | NO | - | 关联的渠道ID（外键关联channels表） |
| enabled | tinyint(1) | YES | NULL | 是否启用（1=启用，0=禁用） |
| priority | bigint | YES | 0 | 优先级（数值越大优先级越高，用于渠道选择排序） |
| weight | bigint UNSIGNED | YES | 0 | 权重（用于加权随机负载均衡，默认+10） |
| tag | varchar(191) | YES | NULL | 标签（用于批量管理，与channel.tag同步） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | (group, model, channel_id) | 复合主键 |
| idx_abilities_channel_id | INDEX | channel_id | 渠道ID索引 |
| idx_abilities_priority | INDEX | priority | 优先级索引 |
| idx_abilities_weight | INDEX | weight | 权重索引 |
| idx_abilities_tag | INDEX | tag | 标签索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 2. channels - 渠道表

**描述**: 存储 AI 服务提供商的访问配置（OpenAI、Claude、Gemini等30+提供商）
**用途**: 管理多个AI服务商的接入密钥、端点、状态等核心配置

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 渠道唯一标识 |
| type | bigint | YES | 0 | 渠道类型（1=OpenAI, 2=Claude, 3=Gemini等，见constant包定义） |
| key | longtext | NO | - | API密钥（支持多Key模式，换行分隔；或JSON数组格式） |
| open_ai_organization | longtext | YES | NULL | OpenAI组织ID（仅OpenAI渠道使用） |
| test_model | longtext | YES | NULL | 测试用模型名称（用于渠道健康检查） |
| status | bigint | YES | 1 | 渠道状态（1=启用, 2=手动禁用, 3=自动禁用） |
| name | varchar(191) | YES | NULL | 渠道名称（用于标识和显示） |
| weight | bigint UNSIGNED | YES | 0 | 权重（用于负载均衡，数值越大被选中概率越高） |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳，秒） |
| test_time | bigint | YES | NULL | 最后测试时间（Unix时间戳，秒） |
| response_time | bigint | YES | NULL | 响应时间（毫秒，用于性能监控） |
| base_url | varchar(191) | YES | '' | API基础URL（用于自定义端点或代理） |
| other | longtext | YES | NULL | 其他配置（JSON格式，存储渠道特有配置） |
| balance | double | YES | NULL | 账户余额（美元，用于显示API账户余额） |
| balance_updated_time | bigint | YES | NULL | 余额更新时间（Unix时间戳） |
| models | longtext | YES | NULL | 支持的模型列表（逗号分隔，如gpt-4,gpt-3.5-turbo） |
| group | varchar(64) | YES | 'default' | 所属用户组（逗号分隔，用于多租户隔离） |
| used_quota | bigint | YES | 0 | 已使用额度（累计消费统计） |
| model_mapping | text | YES | NULL | 模型映射配置（JSON，用于模型名称转换） |
| status_code_mapping | varchar(1024) | YES | '' | HTTP状态码映射配置（用于错误处理定制） |
| priority | bigint | YES | 0 | 优先级（数值越大优先级越高） |
| auto_ban | bigint | YES | 1 | 是否自动禁用（1=是，遇到错误自动禁用；0=否） |
| other_info | longtext | YES | NULL | 额外信息（JSON格式，包含状态原因status_reason和时间status_time） |
| tag | varchar(191) | YES | NULL | 标签（用于批量管理和分类） |
| setting | text | YES | NULL | 渠道额外设置（JSON格式） |
| param_override | text | YES | NULL | 参数覆盖配置（JSON，用于覆盖请求参数） |
| header_override | text | YES | NULL | HTTP头覆盖配置（JSON，用于自定义请求头） |
| remark | varchar(255) | YES | NULL | 备注说明 |
| channel_info | json | YES | NULL | 渠道信息（多Key模式配置：is_multi_key、multi_key_size、multi_key_status_list等） |
| settings | longtext | YES | NULL | 其他设置（如Azure版本等不需要检索的信息） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_channels_tag | INDEX | tag | 标签索引 |
| idx_channels_name | INDEX | name | 名称索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 3. checkins - 签到记录表

**描述**: 记录用户每日签到情况及获得的额度奖励
**用途**: 实现用户签到功能，提供每日随机额度奖励

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 签到记录ID |
| user_id | bigint | NO | - | 用户ID（外键关联users表） |
| checkin_date | varchar(10) | NO | - | 签到日期（格式：YYYY-MM-DD） |
| quota_awarded | bigint | NO | - | 签到奖励的额度（随机范围由配置决定） |
| created_at | bigint | YES | NULL | 创建时间（Unix时间戳，秒） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_user_checkin_date | UNIQUE INDEX | (user_id, checkin_date) | 保证每个用户每天只能签到一次 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 4. logs - 日志表

**描述**: 记录所有用户操作、API调用和系统事件
**用途**: 审计追踪、使用统计、问题排查
**日志类型**: 1=充值, 2=消费, 3=管理, 4=系统, 5=错误, 6=退款

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 日志ID |
| user_id | bigint | YES | NULL | 用户ID |
| created_at | bigint | YES | NULL | 创建时间（Unix时间戳，秒） |
| type | bigint | YES | NULL | 日志类型（1=充值, 2=消费, 3=管理, 4=系统, 5=错误, 6=退款） |
| content | longtext | YES | NULL | 日志内容（文本描述） |
| username | varchar(191) | YES | '' | 用户名（冗余字段，避免关联查询） |
| token_name | varchar(191) | YES | '' | 令牌名称 |
| model_name | varchar(191) | YES | '' | 模型名称（如gpt-4） |
| quota | bigint | YES | 0 | 消费额度 |
| prompt_tokens | bigint | YES | 0 | 提示词token数 |
| completion_tokens | bigint | YES | 0 | 生成内容token数 |
| use_time | bigint | YES | 0 | 请求耗时（秒） |
| is_stream | tinyint(1) | YES | NULL | 是否流式输出（1=是，0=否） |
| channel_id | bigint | YES | NULL | 渠道ID |
| channel_name | longtext | YES | NULL | 渠道名称（虚拟列，从channels表关联查询） |
| token_id | bigint | YES | 0 | 令牌ID |
| group | varchar(191) | YES | NULL | 用户组 |
| ip | varchar(191) | YES | '' | 客户端IP地址（可选，根据用户设置记录） |
| other | longtext | YES | NULL | 其他信息（JSON格式，存储扩展信息） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_created_at_id | INDEX | (id, created_at) | 创建时间和ID复合索引 |
| idx_created_at_type | INDEX | (created_at, type) | 创建时间和类型复合索引 |
| idx_logs_username | INDEX | username | 用户名索引 |
| idx_logs_model_name | INDEX | model_name | 模型名称索引 |
| idx_logs_channel_id | INDEX | channel_id | 渠道ID索引 |
| idx_logs_token_id | INDEX | token_id | 令牌ID索引 |
| idx_logs_ip | INDEX | ip | IP地址索引 |
| idx_logs_user_id | INDEX | user_id | 用户ID索引 |
| index_username_model_name | INDEX | (model_name, username) | 模型名称和用户名复合索引 |
| idx_logs_token_name | INDEX | token_name | 令牌名称索引 |
| idx_logs_group | INDEX | group | 用户组索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 5. midjourneys - Midjourney任务记录表

**描述**: 记录 Midjourney 图片生成任务的详细信息
**用途**: 管理和追踪Midjourney绘图任务，支持imagine、upscale、vary等操作

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 任务ID |
| code | bigint | YES | NULL | 状态码（HTTP响应状态码） |
| user_id | bigint | YES | NULL | 用户ID |
| action | varchar(40) | YES | NULL | 操作类型（imagine=生成, upscale=放大, vary=变体等） |
| mj_id | varchar(191) | YES | NULL | Midjourney任务ID（由MJ系统返回） |
| prompt | longtext | YES | NULL | 原始提示词（用户输入） |
| prompt_en | longtext | YES | NULL | 英文提示词（翻译后） |
| description | longtext | YES | NULL | 任务描述 |
| state | longtext | YES | NULL | 任务状态（详细状态信息） |
| submit_time | bigint | YES | NULL | 提交时间（Unix时间戳） |
| start_time | bigint | YES | NULL | 开始时间（Unix时间戳） |
| finish_time | bigint | YES | NULL | 完成时间（Unix时间戳） |
| image_url | longtext | YES | NULL | 生成的图片URL |
| video_url | longtext | YES | NULL | 生成的视频URL（单个） |
| video_urls | longtext | YES | NULL | 生成的视频URL列表（JSON数组） |
| status | varchar(20) | YES | NULL | 任务状态（SUCCESS=成功, FAILURE=失败等） |
| progress | varchar(30) | YES | NULL | 进度（如：100%，50%等） |
| fail_reason | longtext | YES | NULL | 失败原因（错误信息） |
| channel_id | bigint | YES | NULL | 渠道ID |
| quota | bigint | YES | NULL | 消费额度 |
| buttons | longtext | YES | NULL | 可用操作按钮（JSON数组，如U1-U4, V1-V4等） |
| properties | longtext | YES | NULL | 任务属性（JSON格式，存储额外参数） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_midjourneys_action | INDEX | action | 操作类型索引 |
| idx_midjourneys_mj_id | INDEX | mj_id | MJ任务ID索引 |
| idx_midjourneys_submit_time | INDEX | submit_time | 提交时间索引 |
| idx_midjourneys_start_time | INDEX | start_time | 开始时间索引 |
| idx_midjourneys_finish_time | INDEX | finish_time | 完成时间索引 |
| idx_midjourneys_status | INDEX | status | 状态索引 |
| idx_midjourneys_progress | INDEX | progress | 进度索引 |
| idx_midjourneys_user_id | INDEX | user_id | 用户ID索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 6. models - 模型元数据表

**描述**: 存储 AI 模型的基本信息和配置
**用途**: 管理所有支持的AI模型，提供模型目录和搜索功能

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 模型ID |
| model_name | varchar(128) | NO | - | 模型名称（如 gpt-4, claude-3-opus等，唯一） |
| description | text | YES | NULL | 模型描述（功能介绍、特点说明） |
| icon | varchar(128) | YES | NULL | 图标名称（用于前端显示，使用@lobehub/icons） |
| tags | varchar(255) | YES | NULL | 标签（逗号分隔，用于分类和筛选） |
| vendor_id | bigint | YES | NULL | 供应商ID（关联vendors表，标识模型提供商） |
| endpoints | text | YES | NULL | 支持的端点（JSON数组，如chat、embeddings、images等） |
| status | bigint | YES | 1 | 状态（1=启用, 0=禁用） |
| sync_official | bigint | YES | 1 | 是否同步官方数据（1=是，从官方源更新；0=否，用户自定义） |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| updated_time | bigint | YES | NULL | 更新时间（Unix时间戳） |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |
| name_rule | bigint | YES | 0 | 名称匹配规则（0=精确匹配, 1=前缀匹配, 2=包含匹配, 3=后缀匹配） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| uk_model_name_delete_at | UNIQUE INDEX | (model_name, deleted_at) | 模型名称和删除时间唯一索引 |
| idx_models_vendor_id | INDEX | vendor_id | 供应商ID索引 |
| idx_models_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 7. options - 系统选项配置表

**描述**: 存储系统全局配置（键值对）
**用途**: 保存各类系统设置，支持热加载
**常见配置**: 邮件服务器、支付配置、OAuth配置、额度配置等

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| key | varchar(191) | NO | - | 配置项名称（唯一键） |
| value | longtext | YES | NULL | 配置项值（支持JSON、纯文本等格式） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | key | 主键 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 8. passkey_credentials - Passkey凭证表

**描述**: 存储 WebAuthn/Passkey 认证凭证（FIDO2标准）
**用途**: 实现无密码登录，支持生物识别（指纹、面部识别）

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 凭证ID |
| user_id | bigint | NO | - | 用户ID（唯一，一个用户只能有一个Passkey） |
| credential_id | varchar(512) | NO | - | 凭证ID（Base64编码，由WebAuthn生成） |
| public_key | text | NO | - | 公钥（Base64编码，用于验证签名） |
| attestation_type | varchar(255) | YES | NULL | 认证类型（如none、packed、android-safetynet等） |
| aa_guid | varchar(512) | YES | NULL | 认证器GUID（AAGUID，Base64编码） |
| sign_count | int UNSIGNED | YES | 0 | 签名计数（用于检测克隆攻击，每次认证递增） |
| clone_warning | tinyint(1) | YES | NULL | 克隆警告标志（检测到可疑活动时标记） |
| user_present | tinyint(1) | YES | NULL | 用户在场标志（UP，用户物理存在） |
| user_verified | tinyint(1) | YES | NULL | 用户验证标志（UV，通过生物识别等验证） |
| backup_eligible | tinyint(1) | YES | NULL | 备份资格标志（BE，凭证是否可备份） |
| backup_state | tinyint(1) | YES | NULL | 备份状态标志（BS，凭证是否已备份） |
| transports | text | YES | NULL | 传输方式（JSON数组，如["usb","nfc","ble","internal"]） |
| attachment | varchar(32) | YES | NULL | 连接类型（platform=内置，cross-platform=外部设备） |
| last_used_at | datetime(3) | YES | NULL | 最后使用时间 |
| created_at | datetime(3) | YES | NULL | 创建时间 |
| updated_at | datetime(3) | YES | NULL | 更新时间 |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_passkey_credentials_user_id | UNIQUE INDEX | user_id | 用户ID唯一索引 |
| idx_passkey_credentials_credential_id | UNIQUE INDEX | credential_id | 凭证ID唯一索引 |
| idx_passkey_credentials_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 9. prefill_groups - 预填充组表

**描述**: 存储可复用的配置组（模型组、标签组、端点组等）
**用途**: 简化渠道配置，支持批量选择和模板化管理

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 组ID |
| name | varchar(64) | NO | - | 组名称（唯一，用于标识） |
| type | varchar(32) | NO | - | 组类型（model=模型组, tag=标签组, endpoint=端点组） |
| items | json | YES | NULL | 组项目（JSON数组，如：["gpt-4","gpt-3.5-turbo"]） |
| description | varchar(255) | YES | NULL | 组描述 |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| updated_time | bigint | YES | NULL | 更新时间（Unix时间戳） |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| uk_prefill_name | UNIQUE INDEX | name | 组名称唯一索引 |
| idx_prefill_groups_type | INDEX | type | 组类型索引 |
| idx_prefill_groups_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 10. quota_data - 额度数据统计表

**描述**: 按小时统计用户模型使用情况（用于数据看板）
**用途**: 提供使用量分析、成本统计、趋势图表等

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 记录ID |
| user_id | bigint | YES | NULL | 用户ID |
| username | varchar(64) | YES | '' | 用户名（冗余字段） |
| model_name | varchar(64) | YES | '' | 模型名称 |
| created_at | bigint | YES | NULL | 统计时间（按小时精确，秒级时间戳对3600取整） |
| token_used | bigint | YES | 0 | 使用的token数（prompt_tokens + completion_tokens） |
| count | bigint | YES | 0 | 请求次数 |
| quota | bigint | YES | 0 | 消费额度 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_qdt_model_user_name | INDEX | (model_name, username) | 模型名称和用户名复合索引 |
| idx_qdt_created_at | INDEX | created_at | 创建时间索引 |
| idx_quota_data_user_id | INDEX | user_id | 用户ID索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 11. redemptions - 兑换码表

**描述**: 存储充值兑换码信息
**用途**: 支持批量生成兑换码、用户充值、促销活动

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 兑换码ID |
| user_id | bigint | YES | NULL | 创建者用户ID |
| key | char(32) | YES | NULL | 兑换码（32位随机字符，唯一） |
| status | bigint | YES | 1 | 状态（1=可用, 2=已使用, 3=已禁用） |
| name | varchar(191) | YES | NULL | 兑换码名称/备注（用于标识用途） |
| quota | bigint | YES | 100 | 兑换额度 |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| redeemed_time | bigint | YES | NULL | 兑换时间（Unix时间戳） |
| used_user_id | bigint | YES | NULL | 使用者用户ID |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |
| expired_time | bigint | YES | NULL | 过期时间（Unix时间戳，0=永不过期） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_redemptions_key | UNIQUE INDEX | key | 兑换码唯一索引 |
| idx_redemptions_name | INDEX | name | 名称索引 |
| idx_redemptions_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 12. setups - 系统初始化记录表

**描述**: 记录系统初始化版本和时间
**用途**: 数据库版本管理，防止重复初始化

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint UNSIGNED | NO | AUTO_INCREMENT | 记录ID |
| version | varchar(50) | NO | - | 初始化版本号（如v0.1.0） |
| initialized_at | bigint | NO | - | 初始化时间（Unix时间戳） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 13. tasks - 异步任务表

**描述**: 记录音频、视频等异步生成任务
**用途**: 管理Suno音频、Kling/Luma视频等长时任务
**支持平台**: suno、kling、luma、runway等

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 任务ID |
| created_at | bigint | YES | NULL | 创建时间（Unix时间戳） |
| updated_at | bigint | YES | NULL | 更新时间（Unix时间戳） |
| task_id | varchar(191) | YES | NULL | 第三方任务ID（由服务提供商返回） |
| platform | varchar(30) | YES | NULL | 平台（suno, kling, luma, runway等） |
| user_id | bigint | YES | NULL | 用户ID |
| group | varchar(50) | YES | NULL | 用户组（用于修正计费） |
| channel_id | bigint | YES | NULL | 渠道ID |
| quota | bigint | YES | NULL | 消费额度 |
| action | varchar(40) | YES | NULL | 操作类型（song=音乐, lyrics=歌词, video=视频等） |
| status | varchar(20) | YES | NULL | 任务状态（QUEUED=排队, IN_PROGRESS=进行中, SUCCESS=成功, FAILURE=失败） |
| fail_reason | longtext | YES | NULL | 失败原因 |
| submit_time | bigint | YES | NULL | 提交时间（Unix时间戳） |
| start_time | bigint | YES | NULL | 开始时间（Unix时间戳） |
| finish_time | bigint | YES | NULL | 完成时间（Unix时间戳） |
| progress | varchar(20) | YES | NULL | 进度（如：50%，100%） |
| properties | json | YES | NULL | 任务属性（JSON格式，存储输入参数、模型名等） |
| private_data | json | YES | NULL | 私有数据（包含密钥等敏感信息，不返回给用户） |
| data | json | YES | NULL | 任务结果数据（JSON格式，存储输出结果） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_tasks_created_at | INDEX | created_at | 创建时间索引 |
| idx_tasks_channel_id | INDEX | channel_id | 渠道ID索引 |
| idx_tasks_action | INDEX | action | 操作类型索引 |
| idx_tasks_submit_time | INDEX | submit_time | 提交时间索引 |
| idx_tasks_progress | INDEX | progress | 进度索引 |
| idx_tasks_task_id | INDEX | task_id | 第三方任务ID索引 |
| idx_tasks_platform | INDEX | platform | 平台索引 |
| idx_tasks_user_id | INDEX | user_id | 用户ID索引 |
| idx_tasks_status | INDEX | status | 状态索引 |
| idx_tasks_start_time | INDEX | start_time | 开始时间索引 |
| idx_tasks_finish_time | INDEX | finish_time | 完成时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 14. tokens - API令牌表

**描述**: 存储用户生成的 API 访问令牌
**用途**: API密钥管理、额度控制、IP白名单、模型限制

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 令牌ID |
| user_id | bigint | YES | NULL | 用户ID（外键关联users表） |
| key | char(48) | YES | NULL | API密钥（sk-开头，48位，唯一） |
| status | bigint | YES | 1 | 状态（1=启用, 2=已耗尽, 3=已过期） |
| name | varchar(191) | YES | NULL | 令牌名称（用于标识用途） |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| accessed_time | bigint | YES | NULL | 最后访问时间（Unix时间戳） |
| expired_time | bigint | YES | -1 | 过期时间（Unix时间戳，-1=永不过期） |
| remain_quota | bigint | YES | 0 | 剩余额度 |
| unlimited_quota | tinyint(1) | YES | NULL | 是否无限额度（1=是，不受额度限制） |
| model_limits_enabled | tinyint(1) | YES | NULL | 是否启用模型限制（1=是，只能使用指定模型） |
| model_limits | varchar(1024) | YES | '' | 模型限制列表（逗号分隔） |
| allow_ips | varchar(191) | YES | '' | 允许的IP地址（换行分隔，支持IP白名单） |
| used_quota | bigint | YES | 0 | 已使用额度 |
| group | varchar(191) | YES | '' | 指定用户组（覆盖用户默认组，空=使用用户组） |
| cross_group_retry | tinyint(1) | YES | NULL | 是否允许跨组重试（仅auto组有效，1=是） |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_tokens_key | UNIQUE INDEX | key | API密钥唯一索引 |
| idx_tokens_deleted_at | INDEX | deleted_at | 软删除时间索引 |
| idx_tokens_user_id | INDEX | user_id | 用户ID索引 |
| idx_tokens_name | INDEX | name | 令牌名称索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 15. top_ups - 充值记录表

**描述**: 记录用户在线充值订单
**用途**: 支付管理、订单追踪、对账
**支持支付方式**: stripe（Stripe）、epay（易支付）、creem等

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 充值记录ID |
| user_id | bigint | YES | NULL | 用户ID |
| amount | bigint | YES | NULL | 充值金额（美元，整数，用于易支付等） |
| money | double | YES | NULL | 充值金额（美元，浮点数，用于Stripe等，经分组倍率换算） |
| trade_no | varchar(255) | YES | NULL | 订单号（唯一，用于支付回调） |
| payment_method | varchar(50) | YES | NULL | 支付方式（stripe, epay, creem等） |
| create_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| complete_time | bigint | YES | NULL | 完成时间（Unix时间戳） |
| status | longtext | YES | NULL | 订单状态（pending=待支付, success=成功, failed=失败） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| trade_no | UNIQUE INDEX | trade_no | 订单号唯一索引 |
| idx_top_ups_user_id | INDEX | user_id | 用户ID索引 |
| idx_top_ups_trade_no | INDEX | trade_no | 订单号索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 16. two_fa_backup_codes - 双因素认证备用码表

**描述**: 存储用户的 2FA 备用恢复码
**用途**: 当用户丢失2FA设备时，可使用备用码登录

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 备用码ID |
| user_id | bigint | NO | - | 用户ID |
| code_hash | varchar(255) | NO | - | 备用码哈希值（使用bcrypt加密存储） |
| is_used | tinyint(1) | YES | NULL | 是否已使用（1=已使用，每个备用码只能使用一次） |
| used_at | datetime(3) | YES | NULL | 使用时间 |
| created_at | datetime(3) | YES | NULL | 创建时间 |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| idx_two_fa_backup_codes_user_id | INDEX | user_id | 用户ID索引 |
| idx_two_fa_backup_codes_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 17. two_fas - 双因素认证设置表

**描述**: 存储用户的 2FA (TOTP) 配置
**用途**: 提供额外安全层，支持Google Authenticator等APP

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 2FA设置ID |
| user_id | bigint | NO | - | 用户ID（唯一，一个用户只能有一个2FA设置） |
| secret | varchar(255) | NO | - | TOTP密钥（Base32编码，用于生成验证码，不返回给前端） |
| is_enabled | tinyint(1) | YES | NULL | 是否启用2FA（1=启用） |
| failed_attempts | bigint | YES | 0 | 失败尝试次数（连续失败会触发锁定） |
| locked_until | datetime(3) | YES | NULL | 锁定截止时间（超过此时间自动解锁） |
| last_used_at | datetime(3) | YES | NULL | 最后使用时间 |
| created_at | datetime(3) | YES | NULL | 创建时间 |
| updated_at | datetime(3) | YES | NULL | 更新时间 |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| user_id | UNIQUE INDEX | user_id | 用户ID唯一索引 |
| idx_two_fas_user_id | INDEX | user_id | 用户ID索引 |
| idx_two_fas_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 18. users - 用户表

**描述**: 存储用户基本信息、额度和OAuth绑定
**用途**: 用户管理、认证授权、额度控制
**支持OAuth**: GitHub、Discord、WeChat、Telegram、OIDC、LinuxDO

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 用户ID |
| username | varchar(191) | YES | NULL | 用户名（唯一，用于登录） |
| password | longtext | NO | - | 密码哈希（使用bcrypt加密） |
| display_name | varchar(191) | YES | NULL | 显示名称（昵称） |
| role | bigint | YES | 1 | 角色（1=普通用户, 10=管理员, 100=超级管理员/Root） |
| status | bigint | YES | 1 | 状态（1=启用, 2=禁用） |
| email | varchar(191) | YES | NULL | 邮箱（用于找回密码、接收通知） |
| github_id | varchar(191) | YES | NULL | GitHub OAuth ID（绑定GitHub账号） |
| discord_id | varchar(191) | YES | NULL | Discord OAuth ID（绑定Discord账号） |
| oidc_id | varchar(191) | YES | NULL | OIDC OAuth ID（OpenID Connect） |
| wechat_id | varchar(191) | YES | NULL | 微信 OAuth ID（绑定微信账号） |
| telegram_id | varchar(191) | YES | NULL | Telegram OAuth ID（绑定Telegram账号） |
| access_token | char(32) | YES | NULL | 系统管理访问令牌（32位随机字符，用于API管理操作） |
| quota | bigint | YES | 0 | 剩余额度（可用余额） |
| used_quota | bigint | YES | 0 | 已使用额度（累计消费） |
| request_count | bigint | YES | 0 | 请求次数统计（累计API调用次数） |
| group | varchar(64) | YES | 'default' | 用户组（用于权限控制和定价策略） |
| aff_code | varchar(32) | YES | NULL | 邀请码（唯一，4位随机字符） |
| aff_count | bigint | YES | 0 | 邀请人数（成功邀请的用户数） |
| aff_quota | bigint | YES | 0 | 邀请剩余额度（可转移到主额度） |
| aff_history | bigint | YES | 0 | 邀请历史额度（累计获得的邀请奖励） |
| inviter_id | bigint | YES | NULL | 邀请人ID（谁邀请了这个用户） |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |
| linux_do_id | varchar(191) | YES | NULL | LinuxDO OAuth ID（绑定LinuxDO账号） |
| setting | text | YES | NULL | 用户设置（JSON格式，包含界面配置、隐私设置等） |
| remark | varchar(255) | YES | NULL | 备注说明（管理员使用） |
| stripe_customer | varchar(64) | YES | NULL | Stripe客户ID（用于Stripe支付） |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| username | UNIQUE INDEX | username | 用户名唯一索引 |
| idx_users_access_token | UNIQUE INDEX | access_token | 访问令牌唯一索引 |
| idx_users_aff_code | UNIQUE INDEX | aff_code | 邀请码唯一索引 |
| idx_users_deleted_at | INDEX | deleted_at | 软删除时间索引 |
| idx_users_linux_do_id | INDEX | linux_do_id | LinuxDO ID索引 |
| idx_users_stripe_customer | INDEX | stripe_customer | Stripe客户ID索引 |
| idx_users_username | INDEX | username | 用户名索引 |
| idx_users_display_name | INDEX | display_name | 显示名称索引 |
| idx_users_git_hub_id | INDEX | github_id | GitHub ID索引 |
| idx_users_discord_id | INDEX | discord_id | Discord ID索引 |
| idx_users_oidc_id | INDEX | oidc_id | OIDC ID索引 |
| idx_users_inviter_id | INDEX | inviter_id | 邀请人ID索引 |
| idx_users_email | INDEX | email | 邮箱索引 |
| idx_users_we_chat_id | INDEX | wechat_id | 微信ID索引 |
| idx_users_telegram_id | INDEX | telegram_id | Telegram ID索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 19. vendors - 供应商表

**描述**: 存储 AI 服务供应商信息（OpenAI、Anthropic等）
**用途**: 模型分类、供应商管理、图标显示

### 列定义

| 字段名 | 类型 | 允许NULL | 默认值 | 说明 |
|--------|------|----------|--------|------|
| id | bigint | NO | AUTO_INCREMENT | 供应商ID |
| name | varchar(128) | NO | - | 供应商名称（唯一，如OpenAI、Anthropic、Google等） |
| description | text | YES | NULL | 供应商描述 |
| icon | varchar(128) | YES | NULL | 图标名称（使用@lobehub/icons图标库） |
| status | bigint | YES | 1 | 状态（1=启用, 0=禁用） |
| created_time | bigint | YES | NULL | 创建时间（Unix时间戳） |
| updated_time | bigint | YES | NULL | 更新时间（Unix时间戳） |
| deleted_at | datetime(3) | YES | NULL | 软删除时间 |

### 索引

| 索引名 | 类型 | 列 | 说明 |
|--------|------|-----|------|
| PRIMARY | PRIMARY KEY | id | 主键 |
| uk_vendor_name_delete_at | UNIQUE INDEX | (name, deleted_at) | 供应商名称和删除时间唯一索引 |
| idx_vendors_deleted_at | INDEX | deleted_at | 软删除时间索引 |

**引擎**: InnoDB
**字符集**: utf8mb4
**排序规则**: utf8mb4_0900_ai_ci

---

## 数据库表分类汇总

### 核心业务表 (5)
- **users** - 用户管理和认证
- **tokens** - API密钥管理
- **channels** - AI服务商接入配置
- **abilities** - 模型-渠道-用户组映射
- **logs** - 审计和使用记录

### 任务和资源表 (3)
- **tasks** - 异步任务（音频、视频）
- **midjourneys** - Midjourney绘图任务
- **quota_data** - 使用量统计

### 支付和充值表 (2)
- **top_ups** - 在线充值订单
- **redemptions** - 兑换码

### 元数据和配置表 (4)
- **models** - 模型目录
- **vendors** - 供应商目录
- **prefill_groups** - 配置模板
- **options** - 系统配置

### 安全和认证表 (3)
- **two_fas** - 双因素认证
- **two_fa_backup_codes** - 2FA备用码
- **passkey_credentials** - Passkey无密码登录

### 其他表 (2)
- **checkins** - 签到奖励
- **setups** - 初始化记录

---

## 表关系说明

### 核心关系链
```
users (用户)
  ├─→ tokens (用户的API令牌)
  ├─→ logs (用户的操作日志)
  ├─→ quota_data (用户的额度统计)
  ├─→ tasks (用户的异步任务)
  ├─→ midjourneys (用户的MJ任务)
  ├─→ top_ups (用户的充值记录)
  ├─→ checkins (用户的签到记录)
  ├─→ two_fas (用户的2FA设置)
  └─→ passkey_credentials (用户的Passkey凭证)

channels (渠道)
  ├─→ abilities (渠道支持的模型能力)
  ├─→ logs (渠道的使用日志)
  ├─→ tasks (渠道执行的任务)
  └─→ midjourneys (渠道执行的MJ任务)

models (模型元数据)
  ├─→ abilities (模型与渠道的映射)
  └─→ vendors (模型所属的供应商)
```

### 索引优化说明
- **主键索引**: 所有表均有自增主键
- **唯一索引**: 保证关键字段唯一性（用户名、API Key、兑换码等）
- **外键索引**: 优化关联查询（user_id、channel_id等）
- **业务索引**: 加速常见查询（status、created_at、model_name等）
- **复合索引**: 优化多条件查询

### 软删除机制
以下表支持软删除（deleted_at字段）:
- users, tokens, models, vendors, prefill_groups, redemptions
- two_fas, two_fa_backup_codes, passkey_credentials

软删除的记录:
- 不会被普通查询检索
- 可通过GORM的Unscoped()查询
- 支持数据恢复和审计追踪

---

**文档版本**: 1.0
**生成日期**: 2026-01-16
**数据库引擎**: MySQL 5.7.8+ / PostgreSQL / SQLite
**字符集**: UTF-8 (utf8mb4)
