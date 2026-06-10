package service

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"webshark/internal/config"
	"webshark/internal/gorm"
	"webshark/internal/logger"
	"webshark/internal/utils"

	"go.uber.org/zap"
)

// packetNoCounter 任务组级别的包序号计数器（使用原子操作保证线程安全）
var packetNoCounter = struct {
	counters map[int64]*atomic.Int64
	sync.RWMutex
}{
	counters: make(map[int64]*atomic.Int64),
}

// packetDetailCache 数据包详情缓存（用于处理时序问题）
type packetDetailKey struct {
	TaskID      int64
	FrameNumber int64
}

var packetDetailCache = struct {
	cache map[packetDetailKey]string
	sync.RWMutex
}{
	cache: make(map[packetDetailKey]string),
}

// cachePacketDetail 缓存数据包详情
func cachePacketDetail(taskID, frameNumber int64, content string) {
	packetDetailCache.Lock()
	defer packetDetailCache.Unlock()
	key := packetDetailKey{TaskID: taskID, FrameNumber: frameNumber}
	packetDetailCache.cache[key] = content
}

// getAndRemoveCachedDetail 获取并删除缓存的数据包详情
func getAndRemoveCachedDetail(taskID, frameNumber int64) (string, bool) {
	packetDetailCache.Lock()
	defer packetDetailCache.Unlock()
	key := packetDetailKey{TaskID: taskID, FrameNumber: frameNumber}
	content, exists := packetDetailCache.cache[key]
	if exists {
		delete(packetDetailCache.cache, key)
	}
	return content, exists
}

// retryCachedDetails 定期重试缓存中的详情更新
func retryCachedDetails() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		packetDetailCache.RLock()
		keys := make([]packetDetailKey, 0, len(packetDetailCache.cache))
		for k := range packetDetailCache.cache {
			keys = append(keys, k)
		}
		packetDetailCache.RUnlock()

		for _, key := range keys {
			compressedContent, exists := getAndRemoveCachedDetail(key.TaskID, key.FrameNumber)
			if !exists {
				continue
			}

			// 尝试更新（缓存中已是压缩后的数据）
			if err := updatePacketContentWithCompressed(key.TaskID, key.FrameNumber, compressedContent); err != nil {
				logger.Debug("Retry update packet content failed, re-caching",
					zap.Int64("taskID", key.TaskID),
					zap.Int64("frameNumber", key.FrameNumber),
					zap.Error(err))
				// 重新放回缓存
				cachePacketDetail(key.TaskID, key.FrameNumber, compressedContent)
			} else {
				logger.Debug("Retry update packet content succeeded",
					zap.Int64("taskID", key.TaskID),
					zap.Int64("frameNumber", key.FrameNumber))
			}
		}
	}
}

func init() {
	// 启动定期重试缓存中数据包详情的后台协程
	go retryCachedDetails()
}

// getOrCreateCounter 获取或创建任务组的计数器
func getOrCreateCounter(taskGroupID int64) *atomic.Int64 {
	// 先尝试读取
	packetNoCounter.RLock()
	if counter, exists := packetNoCounter.counters[taskGroupID]; exists {
		packetNoCounter.RUnlock()
		return counter
	}
	packetNoCounter.RUnlock()

	// 需要创建新的计数器
	packetNoCounter.Lock()
	defer packetNoCounter.Unlock()

	// 双重检查，避免并发创建
	if counter, exists := packetNoCounter.counters[taskGroupID]; exists {
		return counter
	}

	counter := &atomic.Int64{}
	counter.Store(0)
	packetNoCounter.counters[taskGroupID] = counter
	return counter
}

type CaptureRequest struct {
	TaskName     string        `json:"taskName" binding:"required" label:"任务名称"`            // 任务名称
	OnlyCapture  bool          `json:"onlyCapture"`                                         // 是否只捕获数据包，不解析
	ParseDetail  bool          `json:"parseDetail"`                                         // 是否解析数据包详情
	DetailFormat string        `json:"detailFormat"`                                        // 数据包详情格式
	HostCaptures []HostCapture `json:"hostCaptures" binding:"required,dive" label:"主机抓包配置"` // 所有目标主机的抓包配置
}

type HostCapture struct {
	HostID   int                 `json:"hostId" binding:"required" label:"目标主机ID"`        // 目标主机ID
	Captures []HostSingleCapture `json:"captures" binding:"required,dive" label:"抓包任务列表"` // 多组抓包任务的配置
}

type HostSingleCapture struct {
	StreamID        int      `json:"streamId" binding:"required" label:"抓包流序号"` // 抓包流序号
	Interfaces      []string `json:"interfaces"`                                // 本次抓包流所在的网卡
	BPFFilter       string   `json:"bpfFilter"`                                 // BPF 过滤器
	WiresharkFilter string   `json:"wiresharkFilter"`                           // Wireshark 过滤器
}

type TaskInfo struct {
	TaskGroupID int64         `json:"taskGroupId"` // 任务组ID，如果不是批量任务，则返回负数
	TaskIDs     map[int]int64 `json:"taskIds"`     // 抓包流序号与taskId映射关系
}

func StartCapture(captureRequest CaptureRequest) (*TaskInfo, error) {
	var taskInfo TaskInfo
	isTaskGroup := false
	if len(captureRequest.HostCaptures) > 1 {
		isTaskGroup = true
	} else if len(captureRequest.HostCaptures[0].Captures) > 1 {
		isTaskGroup = true
	}

	if isTaskGroup {
		taskGroup := gorm.TaskGroup{
			TaskGroupName: captureRequest.TaskName,
		}
		err := gorm.Repo.CreateTaskGroup(&taskGroup)
		if err != nil {
			return nil, fmt.Errorf("failed to create task group: %v", err)
		}
		taskInfo.TaskGroupID = taskGroup.ID
		taskInfo.TaskIDs = make(map[int]int64)

		errOccurred := false
		for _, hostCapture := range captureRequest.HostCaptures {
			streamIdMapTask, err := startHostCaptures(taskGroup.ID, captureRequest.TaskName, captureRequest.OnlyCapture, captureRequest.ParseDetail, captureRequest.DetailFormat, hostCapture)
			if err != nil {
				logger.Error("Failed to start host captures", zap.Error(err))
				// 继续执行其他主机的任务
				errOccurred = true
			}
			maps.Copy(taskInfo.TaskIDs, streamIdMapTask)
		}

		if errOccurred {
			return &taskInfo, fmt.Errorf("some errors occurred while starting host captures")
		}

		return &taskInfo, nil
	}

	taskInfo.TaskIDs = make(map[int]int64)
	streamIdMapTask, err := startHostCaptures(-1, captureRequest.TaskName, captureRequest.OnlyCapture, captureRequest.ParseDetail, captureRequest.DetailFormat, captureRequest.HostCaptures[0])
	if err != nil {
		return nil, fmt.Errorf("failed to start host captures: %v", err)
	}
	maps.Copy(taskInfo.TaskIDs, streamIdMapTask)
	return &taskInfo, nil
}

// startHostCaptures 启动单个主机上的多个抓包任务
func startHostCaptures(taskGroupID int64, taskName string, onlyCapture bool, parseDetail bool, detailFormat string, hostCapture HostCapture) (map[int]int64, error) {
	streamIdMapTask := make(map[int]int64)

	// 获取主机信息
	host, err := GetHostByID(int64(hostCapture.HostID))
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %v", err)
	}

	// 为该主机上的每个抓包流启动任务
	for _, capture := range hostCapture.Captures {
		taskID, err := startSingleCapture(taskGroupID, taskName, onlyCapture, parseDetail, detailFormat, host, capture)
		if err != nil {
			logger.Error("Failed to start single capture",
				zap.String("host", host.IP),
				zap.Int("streamId", capture.StreamID),
				zap.Error(err))
			continue
		}
		streamIdMapTask[capture.StreamID] = taskID
	}

	return streamIdMapTask, nil
}

// startSingleCapture 启动单个抓包任务，返回任务ID
func startSingleCapture(taskGroupID int64, taskName string, onlyCapture bool, parseDetail bool, detailFormat string, host *gorm.Host, capture HostSingleCapture) (int64, error) {
	// 1. 先创建任务记录（不包含文件路径）
	task := gorm.Task{
		TaskName:        taskName,
		StreamID:        int8(capture.StreamID),
		HostID:          host.ID,
		Interfaces:      capture.Interfaces,
		OnlyCapture:     onlyCapture,
		ParseDetail:     parseDetail,
		DetailFormat:    detailFormat,
		BpfFilter:       capture.BPFFilter,
		WiresharkFilter: capture.WiresharkFilter,
		TaskGroupId:     taskGroupID,
		Status:          "created",
		CreatedAt:       time.Now(), // 显式设置创建时间
	}

	// 如果 TaskGroupID 为负数，说明不是任务组
	if taskGroupID < 0 {
		task.TaskGroupId = 0
	}

	err := gorm.Repo.CreateTask(&task)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %v", err)
	}

	// 2. 使用任务ID生成文件路径
	pcapPath, fifoPath, err := GenerateCapturePaths(task.ID)
	if err != nil {
		return task.ID, fmt.Errorf("failed to generate paths: %v", err)
	}

	// 3. 更新任务的文件路径
	task.FilePath = pcapPath
	task.FifoPath = fifoPath
	err = gorm.Repo.UpdateTask(&task)
	if err != nil {
		return task.ID, fmt.Errorf("failed to update task file paths: %v", err)
	}

	// 4. 构建并执行抓包命令
	cmd, fullCommand, err := buildCaptureCommand(host, capture, pcapPath, fifoPath, onlyCapture, parseDetail, detailFormat)
	if err != nil {
		return task.ID, fmt.Errorf("failed to build command: %v", err)
	}

	// 4.1 保存完整命令到数据库
	task.FullCommand = fullCommand
	err = gorm.Repo.UpdateTask(&task)
	if err != nil {
		logger.Warn("Failed to update task full command", zap.Error(err))
	}

	// 5. 启动进程前先设置stderr和stdout捕获
	var stderr strings.Builder
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Warn("Failed to create stderr pipe", zap.Error(err))
	} else {
		// 启动一个协程读取stderr
		go func() {
			buf := make([]byte, 4096)
			for {
				n, readErr := stderrPipe.Read(buf)
				if n > 0 {
					stderr.Write(buf[:n])
				}
				if readErr != nil {
					break
				}
			}
		}()
	}

	// 如果需要解析概览，先获取stdout pipe（必须在Start之前）
	var stdout io.ReadCloser
	if !onlyCapture {
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			logger.Error("Failed to create stdout pipe", zap.Error(err))
			return task.ID, fmt.Errorf("failed to create stdout pipe: %v", err)
		}
	}

	if err := cmd.Start(); err != nil {
		task.Status = "failed"
		task.Message = fmt.Sprintf("Failed to start process: %v", err)
		gorm.Repo.UpdateTask(&task)
		return task.ID, fmt.Errorf("failed to start process: %v", err)
	}

	// 6. 保存进程信息（包含完整命令）
	process := gorm.Process{
		TaskID: task.ID,
		Pid:    int64(cmd.Process.Pid),
		Ppid:   getProcessPpidByProc(int(cmd.Process.Pid)), // 使用方法1：从 /proc 读取
		// Ppid: getProcessPpidByPs(int(cmd.Process.Pid)),  // 或者方法2：使用 ps 命令
		Type:    "sshpass",
		Command: cmd.String(),
		Alive:   true,
	}
	err = gorm.Repo.CreateProcess(&process)
	if err != nil {
		logger.Error("Failed to create process record", zap.Error(err))
	}

	// 7. 启动协程监控进程状态并传递stderr
	go monitorProcessStatusWithStderr(task.ID, cmd, &stderr)

	// 8. 如果需要解析概览，启动协程实时解析输出（传递已获取的stdout）
	if !onlyCapture {
		go parseOverviewFromOutput(task.ID, task.TaskGroupId, stdout, parseDetail, detailFormat, fifoPath, pcapPath)
	}
	logger.Info("Capture started",
		zap.Int64("taskId", task.ID),
		zap.String("host", host.IP),
		zap.String("pcapPath", pcapPath))

	return task.ID, nil
}

// getInterfaceName 获取网卡名称，如果为空则返回 "any"
func getInterfaceName(interfaces []string) string {
	if len(interfaces) == 0 {
		return "any"
	}
	return interfaces[0]
}

// buildCaptureCommand 构建抓包命令
// 命令结构:
//
//	sshpass -p 'password' ssh user@host 'tcpdump -i eth0 -U -w - bpf_filter' | \
//	  tee output.pcap fifo | \
//	  (如果需要解析: tshark -T json -r - > fifo)
func buildCaptureCommand(host *gorm.Host, capture HostSingleCapture, pcapPath, fifoPath string, onlyCapture bool, parseDetail bool, detailFormat string) (*exec.Cmd, string, error) {
	// 1. 构建 tcpdump 命令
	var interfaceArgs []string
	if len(capture.Interfaces) == 0 {
		interfaceArgs = append(interfaceArgs, "-i any")
	} else {
		// 为每个网卡添加 -i 参数
		for _, iface := range capture.Interfaces {
			interfaceArgs = append(interfaceArgs, fmt.Sprintf("-i %s", iface))
		}
	}
	interfaceArg := strings.Join(interfaceArgs, " ")

	bpfFilter := ""
	if capture.BPFFilter != "" {
		bpfFilter = fmt.Sprintf("'%s'", capture.BPFFilter)
	}

	// tcpdump 命令：使用 -U 使输出无缓冲，-w - 输出到 stdout
	// 注意：不要重定向 stderr，以便捕获错误信息
	// 使用 sudo 提升权限以允许抓包
	tcpdumpCmd := fmt.Sprintf("sudo tcpdump %s -U -w - %s", interfaceArg, bpfFilter)

	// 2. 构建 ssh 命令
	sshCmd := fmt.Sprintf("sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s '%s'",
		host.Password, host.UserName, host.IP, tcpdumpCmd)

	// 3. 根据是否需要解析，构建完整的管道命令
	var fullCommand string

	if onlyCapture {
		// 只抓包，不解析：直接保存到 pcap 文件
		fullCommand = fmt.Sprintf("%s > %s", sshCmd, pcapPath)
	} else if !parseDetail {
		// 解析概览：使用 tee 保存 pcap，同时用 tshark 实时解析
		// ssh ... | tee pcap | tshark -l -i - -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info
		fullCommand = fmt.Sprintf("%s | tee %s | tshark -l -i - -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info",
			sshCmd, pcapPath)
	} else {
		// 解析详情：使用 tee 保存 pcap，同时复制流量到 fifo
		// ssh ... | tee pcap fifo | tshark -l -i - -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info
		// 然后由另一个 tshark 进程读取 fifo
		fullCommand = fmt.Sprintf("%s | tee %s %s | tshark -l -i - -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info",
			sshCmd, pcapPath, fifoPath)
	}

	cmd := exec.Command("bash", "-c", fullCommand)

	// 设置进程组，以便后续可以一起终止所有子进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 返回命令和完整命令字符串
	return cmd, fullCommand, nil
}

// monitorProcessStatus 监控进程状态，等待进程结束并更新数据库
func monitorProcessStatus(taskID int64, cmd *exec.Cmd) {
	// 等待命令执行完成
	err := cmd.Wait()

	// 更新任务状态
	task, _ := gorm.Repo.GetTaskByID(taskID)
	if err != nil {
		task.Status = "stopped"
		now := time.Now()
		task.StopAt = &now
		task.Message = fmt.Sprintf("Process exited: %v", err)
	} else {
		task.Status = "stopped"
		now := time.Now()
		task.StopAt = &now
		task.Message = "Normal exit"
	}
	gorm.Repo.UpdateTask(task)

	// 更新进程状态
	processes, err := gorm.Repo.ListProcessesByTaskID(taskID)
	if err == nil && len(processes) > 0 {
		processes[0].Alive = false
		gorm.Repo.UpdateProcess(processes[0])
	}

	logger.Info("Capture process stopped",
		zap.Int64("taskID", taskID),
		zap.String("status", task.Status),
		zap.String("message", task.Message),
		zap.Duration("duration", time.Since(task.CreatedAt)))
}

// monitorProcessStatusWithStderr 监控进程状态并捕获stderr输出
func monitorProcessStatusWithStderr(taskID int64, cmd *exec.Cmd, stderr *strings.Builder) {
	// 等待命令执行完成
	err := cmd.Wait()

	// 更新任务状态
	task, _ := gorm.Repo.GetTaskByID(taskID)
	if err != nil {
		task.Status = "failed"
		now := time.Now()
		task.StopAt = &now
		// 包含stderr输出以便调试
		errMsg := fmt.Sprintf("Process exited: %v", err)
		if stderr != nil && stderr.Len() > 0 {
			errMsg += fmt.Sprintf(", stderr: %s", strings.TrimSpace(stderr.String()))
		}
		task.Message = errMsg
	} else {
		task.Status = "stopped"
		now := time.Now()
		task.StopAt = &now
		if stderr != nil && stderr.Len() > 0 {
			task.Message = fmt.Sprintf("Normal exit with warnings: %s", strings.TrimSpace(stderr.String()))
		} else {
			task.Message = "Normal exit"
		}
	}
	gorm.Repo.UpdateTask(task)

	// 更新进程状态
	processes, listErr := gorm.Repo.ListProcessesByTaskID(taskID)
	if listErr == nil && len(processes) > 0 {
		processes[0].Alive = false
		gorm.Repo.UpdateProcess(processes[0])
	}

	logger.Info("Capture process stopped",
		zap.Int64("taskID", taskID),
		zap.String("status", task.Status),
		zap.String("message", task.Message),
		zap.Duration("duration", time.Since(task.CreatedAt)))
}

// parseOverviewFromOutput 从进程输出实时解析数据包概览
func parseOverviewFromOutput(taskID int64, taskGroupID int64, stdout io.ReadCloser, parseDetail bool, detailFormat string, fifoPath string, pcapPath string) {
	defer stdout.Close()

	// 使用 bufio.Scanner 逐行读取
	scanner := bufio.NewScanner(stdout)
	// 增加缓冲区大小以支持长行
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	packetCounter := int64(0)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		logger.Debug(line)
		// 解析单个数据包
		packet, err := parseTsharkFieldsLine(line, taskID, packetCounter+1, taskGroupID)
		if err != nil {
			logger.Error("Failed to parse tshark fields line",
				zap.String("line", line),
				zap.Error(err))
			continue
		}

		// 保存到数据库
		if err := gorm.Repo.CreatePacket(packet); err != nil {
			logger.Error("Failed to save packet", zap.Error(err))
			continue
		}

		// 检查缓存中是否有该数据包的详情，如果有则立即更新
		if cachedContent, exists := getAndRemoveCachedDetail(taskID, packet.FrameNumber); exists {
			go func(tid, fnum int64, content string) {
				if err := updatePacketContentWithCompressed(tid, fnum, content); err != nil {
					logger.Debug("Update cached packet detail failed",
						zap.Int64("taskID", tid),
						zap.Int64("frameNumber", fnum),
						zap.Error(err))
				}
			}(taskID, packet.FrameNumber, cachedContent)
		}

		packetCounter++
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Overview scanner error", zap.Error(err))
	}

	logger.Info("Overview parsing completed",
		zap.Int64("taskID", taskID),
		zap.Int64("packetCount", packetCounter))

	// 如果需要解析详情，启动详情解析进程
	if parseDetail {
		go parsePacketDetails(taskID, fifoPath, detailFormat)
	}
}

// startTsharkParser 启动 tshark 进程解析 pcap 文件并保存 Packet 记录
func startTsharkParser(taskID int64, taskGroupID int64, pcapPath string, parseDetail bool) error {
	// 从配置中读取 tshark 路径
	tsharkPath := config.GetCaptureTsharkPath()
	if tsharkPath == "" {
		tsharkPath = "tshark" // 默认使用 PATH 中的 tshark
	}

	// 构建 tshark 命令，使用 fields 格式输出（更高效）
	// 输出格式: frame.number|frame.time_epoch|ip.src|ip.dst|Protocol|Info
	cmd := exec.Command(tsharkPath, "-r", pcapPath,
		"-T", "fields",
		"-E", "separator=|",
		"-e", "frame.number",
		"-e", "frame.time_epoch",
		"-e", "ip.src",
		"-e", "ip.dst",
		"-e", "_ws.col.Protocol",
		"-e", "_ws.col.Info",
	)

	// 获取标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tshark: %v", err)
	}

	// 保存 tshark 进程信息
	process := gorm.Process{
		TaskID: taskID,
		Pid:    int64(cmd.Process.Pid),
		Ppid:   getProcessPpidByProc(int(cmd.Process.Pid)), // 使用方法1：从 /proc 读取
		// Ppid: getProcessPpidByPs(int(cmd.Process.Pid)),  // 或者方法2：使用 ps 命令
		Type:    "tshark",
		Command: cmd.String(),
		Alive:   true,
	}
	gorm.Repo.CreateProcess(&process)

	// 异步解析 tshark 输出
	go func() {
		defer stdout.Close()

		// 使用 bufio.Scanner 逐行读取
		scanner := bufio.NewScanner(stdout)
		// 增加缓冲区大小以支持长行
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		packetCounter := int64(0)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// 解析单个数据包
			packet, err := parseTsharkFieldsLine(line, taskID, packetCounter+1, taskGroupID)
			if err != nil {
				logger.Error("Failed to parse tshark fields line",
					zap.String("line", line),
					zap.Error(err))
				continue
			}

			// 保存到数据库
			if err := gorm.Repo.CreatePacket(packet); err != nil {
				logger.Error("Failed to save packet", zap.Error(err))
				continue
			}

			// 检查缓存中是否有该数据包的详情，如果有则立即更新
			if cachedContent, exists := getAndRemoveCachedDetail(taskID, packet.FrameNumber); exists {
				go func(tid, fnum int64, content string) {
					if err := updatePacketContentWithCompressed(tid, fnum, content); err != nil {
						logger.Debug("Update cached packet detail failed",
							zap.Int64("taskID", tid),
							zap.Int64("frameNumber", fnum),
							zap.Error(err))
					}
				}(taskID, packet.FrameNumber, cachedContent)
			}

			packetCounter++
		}

		if err := scanner.Err(); err != nil {
			logger.Error("Scanner error", zap.Error(err))
		}

		// 更新进程状态
		process.Alive = false
		gorm.Repo.UpdateProcess(&process)

		logger.Info("Tshark parsing completed",
			zap.Int64("taskID", taskID),
			zap.Int64("packetCount", packetCounter))
	}()

	return nil
}

// parseTsharkFieldsLine 解析 tshark fields 格式的一行输出
// 格式: frame.number|frame.time_epoch|ip.src|ip.dst|Protocol|Info
func parseTsharkFieldsLine(line string, taskID int64, packetCount int64, taskGroupID int64) (*gorm.Packet, error) {
	// 使用 '|' 分割字段
	fields := strings.Split(line, "|")
	if len(fields) < 6 {
		return nil, fmt.Errorf("expected at least 6 fields, got %d", len(fields))
	}

	packet := &gorm.Packet{
		TaskID: taskID,
	}

	// 分配任务组内唯一编号
	if taskGroupID > 0 {
		counter := getOrCreateCounter(taskGroupID)
		packet.No = counter.Add(1)
	} else {
		packet.No = packetCount
	}

	// 1. Frame Number (已经是数字)
	if frameNum, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
		packet.FrameNumber = frameNum
	}

	// 2. Timestamp (epoch 格式，需要转换为纳秒)
	if timestampStr := fields[1]; timestampStr != "" {
		if seconds, err := strconv.ParseFloat(timestampStr, 64); err == nil {
			packet.Timestamp = int64(seconds * 1e9)
		}
	}

	// 3. Source IP/Address
	packet.Src = fields[2]
	if packet.Src == "" {
		packet.Src = "unknown"
	}

	// 4. Destination IP/Address
	packet.Dst = fields[3]
	if packet.Dst == "" {
		packet.Dst = "unknown"
	}

	// 5. Protocol
	packet.Protocol = fields[4]
	if packet.Protocol == "" {
		packet.Protocol = "UNKNOWN"
	}

	// 6. Length
	if length, err := strconv.ParseInt(fields[5], 10, 64); err == nil {
		packet.Length = length
	}

	// 7. Info
	packet.Info = fields[6]

	return packet, nil
}

// parsePacketDetails 从 FIFO 读取流量并解析详情，更新到对应 Packet 的 Content 字段
func parsePacketDetails(taskID int64, fifoPath string, detailFormat string) {
	logger.Info("Starting packet details parser",
		zap.Int64("taskID", taskID),
		zap.String("fifoPath", fifoPath),
		zap.String("format", detailFormat))

	// 根据 format 选择 tshark 输出格式
	var cmd *exec.Cmd
	// 从配置中读取 tshark 路径
	tsharkPath := config.GetCaptureTsharkPath()
	if tsharkPath == "" {
		tsharkPath = "tshark" // 默认使用 PATH 中的 tshark
	}

	switch strings.ToLower(detailFormat) {
	case "json":
		// JSON 格式（适合程序处理）
		cmd = exec.Command(tsharkPath, "-r", fifoPath, "-T", "json")
	case "pdml":
		// PDML 格式（XML 结构，最详细）
		cmd = exec.Command(tsharkPath, "-r", fifoPath, "-T", "pdml")
	case "ek":
		// ElasticSearch JSON 格式
		cmd = exec.Command(tsharkPath, "-r", fifoPath, "-T", "ek")
	default:
		// 默认使用 text 格式（人类可读）
		cmd = exec.Command(tsharkPath, "-r", fifoPath, "-V")
	}

	// 获取标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to create stdout pipe for details parser",
			zap.Int64("taskID", taskID),
			zap.Error(err))
		return
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start details parser",
			zap.Int64("taskID", taskID),
			zap.Error(err))
		return
	}

	// 保存 tshark 进程信息
	process := gorm.Process{
		TaskID: taskID,
		Pid:    int64(cmd.Process.Pid),
		Ppid:   getProcessPpidByProc(int(cmd.Process.Pid)), // 使用方法1：从 /proc 读取
		// Ppid: getProcessPpidByPs(int(cmd.Process.Pid)),  // 或者方法2：使用 ps 命令
		Type:    "tshark-detail",
		Command: cmd.String(),
		Alive:   true,
	}
	gorm.Repo.CreateProcess(&process)

	// 异步读取和保存详情
	go func() {
		defer stdout.Close()

		scanner := bufio.NewScanner(stdout)
		// 增加缓冲区以支持大输出
		scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

		packetIndex := int64(0)
		currentDetail := strings.Builder{}

		for scanner.Scan() {
			line := scanner.Text()

			// 检测新数据包的开始（根据不同格式）
			isNewPacket := false
			switch strings.ToLower(detailFormat) {
			case "json":
				if line == "{" || strings.HasPrefix(strings.TrimSpace(line), "{\"_index\"") {
					isNewPacket = true
				}
			case "pdml":
				if strings.Contains(line, "<packet>") {
					isNewPacket = true
				}
			default:
				// text 格式通过空行或 "Frame" 开头判断
				if strings.HasPrefix(line, "Frame ") || (line == "" && currentDetail.Len() > 0) {
					isNewPacket = true
				}
			}

			if isNewPacket && currentDetail.Len() > 0 {
				// 保存上一个数据包的详情
				packetIndex++
				detailContent := currentDetail.String()
				if err := updatePacketContent(taskID, packetIndex, detailContent); err != nil {
					logger.Error("Failed to update packet content",
						zap.Int64("taskID", taskID),
						zap.Int64("packetIndex", packetIndex),
						zap.Error(err))
				}
				currentDetail.Reset()
			}

			currentDetail.WriteString(line)
			currentDetail.WriteString("\n")
		}

		// 保存最后一个数据包
		if currentDetail.Len() > 0 {
			packetIndex++
			detailContent := currentDetail.String()
			if err := updatePacketContent(taskID, packetIndex, detailContent); err != nil {
				logger.Error("Failed to update packet content",
					zap.Int64("taskID", taskID),
					zap.Int64("packetIndex", packetIndex),
					zap.Error(err))
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Error("Details scanner error",
				zap.Int64("taskID", taskID),
				zap.Error(err))
		}

		// 更新进程状态
		process.Alive = false
		gorm.Repo.UpdateProcess(&process)

		logger.Info("Packet details parsing completed",
			zap.Int64("taskID", taskID),
			zap.Int64("packetCount", packetIndex))
	}()
}

// updatePacketContent 更新单个数据包的 Content 字段
func updatePacketContent(taskID int64, frameNumber int64, content string) error {
	// 查询对应的 Packet 记录
	packets, _, err := gorm.Repo.ListPacketsByTaskIDAndFrameNumber(taskID, frameNumber, frameNumber, 1, 1)
	if err != nil {
		return fmt.Errorf("failed to query packet: %v", err)
	}

	// 压缩数据包内容
	compressedContent, err := utils.CompressString(content)
	if err != nil {
		logger.Error("Failed to compress packet content for caching",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber),
			zap.Error(err))
		return fmt.Errorf("failed to compress packet content: %v", err)
	}
	if len(packets) == 0 {
		// 数据包尚未写入，放入缓存等待重试
		cachePacketDetail(taskID, frameNumber, compressedContent)
		logger.Debug("Packet not found, cached compressed detail for later update",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber))
		return nil // 返回 nil 表示已缓存，不需要报错
	}

	packet := packets[0]
	packet.Content = compressedContent

	// 更新数据库
	if err := gorm.Repo.UpdatePacket(packet); err != nil {
		return fmt.Errorf("failed to update packet content: %v", err)
	}

	return nil
}

// updatePacketContentWithCompressed 更新单个数据包的 Content 字段（直接使用已压缩的内容）
func updatePacketContentWithCompressed(taskID int64, frameNumber int64, compressedContent string) error {
	// 查询对应的 Packet 记录
	packets, _, err := gorm.Repo.ListPacketsByTaskIDAndFrameNumber(taskID, frameNumber, frameNumber, 1, 1)
	if err != nil {
		return fmt.Errorf("failed to query packet: %v", err)
	}

	if len(packets) == 0 {
		// 数据包尚未写入，重新缓存等待重试
		cachePacketDetail(taskID, frameNumber, compressedContent)
		logger.Debug("Packet not found, re-cached compressed detail for later update",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber))
		return nil
	}

	packet := packets[0]
	packet.Content = compressedContent

	// 更新数据库
	if err := gorm.Repo.UpdatePacket(packet); err != nil {
		return fmt.Errorf("failed to update packet content: %v", err)
	}

	return nil
}

// getProcessPpidByProc 方法1：从 /proc 文件系统读取父进程ID（推荐）
// 优点：快速、可靠、不依赖外部命令
// 缺点：仅适用于 Linux 系统
func getProcessPpidByProc(pid int) int64 {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	content, err := os.ReadFile(statusPath)
	if err != nil {
		logger.Debug("Failed to read /proc status",
			zap.Int("pid", pid),
			zap.Error(err))
		return 0
	}

	// 查找 PPid 行
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "PPid:") {
			ppidStr := strings.TrimSpace(strings.TrimPrefix(line, "PPid:"))
			if ppid, err := strconv.Atoi(ppidStr); err == nil {
				return int64(ppid)
			}
			break
		}
	}

	logger.Debug("PPid not found in /proc status", zap.Int("pid", pid))
	return 0
}

// getProcessPpidByPs 方法2：使用 ps 命令获取父进程ID
// 优点：跨平台（Linux、macOS）
// 缺点：需要执行外部命令，性能稍慢
func getProcessPpidByPs(pid int) int64 {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		logger.Debug("Failed to execute ps command",
			zap.Int("pid", pid),
			zap.Error(err))
		return 0
	}

	ppidStr := strings.TrimSpace(string(output))
	if ppidStr == "" {
		logger.Debug("Empty output from ps command", zap.Int("pid", pid))
		return 0
	}

	if ppid, err := strconv.Atoi(ppidStr); err == nil {
		return int64(ppid)
	}

	logger.Debug("Failed to parse ppid from ps output",
		zap.String("output", ppidStr))
	return 0
}

func StopCapture(taskGroupId, taskId string) error {
	// TODO: 实现停止任务的功能
	return nil
}
