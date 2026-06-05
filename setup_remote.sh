#!/bin/bash

# WebShark 远程主机配置脚本
# 用于在远程主机上配置 tcpdump 权限

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "  WebShark 远程主机配置工具"
echo "========================================"
echo ""

# 检查参数
if [ $# -lt 2 ]; then
    echo -e "${YELLOW}用法:${NC}"
    echo "  $0 <用户名> <主机地址>"
    echo ""
    echo "示例:"
    echo "  $0 root 192.168.1.100"
    echo "  $0 admin example.com"
    echo ""
    exit 1
fi

USERNAME=$1
HOST=$2

echo -e "${GREEN}目标主机:${NC} ${USERNAME}@${HOST}"
echo ""

# 询问密码
read -s -p "请输入 SSH 密码: " PASSWORD
echo ""

# 检查 sshpass 是否安装
if ! command -v sshpass &> /dev/null; then
    echo -e "${RED}错误: sshpass 未安装${NC}"
    echo ""
    echo "请先安装 sshpass:"
    echo "  macOS: brew install sshpass"
    echo "  Ubuntu/Debian: sudo apt-get install sshpass"
    exit 1
fi

# 检测操作系统
echo -e "${YELLOW}[1/5]${NC} 检测远程操作系统..."
OS=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "uname -s" 2>/dev/null)

if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 无法连接到远程主机${NC}"
    exit 1
fi

echo -e "${GREEN}检测到系统:${NC} $OS"
echo ""

# 查找 tcpdump 路径
echo -e "${YELLOW}[2/5]${NC} 查找 tcpdump 路径..."
TCPDUMP_PATH=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "which tcpdump" 2>/dev/null)

if [ -z "$TCPDUMP_PATH" ]; then
    echo -e "${RED}错误: tcpdump 未安装在远程主机上${NC}"
    echo ""
    echo "请在远程主机上安装 tcpdump:"
    if [ "$OS" == "Darwin" ]; then
        echo "  macOS 通常预装 tcpdump，请检查 PATH"
    else
        echo "  Ubuntu/Debian: sudo apt-get install tcpdump"
        echo "  CentOS/RHEL: sudo yum install tcpdump"
    fi
    exit 1
fi

echo -e "${GREEN}tcpdump 路径:${NC} $TCPDUMP_PATH"
echo ""

# 测试当前 sudo 权限
echo -e "${YELLOW}[3/5]${NC} 测试当前 sudo 权限..."
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "sudo -n $TCPDUMP_PATH --version" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ sudo 权限已正确配置（无需密码）${NC}"
    echo ""
    echo -e "${GREEN}配置完成！${NC}"
    echo "你现在可以使用 WebShark 进行抓包了。"
    exit 0
else
    echo -e "${YELLOW}⚠ sudo 需要密码，正在配置...${NC}"
fi

# 尝试配置 sudoers
echo -e "${YELLOW}[4/5]${NC} 配置 sudo NOPASSWD..."

# 根据系统类型选择配置方法
if [ "$OS" == "Darwin" ]; then
    # macOS
    SUDOERS_FILE="/etc/sudoers.d/tcpdump"
else
    # Linux
    SUDOERS_FILE="/etc/sudoers.d/tcpdump"
fi

# 尝试创建 sudoers 配置文件
CONFIG_CMD="echo '${USERNAME} ALL=(ALL) NOPASSWD: ${TCPDUMP_PATH}' | sudo tee ${SUDOERS_FILE} && sudo chmod 440 ${SUDOERS_FILE}"

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "$CONFIG_CMD" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ sudoers 配置成功${NC}"
else
    echo -e "${YELLOW}⚠ 自动配置失败，尝试其他方法...${NC}"
    
    # 尝试使用 visudo
    VISUDO_CMD="echo '${USERNAME} ALL=(ALL) NOPASSWD: ${TCPDUMP_PATH}' | sudo EDITOR='tee -a' visudo"
    
    sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "$VISUDO_CMD" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ visudo 配置成功${NC}"
    else
        echo -e "${RED}✗ 自动配置失败${NC}"
        echo ""
        echo -e "${YELLOW}请手动配置:${NC}"
        echo ""
        echo "1. SSH 登录到远程主机:"
        echo "   ssh ${USERNAME}@${HOST}"
        echo ""
        echo "2. 执行以下命令:"
        echo "   sudo visudo"
        echo ""
        echo "3. 在文件末尾添加:"
        echo "   ${USERNAME} ALL=(ALL) NOPASSWD: ${TCPDUMP_PATH}"
        echo ""
        echo "4. 保存并退出 (:wq)"
        echo ""
        
        read -p "是否继续测试配置？(y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
fi

# 验证配置
echo -e "${YELLOW}[5/5]${NC} 验证配置..."
sleep 1

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no ${USERNAME}@${HOST} "sudo -n $TCPDUMP_PATH --version" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 验证成功！${NC}"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  配置完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "现在你可以使用 WebShark 进行抓包了。"
    echo ""
    echo "下一步:"
    echo "  1. 启动 WebShark: ./webshark"
    echo "  2. 访问 http://localhost:8080"
    echo "  3. 输入主机信息: ${HOST}"
    echo "  4. 开始抓包"
    echo ""
else
    echo -e "${RED}✗ 验证失败${NC}"
    echo ""
    echo -e "${YELLOW}配置可能未生效，请检查:${NC}"
    echo ""
    echo "1. 检查 sudoers 文件语法:"
    echo "   ssh ${USERNAME}@${HOST} 'sudo visudo -c'"
    echo ""
    echo "2. 检查文件权限:"
    echo "   ssh ${USERNAME}@${HOST} 'ls -l /etc/sudoers.d/tcpdump'"
    echo ""
    echo "3. 查看详细错误信息:"
    echo "   ssh ${USERNAME}@${HOST} 'sudo -n tcpdump --version'"
    echo ""
    echo "更多信息请参考: SUDO_SETUP.md"
    exit 1
fi
