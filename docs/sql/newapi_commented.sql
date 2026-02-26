-- ======================================================================
-- New API 数据库表结构（完整注释版本）
-- 这是一个 AI Gateway & API 管理平台的数据库架构
-- 支持 30+ AI 服务提供商的统一接口管理
-- 生成时间: 2026-01-16
-- ======================================================================

-- ======================================
-- 1. abilities 表 - 模型能力表
-- 描述：记录每个渠道支持的AI模型及其调度配置（优先级、权重等）
-- 用途：用于智能路由和负载均衡，实现模型请求的多渠道分发
-- ======================================
CREATE TABLE `abilities`  (
  `group` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户组名称（如default、vip等）',
  `model` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '模型名称（如gpt-4、claude-3-opus等）',
  `channel_id` bigint(0) NOT NULL COMMENT '关联的渠道ID（外键关联channels表）',
  `enabled` tinyint(1) NULL DEFAULT NULL COMMENT '是否启用（1=启用，0=禁用）',
  `priority` bigint(0) NULL DEFAULT 0 COMMENT '优先级（数值越大优先级越高，用于渠道选择排序）',
  `weight` bigint(0) UNSIGNED NULL DEFAULT 0 COMMENT '权重（用于加权随机负载均衡，默认+10）',
  `tag` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '标签（用于批量管理，与channel.tag同步）',
  PRIMARY KEY (`group`, `model`, `channel_id`) USING BTREE,
  INDEX `idx_abilities_channel_id`(`channel_id`) USING BTREE,
  INDEX `idx_abilities_priority`(`priority`) USING BTREE,
  INDEX `idx_abilities_weight`(`weight`) USING BTREE,
  INDEX `idx_abilities_tag`(`tag`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='模型能力表，记录每个渠道支持的AI模型及其调度配置' ROW_FORMAT = Dynamic;

-- ======================================
-- 2. channels 表 - 渠道表
-- 描述：存储 AI 服务提供商的访问配置（OpenAI、Claude、Gemini等30+提供商）
-- 用途：管理多个AI服务商的接入密钥、端点、状态等核心配置
-- ======================================
CREATE TABLE `channels`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '渠道唯一标识',
  `type` bigint(0) NULL DEFAULT 0 COMMENT '渠道类型（1=OpenAI, 2=Claude, 3=Gemini等，见constant包定义）',
  `key` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'API密钥（支持多Key模式，换行分隔；或JSON数组格式）',
  `open_ai_organization` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT 'OpenAI组织ID（仅OpenAI渠道使用）',
  `test_model` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '测试用模型名称（用于渠道健康检查）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '渠道状态（1=启用, 2=手动禁用, 3=自动禁用）',
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '渠道名称（用于标识和显示）',
  `weight` bigint(0) UNSIGNED NULL DEFAULT 0 COMMENT '权重（用于负载均衡，数值越大被选中概率越高）',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳，秒）',
  `test_time` bigint(0) NULL DEFAULT NULL COMMENT '最后测试时间（Unix时间戳，秒）',
  `response_time` bigint(0) NULL DEFAULT NULL COMMENT '响应时间（毫秒，用于性能监控）',
  `base_url` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'API基础URL（用于自定义端点或代理）',
  `other` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '其他配置（JSON格式，存储渠道特有配置）',
  `balance` double NULL DEFAULT NULL COMMENT '账户余额（美元，用于显示API账户余额）',
  `balance_updated_time` bigint(0) NULL DEFAULT NULL COMMENT '余额更新时间（Unix时间戳）',
  `models` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '支持的模型列表（逗号分隔，如gpt-4,gpt-3.5-turbo）',
  `group` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'default' COMMENT '所属用户组（逗号分隔，用于多租户隔离）',
  `used_quota` bigint(0) NULL DEFAULT 0 COMMENT '已使用额度（累计消费统计）',
  `model_mapping` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '模型映射配置（JSON，用于模型名称转换）',
  `status_code_mapping` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT 'HTTP状态码映射配置（用于错误处理定制）',
  `priority` bigint(0) NULL DEFAULT 0 COMMENT '优先级（数值越大优先级越高）',
  `auto_ban` bigint(0) NULL DEFAULT 1 COMMENT '是否自动禁用（1=是，遇到错误自动禁用；0=否）',
  `other_info` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '额外信息（JSON格式，包含状态原因status_reason和时间status_time）',
  `tag` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '标签（用于批量管理和分类）',
  `setting` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '渠道额外设置（JSON格式）',
  `param_override` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '参数覆盖配置（JSON，用于覆盖请求参数）',
  `header_override` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT 'HTTP头覆盖配置（JSON，用于自定义请求头）',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注说明',
  `channel_info` json NULL COMMENT '渠道信息（多Key模式配置：is_multi_key、multi_key_size、multi_key_status_list等）',
  `settings` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '其他设置（如Azure版本等不需要检索的信息）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_channels_tag`(`tag`) USING BTREE,
  INDEX `idx_channels_name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='渠道表，存储各AI服务提供商的接入配置' ROW_FORMAT = Dynamic;

-- ======================================
-- 3. checkins 表 - 签到记录表
-- 描述：记录用户每日签到情况及获得的额度奖励
-- 用途：实现用户签到功能，提供每日随机额度奖励
-- ======================================
CREATE TABLE `checkins`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '签到记录ID',
  `user_id` bigint(0) NOT NULL COMMENT '用户ID（外键关联users表）',
  `checkin_date` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '签到日期（格式：YYYY-MM-DD）',
  `quota_awarded` bigint(0) NOT NULL COMMENT '签到奖励的额度（随机范围由配置决定）',
  `created_at` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳，秒）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_user_checkin_date`(`user_id`, `checkin_date`) USING BTREE COMMENT '保证每个用户每天只能签到一次'
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='签到记录表' ROW_FORMAT = Dynamic;

-- ======================================
-- 4. logs 表 - 日志表
-- 描述：记录所有用户操作、API调用和系统事件
-- 用途：审计追踪、使用统计、问题排查
-- 日志类型：1=充值, 2=消费, 3=管理, 4=系统, 5=错误, 6=退款
-- ======================================
CREATE TABLE `logs`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID',
  `created_at` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳，秒）',
  `type` bigint(0) NULL DEFAULT NULL COMMENT '日志类型（1=充值, 2=消费, 3=管理, 4=系统, 5=错误, 6=退款）',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '日志内容（文本描述）',
  `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '用户名（冗余字段，避免关联查询）',
  `token_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '令牌名称',
  `model_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '模型名称（如gpt-4）',
  `quota` bigint(0) NULL DEFAULT 0 COMMENT '消费额度',
  `prompt_tokens` bigint(0) NULL DEFAULT 0 COMMENT '提示词token数',
  `completion_tokens` bigint(0) NULL DEFAULT 0 COMMENT '生成内容token数',
  `use_time` bigint(0) NULL DEFAULT 0 COMMENT '请求耗时（秒）',
  `is_stream` tinyint(1) NULL DEFAULT NULL COMMENT '是否流式输出（1=是，0=否）',
  `channel_id` bigint(0) NULL DEFAULT NULL COMMENT '渠道ID',
  `channel_name` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '渠道名称（虚拟列，从channels表关联查询）',
  `token_id` bigint(0) NULL DEFAULT 0 COMMENT '令牌ID',
  `group` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户组',
  `ip` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '客户端IP地址（可选，根据用户设置记录）',
  `other` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '其他信息（JSON格式，存储扩展信息）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_created_at_id`(`id`, `created_at`) USING BTREE,
  INDEX `idx_created_at_type`(`created_at`, `type`) USING BTREE,
  INDEX `idx_logs_username`(`username`) USING BTREE,
  INDEX `idx_logs_model_name`(`model_name`) USING BTREE,
  INDEX `idx_logs_channel_id`(`channel_id`) USING BTREE,
  INDEX `idx_logs_token_id`(`token_id`) USING BTREE,
  INDEX `idx_logs_ip`(`ip`) USING BTREE,
  INDEX `idx_logs_user_id`(`user_id`) USING BTREE,
  INDEX `index_username_model_name`(`model_name`, `username`) USING BTREE,
  INDEX `idx_logs_token_name`(`token_name`) USING BTREE,
  INDEX `idx_logs_group`(`group`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='日志表，记录所有用户操作和API调用' ROW_FORMAT = Dynamic;

-- ======================================
-- 5. midjourneys 表 - Midjourney任务记录表
-- 描述：记录 Midjourney 图片生成任务的详细信息
-- 用途：管理和追踪Midjourney绘图任务，支持imagine、upscale、vary等操作
-- ======================================
CREATE TABLE `midjourneys`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `code` bigint(0) NULL DEFAULT NULL COMMENT '状态码（HTTP响应状态码）',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID',
  `action` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '操作类型（imagine=生成, upscale=放大, vary=变体等）',
  `mj_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Midjourney任务ID（由MJ系统返回）',
  `prompt` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '原始提示词（用户输入）',
  `prompt_en` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '英文提示词（翻译后）',
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '任务描述',
  `state` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '任务状态（详细状态信息）',
  `submit_time` bigint(0) NULL DEFAULT NULL COMMENT '提交时间（Unix时间戳）',
  `start_time` bigint(0) NULL DEFAULT NULL COMMENT '开始时间（Unix时间戳）',
  `finish_time` bigint(0) NULL DEFAULT NULL COMMENT '完成时间（Unix时间戳）',
  `image_url` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '生成的图片URL',
  `video_url` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '生成的视频URL（单个）',
  `video_urls` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '生成的视频URL列表（JSON数组）',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '任务状态（SUCCESS=成功, FAILURE=失败等）',
  `progress` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '进度（如：100%，50%等）',
  `fail_reason` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '失败原因（错误信息）',
  `channel_id` bigint(0) NULL DEFAULT NULL COMMENT '渠道ID',
  `quota` bigint(0) NULL DEFAULT NULL COMMENT '消费额度',
  `buttons` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '可用操作按钮（JSON数组，如U1-U4, V1-V4等）',
  `properties` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '任务属性（JSON格式，存储额外参数）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_midjourneys_action`(`action`) USING BTREE,
  INDEX `idx_midjourneys_mj_id`(`mj_id`) USING BTREE,
  INDEX `idx_midjourneys_submit_time`(`submit_time`) USING BTREE,
  INDEX `idx_midjourneys_start_time`(`start_time`) USING BTREE,
  INDEX `idx_midjourneys_finish_time`(`finish_time`) USING BTREE,
  INDEX `idx_midjourneys_status`(`status`) USING BTREE,
  INDEX `idx_midjourneys_progress`(`progress`) USING BTREE,
  INDEX `idx_midjourneys_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='Midjourney图片生成任务记录表' ROW_FORMAT = Dynamic;

-- ======================================
-- 6. models 表 - 模型元数据表
-- 描述：存储 AI 模型的基本信息和配置
-- 用途：管理所有支持的AI模型，提供模型目录和搜索功能
-- ======================================
CREATE TABLE `models`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '模型ID',
  `model_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '模型名称（如 gpt-4, claude-3-opus等，唯一）',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '模型描述（功能介绍、特点说明）',
  `icon` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '图标名称（用于前端显示，使用@lobehub/icons）',
  `tags` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '标签（逗号分隔，用于分类和筛选）',
  `vendor_id` bigint(0) NULL DEFAULT NULL COMMENT '供应商ID（关联vendors表，标识模型提供商）',
  `endpoints` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '支持的端点（JSON数组，如chat、embeddings、images等）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '状态（1=启用, 0=禁用）',
  `sync_official` bigint(0) NULL DEFAULT 1 COMMENT '是否同步官方数据（1=是，从官方源更新；0=否，用户自定义）',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `updated_time` bigint(0) NULL DEFAULT NULL COMMENT '更新时间（Unix时间戳）',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  `name_rule` bigint(0) NULL DEFAULT 0 COMMENT '名称匹配规则（0=精确匹配, 1=前缀匹配, 2=包含匹配, 3=后缀匹配）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_model_name_delete_at`(`model_name`, `deleted_at`) USING BTREE,
  INDEX `idx_models_vendor_id`(`vendor_id`) USING BTREE,
  INDEX `idx_models_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='AI模型元数据表' ROW_FORMAT = Dynamic;

-- ======================================
-- 7. options 表 - 系统选项配置表
-- 描述：存储系统全局配置（键值对）
-- 用途：保存各类系统设置，支持热加载
-- 常见配置：邮件服务器、支付配置、OAuth配置、额度配置等
-- ======================================
CREATE TABLE `options`  (
  `key` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '配置项名称（唯一键）',
  `value` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '配置项值（支持JSON、纯文本等格式）',
  PRIMARY KEY (`key`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='系统配置选项表（键值对存储）' ROW_FORMAT = Dynamic;

-- ======================================
-- 8. passkey_credentials 表 - Passkey凭证表
-- 描述：存储 WebAuthn/Passkey 认证凭证（FIDO2标准）
-- 用途：实现无密码登录，支持生物识别（指纹、面部识别）
-- ======================================
CREATE TABLE `passkey_credentials`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '凭证ID',
  `user_id` bigint(0) NOT NULL COMMENT '用户ID（唯一，一个用户只能有一个Passkey）',
  `credential_id` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '凭证ID（Base64编码，由WebAuthn生成）',
  `public_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '公钥（Base64编码，用于验证签名）',
  `attestation_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '认证类型（如none、packed、android-safetynet等）',
  `aa_guid` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '认证器GUID（AAGUID，Base64编码）',
  `sign_count` int(0) UNSIGNED NULL DEFAULT 0 COMMENT '签名计数（用于检测克隆攻击，每次认证递增）',
  `clone_warning` tinyint(1) NULL DEFAULT NULL COMMENT '克隆警告标志（检测到可疑活动时标记）',
  `user_present` tinyint(1) NULL DEFAULT NULL COMMENT '用户在场标志（UP，用户物理存在）',
  `user_verified` tinyint(1) NULL DEFAULT NULL COMMENT '用户验证标志（UV，通过生物识别等验证）',
  `backup_eligible` tinyint(1) NULL DEFAULT NULL COMMENT '备份资格标志（BE，凭证是否可备份）',
  `backup_state` tinyint(1) NULL DEFAULT NULL COMMENT '备份状态标志（BS，凭证是否已备份）',
  `transports` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '传输方式（JSON数组，如["usb","nfc","ble","internal"]）',
  `attachment` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '连接类型（platform=内置，cross-platform=外部设备）',
  `last_used_at` datetime(3) NULL DEFAULT NULL COMMENT '最后使用时间',
  `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_passkey_credentials_user_id`(`user_id`) USING BTREE,
  UNIQUE INDEX `idx_passkey_credentials_credential_id`(`credential_id`) USING BTREE,
  INDEX `idx_passkey_credentials_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='Passkey认证凭证表（WebAuthn/FIDO2）' ROW_FORMAT = Dynamic;

-- ======================================
-- 9. prefill_groups 表 - 预填充组表
-- 描述：存储可复用的配置组（模型组、标签组、端点组等）
-- 用途：简化渠道配置，支持批量选择和模板化管理
-- ======================================
CREATE TABLE `prefill_groups`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '组ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '组名称（唯一，用于标识）',
  `type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '组类型（model=模型组, tag=标签组, endpoint=端点组）',
  `items` json NULL COMMENT '组项目（JSON数组，如：["gpt-4","gpt-3.5-turbo"]）',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '组描述',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `updated_time` bigint(0) NULL DEFAULT NULL COMMENT '更新时间（Unix时间戳）',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_prefill_name`(`name`) USING BTREE,
  INDEX `idx_prefill_groups_type`(`type`) USING BTREE,
  INDEX `idx_prefill_groups_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='预填充配置组表' ROW_FORMAT = Dynamic;

-- ======================================
-- 10. quota_data 表 - 额度数据统计表
-- 描述：按小时统计用户模型使用情况（用于数据看板）
-- 用途：提供使用量分析、成本统计、趋势图表等
-- ======================================
CREATE TABLE `quota_data`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID',
  `username` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '用户名（冗余字段）',
  `model_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '模型名称',
  `created_at` bigint(0) NULL DEFAULT NULL COMMENT '统计时间（按小时精确，秒级时间戳对3600取整）',
  `token_used` bigint(0) NULL DEFAULT 0 COMMENT '使用的token数（prompt_tokens + completion_tokens）',
  `count` bigint(0) NULL DEFAULT 0 COMMENT '请求次数',
  `quota` bigint(0) NULL DEFAULT 0 COMMENT '消费额度',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_qdt_model_user_name`(`model_name`, `username`) USING BTREE,
  INDEX `idx_qdt_created_at`(`created_at`) USING BTREE,
  INDEX `idx_quota_data_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='额度统计数据表（按小时聚合）' ROW_FORMAT = Dynamic;

-- ======================================
-- 11. redemptions 表 - 兑换码表
-- 描述：存储充值兑换码信息
-- 用途：支持批量生成兑换码、用户充值、促销活动
-- ======================================
CREATE TABLE `redemptions`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '兑换码ID',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '创建者用户ID',
  `key` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '兑换码（32位随机字符，唯一）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '状态（1=可用, 2=已使用, 3=已禁用）',
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '兑换码名称/备注（用于标识用途）',
  `quota` bigint(0) NULL DEFAULT 100 COMMENT '兑换额度',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `redeemed_time` bigint(0) NULL DEFAULT NULL COMMENT '兑换时间（Unix时间戳）',
  `used_user_id` bigint(0) NULL DEFAULT NULL COMMENT '使用者用户ID',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  `expired_time` bigint(0) NULL DEFAULT NULL COMMENT '过期时间（Unix时间戳，0=永不过期）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_redemptions_key`(`key`) USING BTREE,
  INDEX `idx_redemptions_name`(`name`) USING BTREE,
  INDEX `idx_redemptions_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='兑换码表' ROW_FORMAT = Dynamic;

-- ======================================
-- 12. setups 表 - 系统初始化记录表
-- 描述：记录系统初始化版本和时间
-- 用途：数据库版本管理，防止重复初始化
-- ======================================
CREATE TABLE `setups`  (
  `id` bigint(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `version` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '初始化版本号（如v0.1.0）',
  `initialized_at` bigint(0) NOT NULL COMMENT '初始化时间（Unix时间戳）',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='系统初始化记录表' ROW_FORMAT = Dynamic;

-- ======================================
-- 13. tasks 表 - 异步任务表
-- 描述：记录音频、视频等异步生成任务
-- 用途：管理Suno音频、Kling/Luma视频等长时任务
-- 支持平台：suno、kling、luma、runway等
-- ======================================
CREATE TABLE `tasks`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `created_at` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `updated_at` bigint(0) NULL DEFAULT NULL COMMENT '更新时间（Unix时间戳）',
  `task_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '第三方任务ID（由服务提供商返回）',
  `platform` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '平台（suno, kling, luma, runway等）',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID',
  `group` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户组（用于修正计费）',
  `channel_id` bigint(0) NULL DEFAULT NULL COMMENT '渠道ID',
  `quota` bigint(0) NULL DEFAULT NULL COMMENT '消费额度',
  `action` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '操作类型（song=音乐, lyrics=歌词, video=视频等）',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '任务状态（QUEUED=排队, IN_PROGRESS=进行中, SUCCESS=成功, FAILURE=失败）',
  `fail_reason` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '失败原因',
  `submit_time` bigint(0) NULL DEFAULT NULL COMMENT '提交时间（Unix时间戳）',
  `start_time` bigint(0) NULL DEFAULT NULL COMMENT '开始时间（Unix时间戳）',
  `finish_time` bigint(0) NULL DEFAULT NULL COMMENT '完成时间（Unix时间戳）',
  `progress` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '进度（如：50%，100%）',
  `properties` json NULL COMMENT '任务属性（JSON格式，存储输入参数、模型名等）',
  `private_data` json NULL COMMENT '私有数据（包含密钥等敏感信息，不返回给用户）',
  `data` json NULL COMMENT '任务结果数据（JSON格式，存储输出结果）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tasks_created_at`(`created_at`) USING BTREE,
  INDEX `idx_tasks_channel_id`(`channel_id`) USING BTREE,
  INDEX `idx_tasks_action`(`action`) USING BTREE,
  INDEX `idx_tasks_submit_time`(`submit_time`) USING BTREE,
  INDEX `idx_tasks_progress`(`progress`) USING BTREE,
  INDEX `idx_tasks_task_id`(`task_id`) USING BTREE,
  INDEX `idx_tasks_platform`(`platform`) USING BTREE,
  INDEX `idx_tasks_user_id`(`user_id`) USING BTREE,
  INDEX `idx_tasks_status`(`status`) USING BTREE,
  INDEX `idx_tasks_start_time`(`start_time`) USING BTREE,
  INDEX `idx_tasks_finish_time`(`finish_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='异步任务表（音频、视频生成等）' ROW_FORMAT = Dynamic;

-- ======================================
-- 14. tokens 表 - API令牌表
-- 描述：存储用户生成的 API 访问令牌
-- 用途：API密钥管理、额度控制、IP白名单、模型限制
-- ======================================
CREATE TABLE `tokens`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID（外键关联users表）',
  `key` char(48) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'API密钥（sk-开头，48位，唯一）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '状态（1=启用, 2=已耗尽, 3=已过期）',
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '令牌名称（用于标识用途）',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `accessed_time` bigint(0) NULL DEFAULT NULL COMMENT '最后访问时间（Unix时间戳）',
  `expired_time` bigint(0) NULL DEFAULT -1 COMMENT '过期时间（Unix时间戳，-1=永不过期）',
  `remain_quota` bigint(0) NULL DEFAULT 0 COMMENT '剩余额度',
  `unlimited_quota` tinyint(1) NULL DEFAULT NULL COMMENT '是否无限额度（1=是，不受额度限制）',
  `model_limits_enabled` tinyint(1) NULL DEFAULT NULL COMMENT '是否启用模型限制（1=是，只能使用指定模型）',
  `model_limits` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '模型限制列表（逗号分隔）',
  `allow_ips` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '允许的IP地址（换行分隔，支持IP白名单）',
  `used_quota` bigint(0) NULL DEFAULT 0 COMMENT '已使用额度',
  `group` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '指定用户组（覆盖用户默认组，空=使用用户组）',
  `cross_group_retry` tinyint(1) NULL DEFAULT NULL COMMENT '是否允许跨组重试（仅auto组有效，1=是）',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_tokens_key`(`key`) USING BTREE,
  INDEX `idx_tokens_deleted_at`(`deleted_at`) USING BTREE,
  INDEX `idx_tokens_user_id`(`user_id`) USING BTREE,
  INDEX `idx_tokens_name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='API令牌表' ROW_FORMAT = Dynamic;

-- ======================================
-- 15. top_ups 表 - 充值记录表
-- 描述：记录用户在线充值订单
-- 用途：支付管理、订单追踪、对账
-- 支持支付方式：stripe（Stripe）、epay（易支付）、creem等
-- ======================================
CREATE TABLE `top_ups`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '充值记录ID',
  `user_id` bigint(0) NULL DEFAULT NULL COMMENT '用户ID',
  `amount` bigint(0) NULL DEFAULT NULL COMMENT '充值金额（美元，整数，用于易支付等）',
  `money` double NULL DEFAULT NULL COMMENT '充值金额（美元，浮点数，用于Stripe等，经分组倍率换算）',
  `trade_no` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '订单号（唯一，用于支付回调）',
  `payment_method` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '支付方式（stripe, epay, creem等）',
  `create_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `complete_time` bigint(0) NULL DEFAULT NULL COMMENT '完成时间（Unix时间戳）',
  `status` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '订单状态（pending=待支付, success=成功, failed=失败）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `trade_no`(`trade_no`) USING BTREE,
  INDEX `idx_top_ups_user_id`(`user_id`) USING BTREE,
  INDEX `idx_top_ups_trade_no`(`trade_no`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='充值记录表' ROW_FORMAT = Dynamic;

-- ======================================
-- 16. two_fa_backup_codes 表 - 双因素认证备用码表
-- 描述：存储用户的 2FA 备用恢复码
-- 用途：当用户丢失2FA设备时，可使用备用码登录
-- ======================================
CREATE TABLE `two_fa_backup_codes`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '备用码ID',
  `user_id` bigint(0) NOT NULL COMMENT '用户ID',
  `code_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '备用码哈希值（使用bcrypt加密存储）',
  `is_used` tinyint(1) NULL DEFAULT NULL COMMENT '是否已使用（1=已使用，每个备用码只能使用一次）',
  `used_at` datetime(3) NULL DEFAULT NULL COMMENT '使用时间',
  `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_two_fa_backup_codes_user_id`(`user_id`) USING BTREE,
  INDEX `idx_two_fa_backup_codes_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='双因素认证备用码表' ROW_FORMAT = Dynamic;

-- ======================================
-- 17. two_fas 表 - 双因素认证设置表
-- 描述：存储用户的 2FA (TOTP) 配置
-- 用途：提供额外安全层，支持Google Authenticator等APP
-- ======================================
CREATE TABLE `two_fas`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '2FA设置ID',
  `user_id` bigint(0) NOT NULL COMMENT '用户ID（唯一，一个用户只能有一个2FA设置）',
  `secret` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'TOTP密钥（Base32编码，用于生成验证码，不返回给前端）',
  `is_enabled` tinyint(1) NULL DEFAULT NULL COMMENT '是否启用2FA（1=启用）',
  `failed_attempts` bigint(0) NULL DEFAULT 0 COMMENT '失败尝试次数（连续失败会触发锁定）',
  `locked_until` datetime(3) NULL DEFAULT NULL COMMENT '锁定截止时间（超过此时间自动解锁）',
  `last_used_at` datetime(3) NULL DEFAULT NULL COMMENT '最后使用时间',
  `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_two_fas_user_id`(`user_id`) USING BTREE,
  INDEX `idx_two_fas_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='双因素认证设置表' ROW_FORMAT = Dynamic;

-- ======================================
-- 18. users 表 - 用户表
-- 描述：存储用户基本信息、额度和OAuth绑定
-- 用途：用户管理、认证授权、额度控制
-- 支持OAuth：GitHub、Discord、WeChat、Telegram、OIDC、LinuxDO
-- ======================================
CREATE TABLE `users`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户名（唯一，用于登录）',
  `password` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码哈希（使用bcrypt加密）',
  `display_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称（昵称）',
  `role` bigint(0) NULL DEFAULT 1 COMMENT '角色（1=普通用户, 10=管理员, 100=超级管理员/Root）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '状态（1=启用, 2=禁用）',
  `email` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '邮箱（用于找回密码、接收通知）',
  `github_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'GitHub OAuth ID（绑定GitHub账号）',
  `discord_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Discord OAuth ID（绑定Discord账号）',
  `oidc_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'OIDC OAuth ID（OpenID Connect）',
  `wechat_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '微信 OAuth ID（绑定微信账号）',
  `telegram_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Telegram OAuth ID（绑定Telegram账号）',
  `access_token` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '系统管理访问令牌（32位随机字符，用于API管理操作）',
  `quota` bigint(0) NULL DEFAULT 0 COMMENT '剩余额度（可用余额）',
  `used_quota` bigint(0) NULL DEFAULT 0 COMMENT '已使用额度（累计消费）',
  `request_count` bigint(0) NULL DEFAULT 0 COMMENT '请求次数统计（累计API调用次数）',
  `group` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'default' COMMENT '用户组（用于权限控制和定价策略）',
  `aff_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '邀请码（唯一，4位随机字符）',
  `aff_count` bigint(0) NULL DEFAULT 0 COMMENT '邀请人数（成功邀请的用户数）',
  `aff_quota` bigint(0) NULL DEFAULT 0 COMMENT '邀请剩余额度（可转移到主额度）',
  `aff_history` bigint(0) NULL DEFAULT 0 COMMENT '邀请历史额度（累计获得的邀请奖励）',
  `inviter_id` bigint(0) NULL DEFAULT NULL COMMENT '邀请人ID（谁邀请了这个用户）',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  `linux_do_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'LinuxDO OAuth ID（绑定LinuxDO账号）',
  `setting` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '用户设置（JSON格式，包含界面配置、隐私设置等）',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注说明（管理员使用）',
  `stripe_customer` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Stripe客户ID（用于Stripe支付）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE,
  UNIQUE INDEX `idx_users_access_token`(`access_token`) USING BTREE,
  UNIQUE INDEX `idx_users_aff_code`(`aff_code`) USING BTREE,
  INDEX `idx_users_deleted_at`(`deleted_at`) USING BTREE,
  INDEX `idx_users_linux_do_id`(`linux_do_id`) USING BTREE,
  INDEX `idx_users_stripe_customer`(`stripe_customer`) USING BTREE,
  INDEX `idx_users_username`(`username`) USING BTREE,
  INDEX `idx_users_display_name`(`display_name`) USING BTREE,
  INDEX `idx_users_git_hub_id`(`github_id`) USING BTREE,
  INDEX `idx_users_discord_id`(`discord_id`) USING BTREE,
  INDEX `idx_users_oidc_id`(`oidc_id`) USING BTREE,
  INDEX `idx_users_inviter_id`(`inviter_id`) USING BTREE,
  INDEX `idx_users_email`(`email`) USING BTREE,
  INDEX `idx_users_we_chat_id`(`wechat_id`) USING BTREE,
  INDEX `idx_users_telegram_id`(`telegram_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='用户表' ROW_FORMAT = Dynamic;

-- ======================================
-- 19. vendors 表 - 供应商表
-- 描述：存储 AI 服务供应商信息（OpenAI、Anthropic等）
-- 用途：模型分类、供应商管理、图标显示
-- ======================================
CREATE TABLE `vendors`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '供应商ID',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '供应商名称（唯一，如OpenAI、Anthropic、Google等）',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '供应商描述',
  `icon` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '图标名称（使用@lobehub/icons图标库）',
  `status` bigint(0) NULL DEFAULT 1 COMMENT '状态（1=启用, 0=禁用）',
  `created_time` bigint(0) NULL DEFAULT NULL COMMENT '创建时间（Unix时间戳）',
  `updated_time` bigint(0) NULL DEFAULT NULL COMMENT '更新时间（Unix时间戳）',
  `deleted_at` datetime(3) NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_vendor_name_delete_at`(`name`, `deleted_at`) USING BTREE,
  INDEX `idx_vendors_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='AI服务供应商表' ROW_FORMAT = Dynamic;

-- ======================================
-- 表结构说明总结
-- ======================================
--
-- 核心业务表：
-- - users: 用户管理和认证
-- - tokens: API密钥管理
-- - channels: AI服务商接入配置
-- - abilities: 模型-渠道-用户组映射
-- - logs: 审计和使用记录
--
-- 任务和资源表：
-- - tasks: 异步任务（音频、视频）
-- - midjourneys: Midjourney绘图任务
-- - quota_data: 使用量统计
--
-- 支付和充值表：
-- - top_ups: 在线充值订单
-- - redemptions: 兑换码
--
-- 元数据和配置表：
-- - models: 模型目录
-- - vendors: 供应商目录
-- - prefill_groups: 配置模板
-- - options: 系统配置
--
-- 安全和认证表：
-- - two_fas: 双因素认证
-- - two_fa_backup_codes: 2FA备用码
-- - passkey_credentials: Passkey无密码登录
--
-- 其他：
-- - checkins: 签到奖励
-- - setups: 初始化记录
--
-- ======================================
-- 索引说明：
-- ======================================
-- 所有表都有适当的索引以优化查询性能：
-- - 主键索引：用于唯一标识记录
-- - 外键索引：优化关联查询（如user_id, channel_id）
-- - 业务索引：优化常见查询（如status, created_at, model_name）
-- - 唯一索引：保证数据唯一性（如username, api_key）
-- - 组合索引：优化多条件查询
--
-- ======================================
-- 软删除：
-- ======================================
-- 多个表使用deleted_at字段实现软删除（GORM标准）：
-- - users, tokens, channels, models, vendors等
-- - 软删除的记录不会被普通查询检索
-- - 可以通过Unscoped()查询已删除记录
-- - 支持数据恢复和审计追踪
--
-- ======================================
