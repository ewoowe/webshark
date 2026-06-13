package service

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"webshark/internal/config"
)

// GenerateCapturePaths 生成抓包文件的存储路径
// 参数:
//   - taskID: 任务ID（用于生成唯一文件名）
//
// 返回:
//   - pcapPath: PCAP文件完整路径
//   - err: 错误信息
func GenerateCapturePaths(taskID int64) (pcapPath string, err error) {
	// 获取配置中的目录
	pcapDir := config.GetCapturePcapDir()

	// 确保目录存在
	if err := ensureDirectory(pcapDir); err != nil {
		return "", fmt.Errorf("failed to create pcap directory: %w", err)
	}

	// 生成时间戳
	timestamp := time.Now().Format("20060102_150405")

	// 生成文件名
	// PCAP 文件: {taskID}_{timestamp}.pcap
	pcapFilename := fmt.Sprintf("%d_%s.pcap", taskID, timestamp)
	pcapPath = filepath.Join(pcapDir, pcapFilename)

	return pcapPath, nil
}

// CreateFIFO 创建 FIFO 命名管道文件
// 如果文件已存在，先删除再创建
func CreateFIFO(fifoPath string) error {
	// 如果文件已存在，先删除
	if _, err := os.Stat(fifoPath); err == nil {
		if err := os.Remove(fifoPath); err != nil {
			return fmt.Errorf("failed to remove existing fifo %s: %w", fifoPath, err)
		}
	}

	// 创建 FIFO（权限：所有者读写，组和其他人只读）
	if err := syscall.Mkfifo(fifoPath, 0644); err != nil {
		return fmt.Errorf("failed to create fifo %s: %w", fifoPath, err)
	}

	return nil
}

// ensureDirectory 确保目录存在，如果不存在则创建
func ensureDirectory(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 创建目录（包括父目录）
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
