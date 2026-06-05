#!/bin/bash

echo "======================================"
echo "  WebShark - 网络抓包工具"
echo "======================================"
echo ""

# 检查依赖
echo "检查依赖..."

if ! command -v sshpass &> /dev/null; then
    echo "❌ sshpass 未安装"
    echo "   macOS: brew install sshpass"
    echo "   Ubuntu/Debian: sudo apt-get install sshpass"
    exit 1
else
    echo "✅ sshpass 已安装"
fi

if ! command -v tshark &> /dev/null; then
    echo "❌ tshark 未安装"
    echo "   macOS: brew install wireshark"
    echo "   Ubuntu/Debian: sudo apt-get install tshark"
    exit 1
else
    echo "✅ tshark 已安装"
fi

echo ""
echo "启动 WebShark 服务器..."
echo "访问地址: http://localhost:8080"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 运行程序
./webshark
