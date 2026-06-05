# tshark JSON 输出问题修复

## 🐛 问题描述

执行以下命令时返回空数组 `[]`：

```bash
sshpass -p 'password' ssh -o StrictHostKeyChecking=no user@host \
  'sudo tcpdump -U -w - 2>/dev/null' | tshark -T json -r -
```

### 原因分析

1. **tshark JSON 格式问题**
   - `tshark -T json` 输出的是一个完整的 JSON 数组，而不是每行一个 JSON 对象
   - 输出格式：
     ```json
     [
       {"_source": {...}},
       {"_source": {...}},
       ...
     ]
     ```
   - 之前的代码试图逐行解析，但每行不是独立的 JSON

2. **缓冲问题**
   - tshark 默认可能缓冲输出，导致数据包不会立即显示
   - 需要添加 `-l` 参数启用行缓冲

3. **tcpdump 权限问题**
   - sudo 可能需要密码输入
   - 确保远程用户的 sudoers 配置正确

## ✅ 解决方案

### 1. 添加行缓冲参数

在 tshark 命令中添加 `-l` 参数：

```bash
# 之前
tshark -T json -r -

# 现在
tshark -T json -r - -l
```

### 2. 使用流式 JSON 解析

修改 Go 代码，使用 `json.Decoder` 进行流式解析：

```go
// 之前：逐行读取
scanner := bufio.NewScanner(session.Stdout)
for scanner.Scan() {
    line := scanner.Text()
    // 尝试解析每行为 JSON...
}

// 现在：流式解析 JSON 数组
decoder := json.NewDecoder(session.Stdout)

// 读取开始的 '['
token, _ := decoder.Token()

// 逐个读取数组中的对象
for decoder.More() {
    var rawPacket json.RawMessage
    decoder.Decode(&rawPacket)
    // 解析数据包...
}
```

### 3. 解析 tshark JSON 格式

添加专门的解析函数处理 tshark 的输出结构：

```go
func parseTsharkJSON(rawPacket json.RawMessage, packet *PacketData) error {
    // tshark 输出格式：
    // {
    //   "_source": {
    //     "layers": {
    //       "frame": {...},
    //       "ip": {...},
    //       "tcp": {...}
    //     }
    //   }
    // }
    
    var tsharkPacket struct {
        Source struct {
            Layers map[string]json.RawMessage `json:"layers"`
        } `json:"_source"`
    }
    
    json.Unmarshal(rawPacket, &tsharkPacket)
    
    // 从 layers 中提取信息
    layers := tsharkPacket.Source.Layers
    
    // 提取帧信息（时间戳、长度）
    if frameData, ok := layers["frame"]; ok {
        // 解析 frame 层...
    }
    
    // 提取 IP 信息（源地址、目标地址、协议）
    if ipData, ok := layers["ip"]; ok {
        // 解析 IP 层...
    }
    
    return nil
}
```

## 🔍 tshark JSON 输出格式详解

### 完整示例

```json
{
  "_index": "packets-2024-06-04",
  "_type": "doc",
  "_score": null,
  "_source": {
    "layers": {
      "frame": {
        "frame.time_epoch": "1717500000.123456",
        "frame.len": "128",
        "frame.cap_len": "128"
      },
      "eth": {
        "eth.src": "aa:bb:cc:dd:ee:ff",
        "eth.dst": "11:22:33:44:55:66",
        "eth.type": "IPv4 (0x0800)"
      },
      "ip": {
        "ip.src": "192.168.1.100",
        "ip.dst": "192.168.1.1",
        "ip.proto": "6",
        "ip.len": "114"
      },
      "tcp": {
        "tcp.srcport": "54321",
        "tcp.dstport": "80",
        "tcp.len": "74"
      }
    }
  }
}
```

### 关键字段

| 字段路径 | 说明 | 示例值 |
|---------|------|--------|
| `_source.layers.frame.frame.time_epoch` | 时间戳 | `"1717500000.123456"` |
| `_source.layers.frame.frame.len` | 包长度 | `"128"` |
| `_source.layers.ip.ip.src` | 源 IP | `"192.168.1.100"` |
| `_source.layers.ip.ip.dst` | 目标 IP | `"192.168.1.1"` |
| `_source.layers.ip.ip.proto` | 协议号 | `"6"` (TCP), `"17"` (UDP) |
| `_source.layers.tcp.tcp.srcport` | 源端口 | `"54321"` |
| `_source.layers.tcp.tcp.dstport` | 目标端口 | `"80"` |

## 📝 优化后的命令

### macOS

```bash
sshpass -p '<password>' ssh -o StrictHostKeyChecking=no <user>@<host> \
  'sudo tcpdump -i en0 -U -w - 2>/dev/null' | \
  tshark -T json -r - -l
```

### Linux

```bash
sshpass -p '<password>' ssh -o StrictHostKeyChecking=no <user>@<host> \
  'sudo tcpdump -i any -U -w - 2>/dev/null' | \
  tshark -T json -r - -l
```

### 关键参数说明

- **tcpdump**:
  - `-i`: 指定网卡接口
  - `-U`: 使输出无缓冲（packet buffered mode）
  - `-w -`: 以 pcap 格式写入 stdout
  - `2>/dev/null`: 丢弃错误输出

- **tshark**:
  - `-T json`: 输出 JSON 格式
  - `-r -`: 从 stdin 读取 pcap 数据
  - `-l`: 启用行缓冲模式（关键！）

## 🔧 调试技巧

### 1. 测试 SSH 连接

```bash
sshpass -p 'password' ssh -o StrictHostKeyChecking=no user@host "echo 'SSH works'"
```

### 2. 测试 tcpdump

```bash
# 本地测试
sudo tcpdump -i any -c 1 -w /tmp/test.pcap

# 远程测试
sshpass -p 'password' ssh user@host "sudo tcpdump -i any -c 1 -w -" > /tmp/test.pcap
```

### 3. 测试 tshark

```bash
# 直接读取 pcap 文件
tshark -T json -r /tmp/test.pcap | head -50

# 从管道读取
cat /tmp/test.pcap | tshark -T json -r - -l | head -50
```

### 4. 检查 sudo 权限

```bash
# 确保远程用户可以无密码运行 tcpdump
sshpass -p 'password' ssh user@host "sudo -n tcpdump --version"
```

如果需要密码，编辑 sudoers：

```bash
sudo visudo
# 添加：
username ALL=(ALL) NOPASSWD: /usr/sbin/tcpdump
```

## ⚠️ 常见问题

### Q1: 返回空数组 `[]`

**可能原因：**
1. tshark 没有收到任何数据包
2. 缓冲问题导致输出延迟
3. tcpdump 过滤器太严格

**解决方法：**
```bash
# 1. 添加 -l 参数
tshark -T json -r - -l

# 2. 移除 BPF 过滤器测试
sudo tcpdump -i any -U -w -

# 3. 等待更长时间捕获数据包
```

### Q2: JSON 解析错误

**可能原因：**
1. tshark 输出了非 JSON 内容（警告、错误信息）
2. 输出被截断

**解决方法：**
```bash
# 1. 重定向 stderr
sudo tcpdump ... 2>/dev/null | tshark ...

# 2. 检查 tshark 版本
tshark --version

# 3. 使用 Wireshark 3.x 以上版本
```

### Q3: 权限被拒绝

**可能原因：**
1. sudo 需要密码
2. 用户不在 wheel/admin 组

**解决方法：**
```bash
# 方法1: 配置 sudoers（推荐）
sudo visudo
# 添加：username ALL=(ALL) NOPASSWD: /usr/sbin/tcpdump

# 方法2: 设置 cap_net_raw 能力（Linux）
sudo setcap cap_net_raw,cap_net_admin=eip /usr/bin/tcpdump
```

## 📊 性能优化

### 1. 限制捕获的数据包大小

```bash
# 只捕获前 128 字节
sudo tcpdump -s 128 -U -w -
```

### 2. 使用 BPF 过滤器减少数据量

```bash
# 只捕获 HTTP 流量
sudo tcpdump 'tcp port 80' -U -w -

# 排除 SSH 流量
sudo tcpdump 'not port 22' -U -w -
```

### 3. 在 tshark 层面过滤

```bash
# 只显示特定协议
tshark -T json -r - -Y "http or dns"
```

## ✨ 改进总结

| 项目 | 之前 | 现在 |
|------|------|------|
| **JSON 解析** | ❌ 逐行解析 | ✅ 流式解析 |
| **缓冲处理** | ❌ 可能阻塞 | ✅ 行缓冲模式 |
| **错误处理** | ⚠️ 简单 | ✅ 完善的错误日志 |
| **数据提取** | ⚠️ 假设格式 | ✅ 正确解析 tshark 结构 |
| **兼容性** | ⚠️ 有限 | ✅ 支持多种协议层 |

---

**更新时间**: 2026-06-04  
**版本**: v1.2  
**作者**: WebShark Team
