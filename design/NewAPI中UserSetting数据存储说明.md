# UserSetting 数据存储结构说明

## 📋 数据存储位置

**UserSetting 不是存储在单独的数据库表中**，而是作为 **JSON 字符串** 存储在 `users` 表的 `setting` 字段中。

## 🗄️ 数据库表结构

### users 表中的 setting 字段

```sql
`setting` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '用户设置（JSON格式，包含界面配置、隐私设置等）'
```

从数据库表定义可以看到：
- **字段类型**: `text`（可存储长文本）
- **字符集**: `utf8mb4`（支持中文字符）
- **允许NULL**: 是
- **用途**: 存储用户的个性化设置

## 🔄 数据读写流程

### 1. 数据写入（序列化）

```go
// dto/user_settings.go 中定义的结构体
type UserSetting struct {
    NotifyType            string  `json:"notify_type,omitempty"`
    QuotaWarningThreshold float64 `json:"quota_warning_threshold,omitempty"`
    WebhookUrl            string  `json:"webhook_url,omitempty"`
    // ... 其他字段
}

// model/user.go 中的序列化方法
func (user *User) SetSetting(setting dto.UserSetting) {
    settingBytes, err := json.Marshal(setting)  // 序列化为JSON
    if err != nil {
        common.SysLog("failed to marshal setting: " + err.Error())
        return
    }
    user.Setting = string(settingBytes)  // 存储为字符串
}
```

### 2. 数据读取（反序列化）

```go
// model/user.go 中的反序列化方法
func (user *User) GetSetting() dto.UserSetting {
    setting := dto.UserSetting{}
    if user.Setting != "" {
        err := json.Unmarshal([]byte(user.Setting), &setting)  // 从JSON反序列化
        if err != nil {
            common.SysLog("failed to unmarshal setting: " + err.Error())
        }
    }
    return setting
}
```

## 💾 存储示例

### 数据库中的存储内容
```sql
-- users 表中的一条记录
INSERT INTO users (id, username, setting, ...) VALUES (
    1,
    'testuser',
    '{"notify_type":"sms","quota_warning_threshold":500,"sms_phone_number":"13800000000","webhook_url":"","bark_url":"","gotify_url":"","gotify_token":"","gotify_priority":0,"accept_unset_model_ratio_model":false,"record_ip_log":false,"sidebar_modules":""}',
    ...
);
```

### JSON 结构示例
```json
{
  "notify_type": "sms",
  "quota_warning_threshold": 500,
  "sms_phone_number": "13800000000",
  "webhook_url": "",
  "bark_url": "",
  "gotify_url": "",
  "gotify_token": "",
  "gotify_priority": 0,
  "accept_unset_model_ratio_model": false,
  "record_ip_log": false,
  "sidebar_modules": ""
}
```

## 🎯 设计优势

### 1. **灵活性**
- 无需修改数据库表结构即可添加新设置项
- 支持嵌套的复杂数据结构
- 字段变更不影响现有数据

### 2. **性能优化**
- 减少数据库表的列数
- 避免频繁的ALTER TABLE操作
- 减少JOIN查询的复杂性

### 3. **版本兼容性**
- 旧版本的数据仍然可以被新版本解析
- 新增字段默认为零值或空字符串
- 支持字段的重命名和废弃

## 🔍 缓存机制

### Redis 缓存
UserSetting 还支持 Redis 缓存以提升性能：

```go
// model/user.go
func GetUserSetting(id int, fromDB bool) (settingMap dto.UserSetting, err error) {
    if !fromDB && common.RedisEnabled {
        // 先从Redis缓存读取
        setting, err := getUserSettingCache(id)
        if err == nil {
            return setting, nil
        }
        // 缓存未命中则从数据库读取
    }
    // 从数据库读取并异步更新缓存
}
```

### 缓存 Key 格式
```
user_setting:{userId}
```

## 📝 相关代码位置

### 核心文件
1. **`dto/user_settings.go`** - UserSetting 结构体定义
2. **`model/user.go`** - 数据存取逻辑
3. **`model/user_cache.go`** - 缓存管理逻辑

### 使用示例
```go
// 读取用户设置
userSetting := user.GetSetting()

// 修改设置
userSetting.NotifyType = "sms"
userSetting.SMSPhoneNumber = "13800000000"
user.SetSetting(userSetting)

// 保存到数据库
DB.Save(&user)
```

## 🚀 扩展建议

如果将来需要更复杂的设置管理，可以考虑：

1. **单独的设置表** - 当设置项过多时
2. **版本控制** - 添加设置版本号字段
3. **设置分组** - 将设置按功能分组存储
4. **设置验证** - 添加设置项的合法性验证

这种JSON存储的设计在快速迭代的Web应用中非常常见，既保证了灵活性又维护了良好的性能。 👍