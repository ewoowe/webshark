# WebShark - 基于 Web 的网络流量抓包工具

WebShark 是一个类似 Wireshark 的 Web 界面网络抓包分析工具，支持远程主机抓包和实时数据展示。

## ✨ 功能特性

### 1. 远程主机连接
用户通过 Web 页面输入远程主机地址、用户名和密码后，后台程序会通过执行系统命令获取所有网卡以及对应的网卡 IP，供用户选择。可多选，不选则默认抓取所有网卡。

**特性：**
- ✅ 显示所有网络接口（包括回环接口 lo0、lo 等）
- ✅ 即使网卡没有配置 IP 也会列出
- ✅ 自动检测操作系统（Linux/macOS）并使用相应命令

### 2. 过滤器配置
用户选择完网卡后，需要填写两种过滤器：
- **BPF 过滤器**: 应用在 tcpdump 上（如: `tcp port 80`, `host 192.168.1.1`）
- **Wireshark 过滤器**: 应用在 tshark 上（如: `http`, `dns`, `ip.addr == 192.168.1.1`）

### 3. 实时抓包
用户填写完过滤器后，点击开始抓包，后台程序会执行 sshpass、tcpdump 和 tshark 命令，实时解析的抓包结果会通过 WebSocket 推送到页面上。

### 4. 数据包展示
页面收到 WebSocket 推送的一条条抓包结果后，会逐条展示抓包结果，类似 Wireshark 的界面那样。点击每条抓包记录，会显示信令的详情。

## 📋 前提条件

### 本地系统要求

1. **Go 1.26+**
2. **sshpass** - SSH 密码认证工具
   ```bash
   # macOS
   brew install sshpass
   
   # Ubuntu/Debian
   sudo apt-get install sshpass
   ```

3. **tshark** - Wireshark 命令行工具
   ```bash
   # macOS
   brew install wireshark
   
   # Ubuntu/Debian
   sudo apt-get install tshark
   ```

### 远程主机要求

#### Linux 系统
1. 启用 SSH 服务
2. 用户具有 sudo 权限（用于执行 tcpdump）
3. 安装 tcpdump
   ```bash
   sudo apt-get install tcpdump
   ```
4. **配置 sudo 无密码运行 tcpdump**（重要！）
   ```bash
   # 方法1: 自动配置（WebShark 会尝试）
   # 启动抓包时会自动尝试配置
   
   # 方法2: 手动配置
   echo "username ALL=(ALL) NOPASSWD: /usr/sbin/tcpdump" | sudo tee /etc/sudoers.d/tcpdump
   sudo chmod 440 /etc/sudoers.d/tcpdump
   
   # 方法3: 使用 setcap（无需 sudo）
   sudo setcap cap_net_raw,cap_net_admin=eip /usr/bin/tcpdump
   ```

#### macOS 系统
1. 启用 SSH 服务（系统偏好设置 -> 共享 -> 远程登录）
2. 用户具有管理员权限
3. 安装 tcpdump（macOS 自带）
4. 允许 tcpdump 捕获网络数据（可能需要授予权限）
5. **配置 sudo 无密码运行 tcpdump**（重要！）
   ```bash
   # 方法1: 自动配置（WebShark 会尝试）
   # 启动抓包时会自动尝试配置
   
   # 方法2: 手动配置
   echo "username ALL=(ALL) NOPASSWD: /usr/sbin/tcpdump" | sudo tee /etc/sudoers.d/tcpdump
   sudo chmod 440 /etc/sudoers.d/tcpdump
   
   # 方法3: 添加到 access_bpf 组
   sudo dseditgroup -o edit -n /Local/Default -a username access_bpf
   ```

### 跨平台兼容性

✅ **已支持的系统：**
- Linux (Ubuntu, CentOS, Debian, etc.)
- macOS (Darwin)
- 其他类 Unix 系统（BSD 等）

系统会自动检测远程主机的操作系统类型，并使用相应的命令：
- **Linux**: 使用 `ip` 命令获取网卡，`tcpdump` 抓包
- **macOS**: 使用 `ifconfig` 命令获取网卡，`tcpdump` 抓包

> **技术细节**: macOS 上使用纯 Shell 命令（ifconfig + grep + awk）解析网卡信息，无需 Python 依赖。

## 🚀 快速开始

### 方式一：使用启动脚本

```bash
# 赋予执行权限
chmod +x start.sh

# 运行
./start.sh
```

### 方式二：手动运行

```bash
# 1. 下载依赖
go mod tidy

# 2. 编译
go build -o webshark

# 3. 运行
./webshark
```

或者直接运行：
```bash
go run main.go
```

访问 http://localhost:8080 即可使用。

## 📖 使用说明

### 步骤 1: 连接远程主机
1. 在"远程主机连接"区域输入：
   - 主机地址（IP 或域名）
   - SSH 用户名
   - SSH 密码
2. 点击"获取网卡列表"按钮

### 步骤 2: 选择网卡
1. 在"选择网卡"区域会显示所有可用的网络接口
2. 勾选需要抓包的网卡（可多选）
3. 不选择则默认抓取所有网卡

### 步骤 3: 配置过滤器（可选）
1. **BPF 过滤器**: 用于 tcpdump 层面过滤
   - 示例: `tcp port 80` - 只抓取 HTTP 流量
   - 示例: `host 192.168.1.1` - 只抓取特定主机的流量
   - 示例: `not port 22` - 排除 SSH 流量

2. **Wireshark 过滤器**: 用于 tshark 层面过滤
   - 示例: `http` - 只显示 HTTP 协议
   - 示例: `dns` - 只显示 DNS 协议
   - 示例: `ip.addr == 192.168.1.1` - 只显示特定 IP 的数据包

### 步骤 4: 开始抓包
1. 点击"开始抓包"按钮
2. 等待 WebSocket 连接建立
3. 数据包会实时显示在表格中

### 步骤 5: 查看数据包
1. 点击表格中的任意数据包
2. 下方会显示该数据包的详细信息（JSON 格式）
3. 可以实时查看数据包统计

### 步骤 6: 停止抓包
点击"停止抓包"按钮结束当前抓包会话。

## 🏗️ 项目结构

详见 [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)

## 🔧 API 接口

### REST API

#### 获取网卡列表
```http
GET /api/interfaces?host={host}&username={username}&password={password}
```

#### 开始抓包
```http
POST /api/capture/start
Content-Type: application/json

{
  "host": "string",
  "username": "string",
  "password": "string",
  "interfaces": ["string"],
  "bpf_filter": "string",
  "wireshark_filter": "string"
}
```

#### 停止抓包
```http
POST /api/capture/stop?session_id={sessionId}
```

### WebSocket

#### 实时数据包推送
```ws
WS /ws/capture?session_id={sessionId}
```

## ⚠️ 注意事项

1. **安全性**: 
   - 当前实现在命令行中传递密码，仅适用于测试环境
   - 生产环境应使用 SSH 密钥认证或其他安全方式

2. **权限要求**: 
   - 远程用户需要 sudo 权限来执行 tcpdump
   - 确保远程主机的 sudoers 配置正确
   - 详见 [TSHARK_JSON_FIX.md](TSHARK_JSON_FIX.md) 中的权限配置说明

3. **性能考虑**: 
   - 大量数据包时可能需要优化前端渲染
   - 建议在高流量环境下使用 BPF 过滤器减少数据量

4. **浏览器兼容性**: 
   - 需要使用支持 WebSocket 的现代浏览器
   - 推荐使用 Chrome、Firefox、Safari 最新版

5. **故障排除**:
   - 如果抓包返回空数组 `[]`，请检查：
     - sudo 是否需要密码（配置 NOPASSWD）
     - tcpdump 是否正确安装
     - tshark 版本是否为 3.x 以上
     - 网络是否有流量
   - 详见 [TSHARK_JSON_FIX.md](TSHARK_JSON_FIX.md)
   
6. **Sudo 权限问题**:
   - 如果遇到 "sudo: no tty present" 错误
   - WebShark 会自动尝试配置 sudo 权限
   - 如果失败，请手动配置（见 [SUDO_SETUP.md](SUDO_SETUP.md)）

## 🛠️ 技术栈

### 后端
- **Go 1.26** - 编程语言
- **gorilla/websocket** - WebSocket 库
- **golang.org/x/crypto** - 加密库
- 系统工具: sshpass, tcpdump, tshark

### 前端
- HTML5 / CSS3 / JavaScript (原生)
- WebSocket API
- Fetch API

## 📝 后续改进方向

- [ ] 添加用户认证和授权机制
- [ ] 支持保存抓包结果到 pcap 文件
- [ ] 支持上传 pcap 文件进行分析
- [ ] 更完善的 Wireshark 过滤器语法支持
- [ ] 数据包统计图表可视化
- [ ] 支持多种导出格式（CSV、JSON、PCAP）
- [ ] 会话管理和历史记录
- [ ] 支持 SSH 密钥认证
- [ ] 添加 HTTPS 支持
- [ ] 实现数据包搜索和过滤功能

## 📄 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**注意**: 本项目仅供学习和研究使用，请遵守相关法律法规，不要用于非法用途。