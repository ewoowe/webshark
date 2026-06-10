package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"webshark/internal/config"
)

// GenerateCapturePaths 生成抓包文件的存储路径
// 参数:
//   - taskID: 任务ID（用于生成唯一文件名）
//
// 返回:
//   - pcapPath: PCAP文件完整路径
//   - fifoPath: FIFO命名管道完整路径
//   - err: 错误信息
func GenerateCapturePaths(taskID int64) (pcapPath, fifoPath string, err error) {
	// 获取配置中的目录
	pcapDir := config.GetCapturePcapDir()
	fifoDir := config.GetCaptureFifoDir()

	// 确保目录存在
	if err := ensureDirectory(pcapDir); err != nil {
		return "", "", fmt.Errorf("failed to create pcap directory: %w", err)
	}
	if err := ensureDirectory(fifoDir); err != nil {
		return "", "", fmt.Errorf("failed to create fifo directory: %w", err)
	}

	// 生成时间戳
	timestamp := time.Now().Format("20060102_150405")

	// 生成文件名
	// PCAP 文件: {taskID}_{timestamp}.pcap
	pcapFilename := fmt.Sprintf("%d_%s.pcap", taskID, timestamp)
	pcapPath = filepath.Join(pcapDir, pcapFilename)

	// FIFO 文件: {taskID}_{timestamp}.fifo
	fifoFilename := fmt.Sprintf("%d_%s.fifo", taskID, timestamp)
	fifoPath = filepath.Join(fifoDir, fifoFilename)

	return pcapPath, fifoPath, nil
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
