# VS Code 阿里云短信集成测试调试指南

## 📋 概述

本文档介绍如何在 VS Code 中运行和调试阿里云短信集成测试，包括断点调试、环境配置等。

## 🔧 配置准备

### 1. 更新 launch.json

项目已配置了专门的测试启动配置：

```json
{
    "name": "Test Aliyun SMS Integration",
    "type": "go",
    "request": "launch",
    "mode": "test",
    "program": "${workspaceFolder}",
    "buildFlags": "-tags=integration",
    "env": {
        "ALIYUN_SMS_INTEGRATION_TEST": "true",
        "ALIYUN_SMS_ACCESS_KEY_ID": "your_access_key_id",
        "ALIYUN_SMS_ACCESS_KEY_SECRET": "your_access_key_secret",
        "ALIYUN_SMS_SIGN_NAME": "your_sign_name",
        "ALIYUN_SMS_TEMPLATE_CODE": "SMS_xxxxxxxxx",
        "ALIYUN_SMS_TEST_PHONE": "13800000000"
    },
    "args": [
        "-test.run",
        "TestSendAliyunSMSRealIntegration",
        "-test.v"
    ],
    "showLog": true,
    "cwd": "${workspaceFolder}"
}
```

### 2. 配置环境变量

#### 方法1：使用 launch.json（推荐用于调试）
在 `.vscode/launch.json` 中更新真实的阿里云SMS配置：

```json
"env": {
    "ALIYUN_SMS_INTEGRATION_TEST": "true",
    "ALIYUN_SMS_ACCESS_KEY_ID": "你的AccessKey ID",
    "ALIYUN_SMS_ACCESS_KEY_SECRET": "你的AccessKey Secret",
    "ALIYUN_SMS_SIGN_NAME": "你的短信签名",
    "ALIYUN_SMS_TEMPLATE_CODE": "SMS_你的模板CODE",
    "ALIYUN_SMS_TEST_PHONE": "测试手机号"
}
```

#### 方法2：使用系统环境变量（推荐用于点击运行）
在系统环境变量或 `.env` 文件中设置：

```bash
export ALIYUN_SMS_INTEGRATION_TEST=true
export ALIYUN_SMS_ACCESS_KEY_ID=你的AccessKey ID
export ALIYUN_SMS_ACCESS_KEY_SECRET=你的AccessKey Secret
export ALIYUN_SMS_SIGN_NAME=你的短信签名
export ALIYUN_SMS_TEMPLATE_CODE=SMS_你的模板CODE
export ALIYUN_SMS_TEST_PHONE=测试手机号
```

#### 方法3：使用 VS Code 工作区环境变量
在 `.vscode/settings.json` 中添加：

```json
{
    "go.testEnvVars": {
        "ALIYUN_SMS_INTEGRATION_TEST": "true",
        "ALIYUN_SMS_ACCESS_KEY_ID": "你的AccessKey ID",
        "ALIYUN_SMS_ACCESS_KEY_SECRET": "你的AccessKey Secret",
        "ALIYUN_SMS_SIGN_NAME": "你的短信签名",
        "ALIYUN_SMS_TEMPLATE_CODE": "SMS_你的模板CODE",
        "ALIYUN_SMS_TEST_PHONE": "测试手机号"
    }
}
```

## 🚀 运行方式

### 方法1：直接点击运行（推荐）

1. **设置环境变量**
   - 在系统环境变量、`.env`文件或`.vscode/settings.json`中配置阿里云SMS参数
   - 确保设置了 `ALIYUN_SMS_INTEGRATION_TEST=true`

2. **打开测试文件**
   - 打开 `common/alisms_integration_test.go`

3. **点击测试函数上方的运行按钮**
   - 在 `func TestSendAliyunSMSRealIntegration` 上方会显示一个小的 ▶️ 按钮
   - 点击即可运行该测试
   - **注意**: 如果环境变量未设置，测试会被安全跳过

### 方法2：调试面板运行

1. **打开调试面板**
   - 按 `Ctrl+Shift+D` (Windows/Linux) 或 `Cmd+Shift+D` (Mac)
   - 或点击左侧活动栏的调试图标

2. **选择测试配置**
   - 在调试配置下拉菜单中选择 "Test Aliyun SMS Integration"
   - 点击绿色播放按钮 ▶️ 开始运行

3. **查看测试结果**
   - 测试输出会显示在调试控制台中
   - 可以实时查看测试执行状态

### 方法3：使用命令面板运行

1. **打开命令面板**
   - 按 `Ctrl+Shift+P` (Windows/Linux) 或 `Cmd+Shift+P` (Mac)

2. **选择调试命令**
   - 输入 "Debug: Select and Start Debugging"
   - 选择 "Test Aliyun SMS Integration"

### 方法4：命令行运行

```bash
# 直接运行（需要预设环境变量）
go test -run TestSendAliyunSMSRealIntegration -v ./common

# 或使用makefile
make test-sms
```

## 🐛 断点调试

### 设置断点

1. **在代码中设置断点**
   - 点击行号左侧的空白区域，出现红色圆点表示断点已设置
   - 或按 `F9` 在当前行设置/取消断点

2. **推荐的断点位置**
   ```go
   // alisms_integration_test.go
   func TestSendAliyunSMSRealIntegration(t *testing.T) {
       // 第34行：测试开始处
       if os.Getenv("ALIYUN_SMS_INTEGRATION_TEST") != "true" { // 设置断点

       // 第60行：准备测试数据处
       AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"

       // 第70行：发送短信前
       err := SendAliyunSMS(testPhone, templateParam)
   }

   // alisms.go
   func SendAliyunSMS(phoneNumber string, templateParam map[string]string) error {
       // 第20行：参数验证
       if AliyunSMSEndpoint == "" || AliyunSMSAccessKeyId == "" || AliyunSMSAccessKeySecret == "" {

       // 第35行：构建签名
       signature := signStringToString(method, canonicalizedQueryString, AliyunSMSAccessKeySecret)

       // 第55行：发送HTTP请求
       resp, err := client.Get(requestURL)
   }
   ```

### 开始调试

1. **设置断点后开始调试**
   - 选择 "Test Aliyun SMS Integration" 配置
   - 点击绿色调试按钮开始

2. **调试控制**
   - **继续执行**: `F5` 或点击继续按钮
   - **单步执行**: `F10` (逐过程) 或 `F11` (逐语句)
   - **跳出**: `Shift+F11`
   - **重启**: `Ctrl+Shift+F5`
   - **停止**: `Shift+F5`

3. **查看变量值**
   - 在调试时将鼠标悬停在变量上查看值
   - 或在"变量"面板中查看所有变量
   - 在"监视"面板中添加表达式进行监控

### 调试技巧

1. **条件断点**
   - 右键点击断点，选择"编辑断点"
   - 设置条件，如 `phoneNumber == "13800000000"`

2. **日志断点**
   - 右键点击断点，选择"编辑断点"
   - 启用"日志消息"，输入要输出的信息

3. **监视表达式**
   - 在"监视"面板中添加 `AliyunSMSEndpoint`
   - 添加 `len(templateParam)` 查看参数数量

## 📊 查看调试信息

### 调试控制台

测试运行时会显示详细的调试信息：

```
=== RUN   TestSendAliyunSMSRealIntegration
=== RUN   TestSendAliyunSMSRealIntegration/send_simple_sms
    发送简单短信失败: 阿里云短信发送失败: isv.TEMPLATE_MISSING_PARAMETERS - Template param is missing
--- PASS: TestSendAliyunSMSRealIntegration/send_simple_sms
=== RUN   TestSendAliyunSMSRealIntegration/send_template_sms
    模板短信发送成功
--- PASS: TestSendAliyunSMSRealIntegration/send_template_sms
```

### 变量面板

在调试时可以查看：
- **本地变量**: 函数内的局部变量
- **全局变量**: 包级别的变量
- **函数参数**: 测试函数的参数

## 🔍 故障排除

### 常见问题

1. **点击小按钮没有运行结果 / [no tests to run]**
   ```
   ok  	github.com/QuantumNous/new-api/common	(cached) [no tests to run]
   ```
   **原因**: VS Code Go扩展默认不使用构建标签，找不到集成测试
   **解决**:
   - 方法1：在 `.vscode/settings.json` 中添加 `"go.testTags": "integration"`
   - 方法2：使用调试面板的 "Test Aliyun SMS Integration" 配置
   - 方法3：命令行运行 `go test -tags integration -run TestSendAliyunSMSRealIntegration -v ./common`

2. **测试被跳过**
   ```
   集成测试未启用，请设置 ALIYUN_SMS_INTEGRATION_TEST=true 启用
   ```
   **解决**:
   - 在系统环境变量中设置 `ALIYUN_SMS_INTEGRATION_TEST=true`
   - 或在 `.vscode/settings.json` 的 `go.testEnvVars` 中设置
   - 或在 `.env` 文件中设置

3. **缺少环境变量**
   ```
   缺少必需的环境变量: ALIYUN_SMS_ACCESS_KEY_ID
   ```
   **解决**:
   - 检查所有必需的环境变量是否已设置
   - 确保变量名拼写正确
   - 重启VS Code使环境变量生效

4. **阿里云API错误**
   ```
   阿里云短信发送失败: isv.INVALID_PARAMETERS
   ```
   **解决**:
   - 检查 AccessKey ID 和 Secret 是否正确
   - 确认短信签名已审核通过
   - 确认短信模板已审核通过且参数格式正确
   - 检查账户是否有余额

5. **断点不生效**
   **解决**: 确保选择的是 "Test Aliyun SMS Integration" 测试配置，而不是应用程序启动配置

6. **Go扩展找不到测试**
   **解决**:
   - 确保 `.vscode/settings.json` 中设置了 `"go.testTags": "integration"`
   - 或使用命令行运行时添加 `-tags integration` 参数
   - 检查文件路径是否正确

### 调试日志

启用详细日志输出：

```go
// 在测试中添加日志
t.Logf("发送请求到: %s", requestURL)
t.Logf("请求参数: %v", params)
t.Logf("响应状态: %s", resp.Status)
```

## 📝 最佳实践

1. **分阶段调试**
   - 先调试环境变量读取
   - 再调试API调用构建
   - 最后调试HTTP请求

2. **使用合适的断点**
   - 在关键决策点设置断点
   - 在API调用前后设置断点

3. **监控变量变化**
   - 关注 `AliyunSMSEndpoint` 等配置变量
   - 监控 `templateParam` 的构建过程

4. **安全注意**
   - 不要在生产环境的 AccessKey 中设置断点
   - 测试完成后清理敏感信息

## 🎯 快速开始

1. **配置环境变量**（在 launch.json 中）
2. **设置断点**（在关键代码行）
3. **选择调试配置**（"Test Aliyun SMS Integration"）
4. **点击调试按钮**（F5）
5. **单步执行**（F10）观察变量变化

通过这个调试指南，您可以轻松地在 VS Code 中进行阿里云短信集成测试的断点调试，快速定位和解决问题！ 🚀