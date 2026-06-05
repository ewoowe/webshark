# 跨平台兼容性实现

## 🎯 概述

WebShark 现已完全支持 Linux 和 macOS 系统，能够自动检测远程主机类型并使用相应的命令。

## 🔍 自动检测机制

### 1. 操作系统检测

```go
func detectRemoteOS(host, username, password string) (string, error) {
    cmd := exec.Command("sshpass", "-p", password, "ssh", 
        "-o", "StrictHostKeyChecking=no",
        fmt.Sprintf("%s@%s", username, host),
        "uname -s")
    
    // 执行命令并解析结果
    osName := strings.TrimSpace(stdout.String())
    switch osName {
    case "Darwin":
        return "darwin", nil  // macOS
    case "Linux":
        return "linux", nil   // Linux
    default:
        return osName, nil
    }
}
```

**工作原理:**
- 通过 SSH 执行 `uname -s` 命令
- Darwin = macOS
- Linux = Linux

### 2. 网卡获取策略

```go
func GetRemoteInterfaces(host, username, password string) ([]NetworkInterface, error) {
    // 首先尝试 Linux 的 ip 命令
    interfaces, err = getInterfacesWithIp(host, username, password)
    if err == nil && len(interfaces) > 0 {
        return interfaces, nil
    }
    
    // 如果失败，尝试 macOS 的 ifconfig 命令
    interfaces, err = getInterfacesWithIfconfig(host, username, password)
    if err != nil {
        return nil, fmt.Errorf("failed to get interfaces (tried both ip and ifconfig): %v", err)
    }
    
    return interfaces, nil
}
```

**回退机制:**
1. 先尝试 `ip` 命令（Linux）
2. 如果失败，尝试 `ifconfig` 命令（macOS/BSD）
3. 都失败则返回错误

## 📊 命令对比

### 网卡查询命令

| 系统 | 命令 | 输出格式 |
|------|------|----------|
| **Linux** | `ip -o addr show \| grep 'inet ' \| awk '{print $2, $4}'` | `eth0 192.168.1.100/24` |
| **macOS** | `ifconfig \| grep -E '^[a-zA-Z]\|inet [0-9]' \| paste - - \| awk '{print $1, $2}'` | `en0 192.168.1.100` |

### Linux: ip 命令详解

```bash
ip -o addr show | grep 'inet ' | awk '{print $2, $4}'
```

**分解:**
- `ip -o addr show`: 显示所有网络接口（单行格式）
- `grep 'inet '`: 过滤 IPv4 地址
- `awk '{print $2, $4}'`: 提取网卡名和 IP

**示例输出:**
```
lo 127.0.0.1/8
eth0 192.168.1.100/24
docker0 172.17.0.1/16
```

### macOS: ifconfig 命令详解

```bash
ifconfig | grep -E '^[a-zA-Z]|inet [0-9]' | paste - - | awk '{print $1, $2}'
```

**分解:**
- `ifconfig`: 显示网络接口信息
- `grep -E '^[a-zA-Z]|inet [0-9]'`: 匹配接口名和 inet 行
- `paste - -`: 将两行合并为一行
- `awk '{print $1, $2}'`: 提取网卡名和 IP

**示例输出:**
```
en0 192.168.1.100
en1 10.0.0.5
bridge0 172.16.0.1
```

## 🌐 默认网卡选择

```go
if len(interfaces) == 0 {
    osType, err := detectRemoteOS(host, username, password)
    if err != nil {
        interfaces = []string{"any"}  // Linux 默认
    } else if osType == "darwin" {
        interfaces = []string{"en0"}  // macOS 默认
    } else {
        interfaces = []string{"any"}  // Linux 默认
    }
}
```

**选择逻辑:**
- **Linux**: `any` - 捕获所有接口
- **macOS**: `en0` - 主要网络接口（WiFi 或以太网）

## 🔧 tcpdump 命令差异

### Linux vs macOS

```go
// 基本命令相同，但有一些细微差别
command := fmt.Sprintf(
    "sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s 'sudo tcpdump %s -U -w - %s 2>/dev/null' | tshark -T json -r -",
    password, username, host, ifaceArgs, bpfArg,
)
```

**关键参数:**
- `-U`: 包缓冲模式（实时输出）
- `-w -`: 输出到标准输出
- `2>/dev/null`: 重定向错误输出（避免干扰）

### 注意事项

| 特性 | Linux | macOS |
|------|-------|-------|
| tcpdump 位置 | `/usr/bin/tcpdump` | `/usr/sbin/tcpdump` |
| 权限要求 | sudo | sudo + 可能需要 access_bpf |
| any 接口 | ✅ 支持 | ❌ 不支持 |
| 实时监控 | ✅ | ✅ |

## 💻 前端适配

前端代码无需修改，因为后端已经处理了平台差异：

```javascript
// 前端只需发送请求，后端自动处理
const response = await fetch('/api/interfaces?host=...');
// 返回的格式统一，无论后端是 Linux 还是 macOS
```

## 🧪 测试验证

### 测试场景

1. **Linux → Linux**
   - ✅ 使用 ip 命令
   - ✅ 默认网卡 any

2. **Linux → macOS**
   - ✅ 检测到 Darwin
   - ✅ 使用 ifconfig 命令
   - ✅ 默认网卡 en0

3. **macOS → Linux**
   - ✅ 检测到 Linux
   - ✅ 使用 ip 命令
   - ✅ 默认网卡 any

4. **macOS → macOS**
   - ✅ 检测到 Darwin
   - ✅ 使用 ifconfig 命令
   - ✅ 默认网卡 en0

### 测试命令

```bash
# 本地测试（macOS）
./webshark

# 连接远程 Linux
curl "http://localhost:8080/api/interfaces?host=192.168.1.100&username=root&password=xxx"

# 连接远程 macOS
curl "http://localhost:8080/api/interfaces?host=192.168.1.101&username=admin&password=xxx"
```

## ⚠️ 已知限制

### Linux 特有
- `any` 接口在 macOS 上不可用
- `ip` 命令在旧版 Linux 可能不存在

### macOS 特有
- 需要额外权限（access_bpf 组）
- WiFi 抓包可能需要特殊处理
- SIP 可能影响某些操作

### 通用限制
- 都需要 sudo 权限执行 tcpdump
- 都需要安装 tcpdump 和 tshark
- 网络延迟可能影响性能

## 🚀 最佳实践

### 1. 统一的错误处理

```go
interfaces, err = getInterfacesWithIp(host, username, password)
if err == nil && len(interfaces) > 0 {
    return interfaces, nil
}
// 回退到 ifconfig
```

### 2. 明确的错误提示

```go
return nil, fmt.Errorf("failed to get interfaces (tried both ip and ifconfig): %v", err)
```

### 3. 日志记录

```go
log.Printf("Failed to detect remote OS, using default: %v", err)
```

### 4. 文档化

为每个平台创建专门的文档：
- `README.md` - 通用说明
- `MACOS_COMPATIBILITY.md` - macOS 特定配置
- `CROSS_PLATFORM.md` - 技术实现细节

## 📈 未来扩展

### 可能添加的平台支持

1. **Windows (WSL)**
   - 检测: `uname -s` 返回包含 "Microsoft"
   - 命令: 使用 WSL 的 Linux 命令

2. **FreeBSD/OpenBSD**
   - 检测: `uname -s` 返回 "FreeBSD" 或 "OpenBSD"
   - 命令: 使用 ifconfig（与 macOS 类似）

3. **Docker 容器**
   - 检测: 检查 `/.dockerenv` 文件
   - 命令: 根据容器内系统选择

### 改进方向

1. **更智能的默认网卡选择**
   - 检测活动网络连接
   - 优先选择有流量的网卡

2. **缓存检测结果**
   - 避免重复执行 uname
   - 会话级别缓存

3. **用户覆盖选项**
   - 允许手动指定系统类型
   - 允许手动指定默认网卡

## 📝 总结

WebShark 的跨平台兼容性实现特点：

✅ **自动检测**: 无需用户指定系统类型  
✅ **优雅回退**: ip 失败自动尝试 ifconfig  
✅ **统一接口**: 前端无需关心后端平台  
✅ **明确错误**: 清晰的错误提示信息  
✅ **文档完善**: 每个平台都有详细说明  
✅ **易于扩展**: 可轻松添加新平台支持  

这种设计使得 WebShark 能够在不同的 Unix-like 系统上无缝工作，为用户提供一致的体验。
