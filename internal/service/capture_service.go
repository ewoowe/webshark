package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
)

type CaptureSession struct {
	SessionID       string
	Host            string
	Username        string
	Password        string
	Interfaces      []string
	BPFFilter       string
	WiresharkFilter string
	Cmd             *exec.Cmd
	Stdout          io.ReadCloser
	IsActive        bool
}

var (
	sessions     = make(map[string]*CaptureSession)
	sessionsLock sync.RWMutex
	clientMap    = make(map[string]*websocket.Conn)
	clientLock   sync.RWMutex
)

// PacketData 表示一个捕获的数据包
type PacketData struct {
	Timestamp string            `json:"timestamp"`
	Source    string            `json:"source"`
	Dest      string            `json:"dest"`
	Protocol  string            `json:"protocol"`
	Length    int               `json:"length"`
	Info      string            `json:"info"`
	Details   map[string]string `json:"details,omitempty"`
}

func StartCapture(host, username, password string, interfaces []string, bpfFilter, wiresharkFilter string) (string, error) {
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	
	// 如果没有选择网卡，macOS 使用 lo0，Linux 使用 any
	if len(interfaces) == 0 {
		// 先检测系统类型
		osType, err := detectRemoteOS(host, username, password)
		if err != nil {
			log.Printf("Failed to detect remote OS, using default: %v", err)
			interfaces = []string{"any"} // Linux 默认
		} else if osType == "darwin" {
			interfaces = []string{"en0"} // macOS 默认使用 en0（通常是 WiFi 或以太网）
		} else {
			interfaces = []string{"any"}
		}
	}
	
	// 构建 tcpdump 命令
	var ifaceArgs string
	if len(interfaces) == 1 {
		ifaceArgs = fmt.Sprintf("-i %s", interfaces[0])
	} else {
		// 多个网卡需要多个 -i 参数（tcpdump 可能不支持，这里简化处理）
		ifaceArgs = fmt.Sprintf("-i %s", strings.Join(interfaces, " -i "))
	}
	
	bpfArg := ""
	if bpfFilter != "" {
		bpfArg = fmt.Sprintf("'%s'", bpfFilter)
	}
	
	// 检测远程系统类型以优化命令
	osType, _ := detectRemoteOS(host, username, password)
	
	// 尝试配置 sudo 权限（如果失败也不影响，继续执行）
	configureSudoPermissions(host, username, password, osType)
	
	var command string
	if osType == "darwin" {
		// macOS 的 tcpdump 可能需要不同的参数
		// 使用 -l 使 tcpdump 行缓冲，添加 -c 0 不限制包数量
		command = fmt.Sprintf(
			"sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s 'sudo -n tcpdump %s -U -w - %s 2>/dev/null' | tshark -T json -r - -l",
			password, username, host, ifaceArgs, bpfArg,
		)
	} else {
		// Linux
		command = fmt.Sprintf(
			"sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s 'sudo -n tcpdump %s -U -w - %s 2>/dev/null' | tshark -T json -r - -l",
			password, username, host, ifaceArgs, bpfArg,
		)
	}
	
	cmd := exec.Command("bash", "-c", command)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start capture: %v", err)
	}
	
	session := &CaptureSession{
		SessionID:       sessionID,
		Host:            host,
		Username:        username,
		Password:        password,
		Interfaces:      interfaces,
		BPFFilter:       bpfFilter,
		WiresharkFilter: wiresharkFilter,
		Cmd:             cmd,
		Stdout:          stdout,
		IsActive:        true,
	}
	
	sessionsLock.Lock()
	sessions[sessionID] = session
	sessionsLock.Unlock()
	
	// 启动 goroutine 读取并解析数据包
	go readAndParsePackets(session, wiresharkFilter)
	
	return sessionID, nil
}

// detectRemoteOS 检测远程主机的操作系统类型
func detectRemoteOS(host, username, password string) (string, error) {
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", username, host),
		"uname -s")
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to detect OS: %v, stderr: %s", err, stderr.String())
	}
	
	osName := strings.TrimSpace(stdout.String())
	switch osName {
	case "Darwin":
		return "darwin", nil
	case "Linux":
		return "linux", nil
	default:
		return osName, nil
	}
}

// configureSudoPermissions 尝试配置远程主机的 sudo 权限
func configureSudoPermissions(host, username, password, osType string) {
	log.Printf("Attempting to configure sudo permissions for %s@%s", username, host)
	
	// 方法1: 尝试使用 visudo 添加 NOPASSWD 配置（需要已经有 sudo 权限）
	visudoCmd := fmt.Sprintf(
		"echo '%s ALL=(ALL) NOPASSWD: /usr/sbin/tcpdump' | sudo tee -a /etc/sudoers.d/tcpdump && sudo chmod 440 /etc/sudoers.d/tcpdump",
		username,
	)
	
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", username, host),
		visudoCmd)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		log.Printf("Successfully configured sudo NOPASSWD for tcpdump")
		return
	}
	
	log.Printf("Failed to configure sudo via visudo: %v, stderr: %s", err, stderr.String())
	log.Printf("Note: You may need to manually configure sudo on the remote host")
	log.Printf("See TSHARK_JSON_FIX.md for instructions")
}

func StopCapture(sessionID string) error {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()
	
	session, exists := sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	if session.Cmd != nil && session.IsActive {
		if err := session.Cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill capture process: %v", err)
		}
		session.IsActive = false
	}
	
	delete(sessions, sessionID)
	return nil
}

func RegisterClient(sessionID string, conn *websocket.Conn) {
	clientLock.Lock()
	defer clientLock.Unlock()
	clientMap[sessionID] = conn
}

func UnregisterClient(sessionID string) {
	clientLock.Lock()
	defer clientLock.Unlock()
	delete(clientMap, sessionID)
}

func readAndParsePackets(session *CaptureSession, wiresharkFilter string) {
	defer func() {
		session.IsActive = false
		if session.Stdout != nil {
			session.Stdout.Close()
		}
	}()
	
	// tshark -T json 输出的是一个 JSON 数组，需要流式解析
	decoder := json.NewDecoder(session.Stdout)
	
	// 读取开始的 [
	token, err := decoder.Token()
	if err != nil {
		log.Printf("Failed to read JSON start: %v", err)
		return
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		log.Printf("Expected JSON array start '[', got: %v", token)
		return
	}
	
	// 逐个读取数组中的对象
	for decoder.More() {
		var rawPacket json.RawMessage
		if err := decoder.Decode(&rawPacket); err != nil {
			log.Printf("Failed to decode packet: %v", err)
			continue
		}
		
		// 解析数据包字段
		var packet PacketData
		if err := parseTsharkJSON(rawPacket, &packet); err != nil {
			log.Printf("Failed to parse tshark JSON: %v", err)
			continue
		}
		
		// 应用 Wireshark 过滤器（简化版本，实际应该在 tshark 命令中应用）
		if wiresharkFilter != "" && !applyWiresharkFilter(packet, wiresharkFilter) {
			continue
		}
		
		// 设置时间戳
		packet.Timestamp = time.Now().Format("2006-01-02 15:04:05.000000")
		
		// 序列化并发送
		packetJSON, err := json.Marshal(packet)
		if err != nil {
			log.Printf("Failed to marshal packet: %v", err)
			continue
		}
		
		// 通过 WebSocket 推送
		clientLock.RLock()
		conn, exists := clientMap[session.SessionID]
		clientLock.RUnlock()
		
		if exists && conn != nil {
			err := conn.WriteMessage(websocket.TextMessage, packetJSON)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				conn.Close()
			}
		}
	}
	
	if err := decoder.InputOffset(); err != 0 {
		// 检查是否有错误
		log.Printf("Decoder finished")
	}
}

// parseTsharkJSON 解析 tshark 的 JSON 输出格式
func parseTsharkJSON(rawPacket json.RawMessage, packet *PacketData) error {
	// tshark -T json 的输出格式是:
	// {
	//   "_index": "packets-2024-01-01",
	//   "_type": "doc",
	//   "_score": null,
	//   "_source": {
	//     "layers": {
	//       "frame": {...},
	//       "eth": {...},
	//       "ip": {...},
	//       ...
	//     }
	//   }
	// }
	
	var tsharkPacket struct {
		Source struct {
			Layers map[string]json.RawMessage `json:"layers"`
		} `json:"_source"`
	}
	
	if err := json.Unmarshal(rawPacket, &tsharkPacket); err != nil {
		return fmt.Errorf("invalid tshark packet format: %v", err)
	}
	
	layers := tsharkPacket.Source.Layers
	
	// 提取基本信息
	if frameData, ok := layers["frame"]; ok {
		var frame map[string]interface{}
		if err := json.Unmarshal(frameData, &frame); err == nil {
			if ts, ok := frame["frame.time_epoch"].(string); ok {
				packet.Timestamp = ts
			}
			if length, ok := frame["frame.len"].(float64); ok {
				packet.Length = int(length)
			}
		}
	}
	
	// 提取协议信息
	if ipData, ok := layers["ip"]; ok {
		var ip map[string]interface{}
		if err := json.Unmarshal(ipData, &ip); err == nil {
			src, srcOk := ip["ip.src"].(string)
			dst, dstOk := ip["ip.dst"].(string)
			proto, _ := ip["ip.proto"].(string)
			
			if srcOk && dstOk {
				packet.Source = src
				packet.Dest = dst
				
				// 确定协议名称
				switch proto {
				case "6":
					packet.Protocol = "TCP"
				case "17":
					packet.Protocol = "UDP"
				case "1":
					packet.Protocol = "ICMP"
				default:
					packet.Protocol = "IP"
				}
			}
		}
	} else if ethData, ok := layers["eth"]; ok {
		// 如果没有 IP 层，使用以太网层
		var eth map[string]interface{}
		if err := json.Unmarshal(ethData, &eth); err == nil {
			src, srcOk := eth["eth.src"].(string)
			dst, dstOk := eth["eth.dst"].(string)
			if srcOk && dstOk {
				packet.Source = src
				packet.Dest = dst
				packet.Protocol = "ETH"
			}
		}
	}
	
	// 如果还没有设置协议，设置为 UNKNOWN
	if packet.Protocol == "" {
		packet.Protocol = "UNKNOWN"
	}
	
	return nil
}

// applyWiresharkFilter 简化的 Wireshark 过滤器实现
func applyWiresharkFilter(packet PacketData, filter string) bool {
	// 这是一个简化版本，实际应该使用 tshark 的显示过滤器
	// 这里只做基本的字符串匹配
	filterLower := strings.ToLower(filter)
	
	if strings.Contains(filterLower, "ip.addr") {
		// 提取 IP 地址进行匹配
		parts := strings.Split(filterLower, "==")
		if len(parts) == 2 {
			targetIP := strings.TrimSpace(parts[1])
			if strings.Contains(packet.Source, targetIP) || strings.Contains(packet.Dest, targetIP) {
				return true
			}
			return false
		}
	}
	
	if strings.Contains(filterLower, "tcp.port") {
		parts := strings.Split(filterLower, "==")
		if len(parts) == 2 {
			targetPort := strings.TrimSpace(parts[1])
			if strings.Contains(packet.Source, ":"+targetPort) || strings.Contains(packet.Dest, ":"+targetPort) {
				return true
			}
			return false
		}
	}
	
	// 协议过滤
	if strings.Contains(filterLower, packet.Protocol) {
		return true
	}
	
	return true // 默认不过滤
}
