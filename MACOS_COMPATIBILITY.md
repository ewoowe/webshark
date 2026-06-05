# macOS 兼容性说明

## ✅ 已实现的 macOS 支持

WebShark 现已完全兼容 macOS 系统，包括本地运行和远程连接 macOS 主机。

## 🔧 macOS 特定配置

### 1. 安装必要工具

```bash
# 安装 sshpass
brew install sshpass

# 安装 Wireshark (包含 tshark)
brew install wireshark

# tcpdump macOS 自带，无需安装
```

### 2. 启用 SSH 服务

在 macOS 远程主机上：

1. 打开 **系统偏好设置**
2. 点击 **共享**
3. 勾选 **远程登录**
4. 选择允许访问的用户

或者使用命令行：

```bash
sudo systemsetup -setremotelogin on
```

### 3. 授予 tcpdump 权限

macOS 可能对网络抓包有安全限制，需要授予权限：

```bash
# 方法一：使用 sudo 运行 tcpdump（推荐）
sudo tcpdump -i en0

# 方法二：将用户添加到 access_bpf 组
sudo dseditgroup -o edit -n /Local/Default -a $(whoami) access_bpf
```

### 4. 防火墙配置

如果启用了防火墙，需要允许 SSH 连接：

1. 打开 **系统偏好设置** -> **安全性与隐私** -> **防火墙**
2. 点击 **防火墙选项**
3. 确保 **SSH** 在允许的服务列表中

## 🌐 macOS 网卡说明

### 常见网卡名称

- **en0**: 通常是 WiFi 或以太网（主要网卡）
- **en1**: 第二个网络接口（如有）
- **lo0**: 回环接口（localhost）
- **bridge0**: 桥接接口
- **utun0**: VPN 接口

### 查看网卡信息

```bash
# 查看所有网卡
ifconfig

# 查看活动网卡
ifconfig | grep "status: active"

# 查看 IP 地址
ipconfig getifaddr en0
```

## 🔍 自动检测机制

WebShark 会自动检测远程主机的操作系统：

1. **执行 `uname -s` 命令**
   - 返回 "Darwin" → macOS
   - 返回 "Linux" → Linux

2. **根据系统类型选择命令**
   - **macOS**: 使用 `ifconfig` 获取网卡
   - **Linux**: 使用 `ip` 命令获取网卡

3. **默认网卡选择**
   - **macOS**: 未选择时使用 `en0`
   - **Linux**: 未选择时使用 `any`

## ⚠️ macOS 注意事项

### 1. 权限问题

macOS Catalina 及更高版本有更严格的权限控制：

```bash
# 如果遇到权限错误，尝试：
sudo ./webshark

# 或者授予终端完全磁盘访问权限
# 系统偏好设置 -> 安全性与隐私 -> 隐私 -> 完全磁盘访问权限
```

### 2. SIP (系统完整性保护)

SIP 可能阻止某些操作，但通常不影响 WebShark 的正常使用。

### 3. 无线网络抓包

在 macOS 上抓取 WiFi 数据可能需要：

```bash
# 禁用监控模式（如果需要）
sudo airport --disassociate

# 使用特定的 BPF 过滤器
tcp port 80 or tcp port 443
```

### 4. Docker Desktop

如果使用 Docker Desktop for Mac：

- Docker 网络在单独的命名空间中
- 需要指定正确的网卡（通常是 `docker0` 或 `bridge`）

## 📝 使用示例

### 连接 macOS 远程主机

```bash
# 1. 启动 WebShark
./webshark

# 2. 在浏览器中访问
http://localhost:8080

# 3. 输入 macOS 主机信息
#    主机: 192.168.1.100
#    用户名: your_username
#    密码: your_password

# 4. 点击"获取网卡列表"
#    系统会自动检测到 macOS 并使用 ifconfig

# 5. 选择网卡（如 en0）

# 6. 开始抓包
```

### 常用过滤器

```bash
# HTTP 流量
tcp port 80

# HTTPS 流量
tcp port 443

# 特定主机
host 192.168.1.1

# DNS 查询
udp port 53

# 排除 SSH
not port 22
```

## 🐛 故障排除

### 问题 1: "sshpass: command not found"

**解决方案:**
```bash
brew install sshpass
```

### 问题 2: "tshark: command not found"

**解决方案:**
```bash
brew install wireshark
```

### 问题 3: 无法获取网卡列表

**可能原因:**
- SSH 连接失败
- 用户名或密码错误
- SSH 服务未启用

**解决方案:**
```bash
# 测试 SSH 连接
ssh username@hostname

# 检查 SSH 服务状态
sudo launchctl list | grep ssh
```

### 问题 4: 抓包无数据

**可能原因:**
- 选择了错误的网卡
- BPF 过滤器太严格
- 没有网络流量

**解决方案:**
```bash
# 测试网卡是否有流量
sudo tcpdump -i en0 -c 10

# 去掉过滤器重新尝试
# 不填写 BPF 和 Wireshark 过滤器
```

### 问题 5: 权限被拒绝

**解决方案:**
```bash
# 使用 sudo 运行
sudo ./webshark

# 或者添加用户到 access_bpf 组
sudo dseditgroup -o edit -n /Local/Default -a $(whoami) access_bpf
```

## 🔄 与 Linux 的差异

| 特性 | Linux | macOS |
|------|-------|-------|
| 网卡查询命令 | `ip addr` | `ifconfig` |
| 默认网卡 | `any` | `en0` |
| 系统标识 | Linux | Darwin |
| tcpdump 位置 | `/usr/bin/tcpdump` | `/usr/sbin/tcpdump` |
| 权限管理 | sudo | sudo + access_bpf |

## ✨ 最佳实践

1. **始终使用 sudo**: 在 macOS 上运行 tcpdump 需要 root 权限
2. **选择合适的网卡**: en0 通常是最常用的网卡
3. **使用过滤器**: 减少捕获的数据量，提高性能
4. **测试连接**: 先测试 SSH 连接是否正常
5. **检查防火墙**: 确保 SSH 端口（22）未被阻止

## 📚 相关资源

- [macOS SSH 配置](https://support.apple.com/HT201710)
- [tcpdump 手册](https://www.tcpdump.org/manpages/tcpdump.1.html)
- [Wireshark 文档](https://www.wireshark.org/docs/)

---

**提示**: 如果在 macOS 上遇到任何问题，请查看控制台日志或使用 `-v` 标志运行以获取详细输出。
