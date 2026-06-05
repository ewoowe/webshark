# WebShark 快速开始指南

## 🚀 5分钟快速上手

### 第一步：检查系统依赖

运行启动脚本会自动检查依赖：

```bash
chmod +x start.sh
./start.sh
```

如果缺少依赖，请根据提示安装：

**macOS:**
```bash
brew install sshpass wireshark
```

**Ubuntu/Debian:**
```bash
sudo apt-get install sshpass tshark tcpdump
```

> **注意**: WebShark 现已支持 Linux 和 macOS 系统，会自动检测远程主机类型并使用相应命令。

### 第二步：启动服务

```bash
./webshark
```

或

```bash
go run main.go
```

服务将在 http://localhost:8080 启动

### 第三步：访问 Web 界面

打开浏览器访问：http://localhost:8080

### 第四步：连接远程主机

1. **输入连接信息：**
   - 主机地址：例如 `192.168.1.100`
   - 用户名：例如 `root`
   - 密码：输入 SSH 密码

2. **点击"获取网卡列表"**

### 第五步：选择网卡

- 勾选需要抓包的网卡（可多选）
- 不选择则默认抓取所有网卡

### 第六步：配置过滤器（可选）

**BPF 过滤器示例：**
- `tcp port 80` - 只抓取 HTTP 流量
- `host 192.168.1.1` - 只抓取特定主机
- `not port 22` - 排除 SSH 流量

**Wireshark 过滤器示例：**
- `http` - 只显示 HTTP 协议
- `dns` - 只显示 DNS 协议
- `ip.addr == 192.168.1.1` - 特定 IP

### 第七步：开始抓包

点击"开始抓包"按钮，数据包会实时显示在表格中。

### 第八步：查看数据

- 点击任意数据包查看详情
- 实时查看数据包统计
- 随时停止或清空

---

## 💡 常用命令速查

### 编译
```bash
go build -o webshark
```

### 运行
```bash
./webshark
```

### 清理
```bash
go clean
```

### 更新依赖
```bash
go mod tidy
```

---

## 🔧 常见问题

### Q1: 提示 "sshpass: command not found"
**A:** 安装 sshpass
```bash
# macOS
brew install sshpass

# Ubuntu
sudo apt-get install sshpass
```

### Q2: 提示 "tshark: command not found"
**A:** 安装 Wireshark（包含 tshark）
```bash
# macOS
brew install wireshark

# Ubuntu
sudo apt-get install tshark
```

### Q3: 无法获取网卡列表
**A:** 检查以下几点：
1. 远程主机 SSH 是否可达
2. 用户名和密码是否正确
3. 远程主机是否安装了 tcpdump
4. 用户是否有 sudo 权限

### Q4: 抓包没有数据
**A:** 可能的原因：
1. 网卡选择错误
2. BPF 过滤器太严格
3. 网络流量太少
4. 尝试去掉过滤器重新抓包

### Q5: WebSocket 连接失败
**A:** 
1. 检查浏览器是否支持 WebSocket
2. 确保防火墙没有阻止
3. 刷新页面重试

---

## 📚 更多信息

- 详细文档：[README.md](README.md)
- 项目结构：[PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
- 实现总结：[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)

---

## ⚠️ 注意事项

1. **安全性**：当前版本密码以明文传递，仅用于测试环境
2. **权限**：远程用户需要 sudo 权限执行 tcpdump
3. **性能**：高流量环境建议使用过滤器减少数据量
4. **浏览器**：需要使用现代浏览器（Chrome、Firefox、Safari）

---

祝你使用愉快！🎉
