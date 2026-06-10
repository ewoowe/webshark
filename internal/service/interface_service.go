package service

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type NetworkInterface struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
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

// GetRemoteInterfaces 通过 SSH 获取远程主机的网络接口信息
func GetRemoteInterfaces(host, username, password string) ([]NetworkInterface, error) {
	// 1. 先探测远程主机的操作系统类型
	osType, err := detectRemoteOS(host, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to detect remote OS: %v", err)
	}

	// 2. 根据操作系统类型选择对应的命令
	switch osType {
	case "linux":
		// Linux 系统使用 ip 命令
		interfaces, err := getInterfacesWithIp(host, username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to get interfaces on Linux: %v", err)
		}
		return interfaces, nil
	case "darwin":
		// macOS/BSD 系统使用 ifconfig 命令
		interfaces, err := getInterfacesWithIfconfig(host, username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to get interfaces on macOS/BSD: %v", err)
		}
		return interfaces, nil
	default:
		// 未知系统，尝试两种方法
		interfaces, err := getInterfacesWithIp(host, username, password)
		if err == nil && len(interfaces) > 0 {
			return interfaces, nil
		}

		interfaces, err = getInterfacesWithIfconfig(host, username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to get interfaces (tried both ip and ifconfig): %v", err)
		}
		return interfaces, nil
	}
}

// getInterfacesWithIp 使用 ip 命令获取网卡信息（Linux）
func getInterfacesWithIp(host, username, password string) ([]NetworkInterface, error) {
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", username, host),
		"ip -o addr show | grep 'inet ' | awk '{print $2, $4}'")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ip command failed: %v, stderr: %s", err, stderr.String())
	}

	var interfaces []NetworkInterface
	lines := strings.Split(stdout.String(), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			iface := NetworkInterface{
				Name: parts[0],
				IP:   parts[1],
			}
			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}

// getInterfacesWithIfconfig 使用 ifconfig 命令获取网卡信息（macOS/BSD）
func getInterfacesWithIfconfig(host, username, password string) ([]NetworkInterface, error) {
	// 使用简单的 shell 脚本解析 ifconfig 输出 - 获取所有网卡,包括回环地址
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", username, host),
		"ifconfig -l | tr ' ' '\\n' | while read iface; do ip=$(ifconfig $iface | grep 'inet ' | awk '{print $2}' | head -1); echo \"$iface ${ip:-}\"; done")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ifconfig command failed: %v, stderr: %s", err, stderr.String())
	}

	var interfaces []NetworkInterface
	lines := strings.Split(stdout.String(), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 1 {
			name := parts[0]
			ip := ""
			if len(parts) >= 2 {
				ip = parts[1]
			}

			iface := NetworkInterface{
				Name: name,
				IP:   ip,
			}
			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}
