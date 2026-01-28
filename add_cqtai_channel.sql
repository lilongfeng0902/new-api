-- ============================================
-- Cqtai 渠道数据插入脚本
-- 渠道类型: 58 (ChannelTypeCqtai)
-- 支持模型: suno_fetch, suno_music, suno_lyrics, suno_voice_separation, suno_instrument_separation
-- ============================================

-- 1. 插入 Cqtai 渠道配置
-- 请根据实际情况修改以下字段:
--   - key: 你的 Cqtai API Key
--   - name: 渠道名称（可自定义）
--   - base_url: API地址（默认 https://api.cqtai.com）
--   - models: 支持的模型列表
--   - status: 1=启用, 2=手动禁用, 3=自动禁用

INSERT INTO channels (
    type,
    name,
    key,
    base_url,
    models,
    status,
    weight,
    priority,
    created_time,
    test_time,
    balance_updated_time,
    used_quota,
    `group`,
    auto_ban
) VALUES (
    58,                                          -- type: Cqtai渠道类型
    'Cqtai',                                     -- name: 渠道名称（可修改）
    'sk-6d1cefce7ccc41c8b31d06d3659be36efzrWxMh9WVLk8ZYb',                   -- key: 请替换为你的实际API Key
    'https://api.cqtai.com',                     -- base_url: Cqtai API地址
    'suno_fetch, suno_music, suno_lyrics, suno_voice_separation, suno_instrument_separation',                  -- models: 支持的模型
    1,                                           -- status: 1=启用
    0,                                           -- weight: 权重
    0,                                           -- priority: 优先级
    strftime('%s', 'now'),                       -- created_time: 创建时间
    0,                                           -- test_time: 测试时间
    0,                                           -- balance_updated_time: 余额更新时间
    0,                                           -- used_quota: 已用配额
    'default',                                   -- group: 分组
    1                                            -- auto_ban: 自动禁用
);

-- 2. 获取刚插入的渠道ID（用于后续插入abilities）
-- 注意：SQLite使用 last_insert_rowid()
-- 如果使用 MySQL，则改为 LAST_INSERT_ID()
-- 如果使用 PostgreSQL，则改为 LASTVAL()

-- 3. 插入模型能力数据（abilities表）
-- 需要为每个模型创建一条记录

-- 为 suno_music 模型创建ability
INSERT INTO abilities (
    channel_id,
    enabled,
    priority,
    weight,
    model,
    `group`
) VALUES (
    (SELECT id FROM channels WHERE type = 58 ORDER BY id DESC LIMIT 1),  -- 获取最新的Cqtai渠道ID
    1,                                           -- enabled: 1=启用
    0,                                           -- priority: 优先级
    0,                                           -- weight: 权重
    'suno_music',                                -- model: 模型名称
    'default'                                    -- group: 分组
);

-- 为 suno_lyrics 模型创建ability
INSERT INTO abilities (
    channel_id,
    enabled,
    priority,
    weight,
    model,
    `group`
) VALUES (
    (SELECT id FROM channels WHERE type = 58 ORDER BY id DESC LIMIT 1),  -- 获取最新的Cqtai渠道ID
    1,                                           -- enabled: 1=启用
    0,                                           -- priority: 优先级
    0,                                           -- weight: 权重
    'suno_lyrics',                               -- model: 模型名称
    'default'                                    -- group: 分组
);


-- 为 suno_fetch 模型创建ability
INSERT INTO abilities (
    channel_id,
    enabled,
    priority,
    weight,
    model,
    `group`
) VALUES (
    (SELECT id FROM channels WHERE type = 58 ORDER BY id DESC LIMIT 1),  -- 获取最新的Cqtai渠道ID
    1,                                           -- enabled: 1=启用
    0,                                           -- priority: 优先级
    0,                                           -- weight: 权重
    'suno_fetch',                                -- model: 模型名称
    'default'                                    -- group: 分组
);


-- 为 suno_voice_separation 模型创建ability
INSERT INTO abilities (
    channel_id,
    enabled,
    priority,
    weight,
    model,
    `group`
) VALUES (
    (SELECT id FROM channels WHERE type = 58 ORDER BY id DESC LIMIT 1),  -- 获取最新的Cqtai渠道ID
    1,                                           -- enabled: 1=启用
    0,                                           -- priority: 优先级
    0,                                           -- weight: 权重
    'suno_voice_separation',                     -- model: 模型名称
    'default'                                    -- group: 分组
);


-- 为 suno_instrument_separation 模型创建ability
INSERT INTO abilities (
    channel_id,
    enabled,
    priority,
    weight,
    model,
    `group`
) VALUES (
    (SELECT id FROM channels WHERE type = 58 ORDER BY id DESC LIMIT 1),  -- 获取最新的Cqtai渠道ID
    1,                                           -- enabled: 1=启用
    0,                                           -- priority: 优先级
    0,                                           -- weight: 权重
    'suno_instrument_separation',                -- model: 模型名称
    'default'                                    -- group: 分组
);
-- ============================================
-- 验证脚本
-- ============================================

-- 查询刚添加的Cqtai渠道
SELECT id, name, type, base_url, models, status
FROM channels
WHERE type = 58;

-- 查询Cqtai渠道的abilities
SELECT a.id, a.channel_id, a.model, a.enabled, c.name as channel_name
FROM abilities a
JOIN channels c ON a.channel_id = c.id
WHERE c.type = 58;

-- ============================================
-- 使用说明
-- ============================================
--
-- 执行方式：
-- 1. 使用 sqlite3 命令行工具：
--    sqlite3 one-api.db < add_cqtai_channel.sql
--
-- 2. 或者分步执行：
--    sqlite3 one-api.db
--    sqlite> .read add_cqtai_channel.sql
--
-- 3. 或者使用数据库管理工具（如 DB Browser for SQLite）
--    手动执行这些SQL语句
--
-- 重要提示：
-- - 请务必将 'your-cqtai-api-key-here' 替换为你的实际 Cqtai API Key
-- - 如果需要修改渠道名称、分组等，请相应调整SQL语句
-- - 如果使用 MySQL 或 PostgreSQL，请调整时间戳函数和ID获取方式
--
-- ============================================
