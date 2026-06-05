# WebShark 项目结构说明

## 目录结构

```
webshark/
├── main.go                          # 主入口文件
├── go.mod                           # Go 模块依赖
├── internal/
│   ├── handler/                     # HTTP 处理器
│   │   ├── router.go               # 路由配置
│   │   ├── api.go                  # API 处理器（获取网卡、开始/停止抓包）
│   │   └── websocket.go            # WebSocket 处理器
│   ├── service/                     # 业务逻辑层
│   │   ├── interface_service.go    # 获取远程网卡信息
│   │   └── capture_service.go      # 抓包服务（tcpdump + tshark）
│   ├── middleware/                  # 中间件
│   │   └── logging.go              # 日志中间件
│   └── server/                      # 服务器封装
│       └── server.go               # HTTP 服务器配置
└── static/                          # 前端静态文件
    ├── index.html                   # 主页面
    ├── css/
    │   └── style.css               # 样式文件
    └── js/
        └── app.js                  # 前端 JavaScript 逻辑
```

## 功能模块说明

### 1. 远程主机连接
- **API**: `/api/interfaces`
- **功能**: 通过 SSH 连接到远程主机，获取所有网络接口及其 IP 地址
- **实现**: 使用 `sshpass` 执行远程命令

### 2. 网卡选择
- 用户可以在前端多选需要抓包的网卡
- 如果不选择，默认抓取所有网卡（使用 "any"）

### 3. 过滤器配置
- **BPF 过滤器**: 应用于 tcpdump 层面（如: `tcp port 80`）
- **Wireshark 过滤器**: 应用于 tshark 解析层面（如: `http.request.method == GET`）

### 4. 实时抓包
- **API**: `/api/capture/start`
- **功能**: 启动远程抓包会话
- **实现**: 
  - 使用 `sshpass` + `ssh` 连接到远程主机
  - 在远程主机上执行 `tcpdump` 抓包
  - 将抓包数据通过管道传递给本地的 `tshark` 进行 JSON 格式解析
  - 解析后的数据包通过 WebSocket 实时推送到前端

### 5. WebSocket 实时推送
- **端点**: `/ws/capture?session_id={sessionId}`
- **功能**: 实时推送捕获的数据包到前端
- **协议**: WebSocket (gorilla/websocket)

### 6. 前端展示
- 类似 Wireshark 的界面风格
- 数据包列表表格显示
- 点击数据包查看详细信息（JSON 格式）
- 实时统计数据包数量

## 技术栈

### 后端
- **Go 1.26**
- **gorilla/websocket**: WebSocket 支持
- **golang.org/x/crypto**: SSH 相关功能
- 系统命令调用: sshpass, tcpdump, tshark

### 前端
- 原生 HTML5/CSS3/JavaScript
- WebSocket API
- Fetch API

## 运行前提条件

### 系统要求
1. **安装 sshpass**:
   ```bash
   # macOS
   brew install sshpass
   
   # Ubuntu/Debian
   sudo apt-get install sshpass
   ```

2. **安装 tcpdump** (在远程主机上):
   ```bash
   sudo apt-get install tcpdump
   ```

3. **安装 tshark** (Wireshark 命令行工具):
   ```bash
   # macOS
   brew install wireshark
   
   # Ubuntu/Debian
   sudo apt-get install tshark
   ```

### 远程主机要求
1. 启用 SSH 服务
2. 用户具有 sudo 权限（用于执行 tcpdump）
3. 安装 tcpdump

## 运行方式

```bash
# 下载依赖
go mod tidy

# 运行程序
go run main.go

# 或使用 build
go build -o webshark
./webshark
```

访问 http://localhost:8080

## API 接口

### 1. 获取网卡列表
```
GET /api/interfaces?host={host}&username={username}&password={password}
```

### 2. 开始抓包
```
POST /api/capture/start
Body: {
  "host": "string",
  "username": "string",
  "password": "string",
  "interfaces": ["string"],
  "bpf_filter": "string",
  "wireshark_filter": "string"
}
```

### 3. 停止抓包
```
POST /api/capture/stop?session_id={sessionId}
```

### 4. WebSocket 连接
```
WS /ws/capture?session_id={sessionId}
```

## 注意事项

1. **安全性**: 当前实现在命令行中传递密码，生产环境应使用更安全的认证方式
2. **性能**: 大量数据包时可能需要优化前端渲染和后端处理
3. **权限**: 远程用户需要 sudo 权限来执行 tcpdump
4. **过滤器**: Wireshark 过滤器的实现是简化版本，复杂过滤建议在 tshark 命令层面实现

## 后续改进方向

1. 添加用户认证和授权
2. 支持保存抓包结果到 pcap 文件
3. 支持上传 pcap 文件进行分析
4. 更完善的过滤器语法支持
5. 数据包统计图表
6. 支持多种导出格式
7. 会话管理和历史记录
