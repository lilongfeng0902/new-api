#!/bin/bash
# Cqtai渠道添加脚本

echo "🔧 添加 Cqtai 渠道到数据库..."
echo ""

# 检查API Key是否已修改
if grep -q "your-cqtai-api-key-here" scripts/add_cqtai_channel.go; then
    echo "⚠️  警告: 检测到默认API Key"
    echo "请先编辑 scripts/add_cqtai_channel.go 文件"
    echo "将第14行的 'your-cqtai-api-key-here' 改为你的实际Cqtai API Key"
    echo ""
    read -p "如果你已经修改了API Key，按回车继续... (Ctrl+C 取消)" 
fi

# 运行Go脚本
go run scripts/add_cqtai_channel.go

echo ""
echo "✅ 完成！"
