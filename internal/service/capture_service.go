package service

import (
	"bufio"
	"errors"
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
	// key 为任务组 ID
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

// cachedDetail 缓存的详情项（包含重试次数）
type cachedDetail struct {
	Content    string
	RetryCount int
	MaxRetries int
}

var packetDetailCache = struct {
	cache map[packetDetailKey]*cachedDetail
	sync.RWMutex
}{
	cache: make(map[packetDetailKey]*cachedDetail),
}

// cachePacketDetail 缓存数据包详情
func cachePacketDetail(taskID, frameNumber int64, content string) {
	packetDetailCache.Lock()
	defer packetDetailCache.Unlock()
	key := packetDetailKey{TaskID: taskID, FrameNumber: frameNumber}

	// 如果已存在，增加重试次数；否则创建新条目
	if existing, exists := packetDetailCache.cache[key]; exists {
		existing.RetryCount++
		existing.Content = content // 更新为最新的压缩内容
	} else {
		packetDetailCache.cache[key] = &cachedDetail{
			Content:    content,
			RetryCount: 0,
			MaxRetries: 10, // 最多重试10次（50秒）
		}
	}
}

// getAndRemoveCachedDetail 获取并删除缓存的数据包详情
func getAndRemoveCachedDetail(taskID, frameNumber int64) (string, bool, int) {
	packetDetailCache.Lock()
	defer packetDetailCache.Unlock()
	key := packetDetailKey{TaskID: taskID, FrameNumber: frameNumber}
	cached, exists := packetDetailCache.cache[key]
	if exists {
		delete(packetDetailCache.cache, key)
		return cached.Content, true, cached.RetryCount
	}
	return "", false, 0
}

// flushAllCachedDetailsForTask 清空并更新指定任务的所有缓存详情（用于停止任务时）
func flushAllCachedDetailsForTask(taskID int64) {
	logger.Info("Flushing cached details for task",
		zap.Int64("taskID", taskID))

	// 获取该任务的所有缓存键
	packetDetailCache.RLock()
	keys := make([]packetDetailKey, 0)
	for k := range packetDetailCache.cache {
		if k.TaskID == taskID {
			keys = append(keys, k)
		}
	}
	packetDetailCache.RUnlock()

	if len(keys) == 0 {
		logger.Debug("No cached details to flush",
			zap.Int64("taskID", taskID))
		return
	}

	logger.Info("Found cached details to flush",
		zap.Int64("taskID", taskID),
		zap.Int("count", len(keys)))

	// 尝试更新所有缓存的详情
	successCount := 0
	failCount := 0
	for _, key := range keys {
		compressedContent, exists, _ := getAndRemoveCachedDetail(key.TaskID, key.FrameNumber)
		if !exists {
			continue
		}

		// 尝试更新（缓存中已是压缩后的数据）
		if err := updatePacketContentWithCompressed(key.TaskID, key.FrameNumber, compressedContent); err != nil {
			/*logger.Debug("Flush update packet content failed",
			zap.Int64("taskID", key.TaskID),
			zap.Int64("frameNumber", key.FrameNumber),
			zap.Int("retryCount", retryCount),
			zap.Error(err))*/
			failCount++
		} else {
			/*logger.Debug("Flush update packet content succeeded",
			zap.Int64("taskID", key.TaskID),
			zap.Int64("frameNumber", key.FrameNumber),
			zap.Int("retryCount", retryCount))*/
			successCount++
		}
	}

	logger.Info("Cached details flush completed",
		zap.Int64("taskID", taskID),
		zap.Int("successCount", successCount),
		zap.Int("failCount", failCount))
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
			compressedContent, exists, retryCount := getAndRemoveCachedDetail(key.TaskID, key.FrameNumber)
			if !exists {
				continue
			}

			// 尝试更新（缓存中已是压缩后的数据）
			if err := updatePacketContentWithCompressed(key.TaskID, key.FrameNumber, compressedContent); err != nil {
				logger.Debug("Retry update packet content failed, re-caching",
					zap.Int64("taskID", key.TaskID),
					zap.Int64("frameNumber", key.FrameNumber),
					zap.Int("retryCount", retryCount),
					zap.Error(err))
				// 重新放回缓存
				cachePacketDetail(key.TaskID, key.FrameNumber, compressedContent)
			} else {
				//logger.Debug("Retry update packet content succeeded",
				//	zap.Int64("taskID", key.TaskID),
				//	zap.Int64("frameNumber", key.FrameNumber),
				//	zap.Int("retryCount", retryCount))
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
		Interfaces:      gorm.StringArray(capture.Interfaces),
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

	// 2.1 如果需要解析详情，先创建 FIFO 文件
	if parseDetail {
		if err := CreateFIFO(fifoPath); err != nil {
			logger.Error("Failed to create FIFO",
				zap.String("fifoPath", fifoPath),
				zap.Error(err))
			// 更新任务状态为失败
			task.Status = "failed"
			task.Message += fmt.Sprintf("Failed to create FIFO: %v; ", err)
			err := gorm.Repo.UpdateTask(&task)
			if err != nil {
				logger.Error("Failed to update task status", zap.Error(err))
			}
			return task.ID, fmt.Errorf("failed to create FIFO: %v", err)
		}
		logger.Info("FIFO created",
			zap.Int64("taskID", task.ID),
			zap.String("fifoPath", fifoPath))
	}

	// 3. 更新任务的文件路径
	task.FilePath = pcapPath
	task.FifoPath = fifoPath
	err = gorm.Repo.UpdateTask(&task)
	if err != nil {
		logger.Error("Failed to update task file paths", zap.Error(err))
		task.Message += fmt.Sprintf("Failed to update file paths: %v; ", err)
	}

	// 4. 构建并执行抓包命令
	cmd, fullCommand := buildCaptureCommand(host, capture, pcapPath, fifoPath, onlyCapture, parseDetail)

	// 4.1 保存完整命令到数据库
	task.FullCommand = fullCommand
	err = gorm.Repo.UpdateTask(&task)
	if err != nil {
		logger.Error("Failed to update task full command", zap.Error(err))
		task.Message += fmt.Sprintf("Failed to update full command: %v; ", err)
	}

	// 5. 启动进程前先设置stderr和stdout捕获
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to create stderr pipe", zap.Error(err))
		task.Message += fmt.Sprintf("Failed to create stderr pipe: %v; ", err)
	} else {
		// 启动一个协程读取stderr
		go func() {
			buf := make([]byte, 4096)
			for {
				n, readErr := stderrPipe.Read(buf)
				if n > 0 {
					line := string(buf[:n])
					// 实时打印 stderr 内容
					logger.Warn("Capture process stderr output",
						zap.Int64("taskID", task.ID),
						zap.String("output", line))
				}
				if readErr != nil {
					if readErr != io.EOF {
						logger.Warn("Failed to read stderr of sshpass", zap.Error(readErr))
					}
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
			logger.Error("Failed to create stdout pipe for overview", zap.Error(err))
			task.Status = "failed"
			task.Message += fmt.Sprintf("Failed to create stdout pipe for overview: %v; ", err)
			err := gorm.Repo.UpdateTask(&task)
			if err != nil {
				logger.Error("Failed to update task status", zap.Error(err))
			}
			return task.ID, fmt.Errorf("failed to create stdout pipe for overview: %v", err)
		}
	}

	if err := cmd.Start(); err != nil {
		task.Status = "failed"
		task.Message += fmt.Sprintf("Failed to start process: %v; ", err)
		err := gorm.Repo.UpdateTask(&task)
		if err != nil {
			logger.Error("Failed to update task status", zap.Error(err))
		}
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

	// 7. 启动协程监控进程状态
	go monitorProcessStatus(task.ID, cmd, process.ID)

	// 8. 如果需要解析概览，启动协程实时解析输出（传递已获取的stdout）
	if !onlyCapture {
		go parseOverviewFromOutput(task.ID, task.TaskGroupId, stdout)

		// 如果需要解析详情，同时启动详情解析进程（从 FIFO 读取）
		if parseDetail {
			logger.Info("Starting packet details parser from FIFO",
				zap.Int64("taskGroupID", task.TaskGroupId),
				zap.String("taskName", taskName),
				zap.Int64("taskID", task.ID),
				zap.String("fifoPath", fifoPath),
				zap.String("detailFormat", detailFormat),
				zap.String("wiresharkFilter", capture.WiresharkFilter))
			go parsePacketDetails(task.ID, fifoPath, detailFormat, capture.WiresharkFilter)
		}
	}
	logger.Info("sshpass process started",
		zap.Int64("taskId", task.ID),
		zap.Int64("processId", process.ID),
		zap.String("host", host.IP),
		zap.String("pcapPath", pcapPath),
		zap.String("fifoPath", fifoPath))

	return task.ID, nil
}

// buildCaptureCommand 构建抓包命令
// 命令结构:
//
//	sshpass -p 'password' ssh user@host 'tcpdump -i eth0 -U -w - bpf_filter' | \
//	  tee output.pcap fifo | \
//	  (如果需要解析: tshark -T json -r - > fifo)
func buildCaptureCommand(host *gorm.Host, capture HostSingleCapture, pcapPath, fifoPath string, onlyCapture bool, parseDetail bool) (*exec.Cmd, string) {
	// 从配置中读取 tshark 路径
	tsharkPath := config.GetCaptureTsharkPath()
	if tsharkPath == "" {
		tsharkPath = "tshark" // 默认使用 PATH 中的 tshark
	}

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

	// 3. 构建 Wireshark 显示过滤器参数
	wiresharkFilterArg := ""
	if capture.WiresharkFilter != "" {
		wiresharkFilterArg = fmt.Sprintf("-Y '%s'", capture.WiresharkFilter)
	}

	// 4. 根据是否需要解析，构建完整的管道命令
	var fullCommand string

	if onlyCapture {
		// 只抓包，不解析：直接保存到 pcap 文件
		fullCommand = fmt.Sprintf("%s > %s", sshCmd, pcapPath)
	} else if !parseDetail {
		// 解析概览：使用 tee 保存 pcap，同时用 tshark 实时解析
		// ssh ... | tee pcap | tshark -l -i - -Y 'filter' -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info
		fullCommand = fmt.Sprintf("%s | tee %s | %s -l -i - %s -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info",
			sshCmd, pcapPath, tsharkPath, wiresharkFilterArg)
	} else {
		// 解析详情：使用 tee 保存 pcap，同时复制流量到 fifo
		// ssh ... | tee pcap fifo | tshark -l -i - -Y 'filter' -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info
		// 然后由另一个 tshark 进程读取 fifo
		fullCommand = fmt.Sprintf("%s | tee %s %s | %s -l -i - %s -T fields -E separator='|' -e frame.number -e frame.time_epoch -e ip.src -e ip.dst -e _ws.col.Protocol -e frame.len -e _ws.col.Info",
			sshCmd, pcapPath, fifoPath, tsharkPath, wiresharkFilterArg)
	}

	cmd := exec.Command("bash", "-c", fullCommand)

	// 设置进程组，以便后续可以一起终止所有子进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 返回命令和完整命令字符串
	return cmd, fullCommand
}

// monitorProcessStatus 监控进程状态，等待进程结束并更新数据库
func monitorProcessStatus(taskID int64, cmd *exec.Cmd, processID int64) {
	// 等待命令执行完成
	err := cmd.Wait()

	// 更新任务状态
	task, _ := gorm.Repo.GetTaskByID(taskID)
	if err != nil {
		task.Message += fmt.Sprintf("Process exited error: %v; ", err)
	} else {
		task.Message += "Normal exit; "
	}
	err = gorm.Repo.UpdateTask(task)
	if err != nil {
		logger.Error("Failed to update task record", zap.Error(err))
	}

	// 更新进程状态
	process, _ := gorm.Repo.GetProcessByID(processID)
	process.Alive = false
	err = gorm.Repo.UpdateProcess(process)
	if err != nil {
		logger.Error("Failed to update process record", zap.Error(err))
	}

	logger.Info("sshpass process stopped",
		zap.Int64("taskID", taskID),
		zap.Int64("processID", processID),
		zap.String("taskStatus", task.Status),
		zap.String("taskMessage", task.Message))
}

// parseOverviewFromOutput 从进程输出实时解析数据包概览
func parseOverviewFromOutput(taskID int64, taskGroupID int64, stdout io.ReadCloser) {
	defer func(stdout io.ReadCloser) {
		err := stdout.Close()
		if err != nil {
			logger.Error("Failed to close stdout", zap.Error(err))
		}
	}(stdout)

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

		//logger.Debug(line)
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
		if cachedContent, exists, _ := getAndRemoveCachedDetail(taskID, packet.FrameNumber); exists {
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
		zap.Int64("taskGroupID", taskGroupID),
		zap.Int64("taskID", taskID),
		zap.Int64("packetCount", packetCounter))
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
func parsePacketDetails(taskID int64, fifoPath string, detailFormat string, wiresharkFilter string) {
	// 根据 format 选择 tshark 输出格式
	var cmd *exec.Cmd
	// 从配置中读取 tshark 路径
	tsharkPath := config.GetCaptureTsharkPath()
	if tsharkPath == "" {
		tsharkPath = "tshark" // 默认使用 PATH 中的 tshark
	}

	// 构建 tshark 命令参数
	args := []string{"-r", fifoPath}

	// 添加 Wireshark 过滤器（如果提供）
	if wiresharkFilter != "" {
		args = append(args, "-Y", wiresharkFilter)
	}

	// 根据格式添加输出类型参数
	switch strings.ToLower(detailFormat) {
	case "json":
		// JSON 格式（适合程序处理）
		args = append(args, "-T", "json")
	case "pdml":
		// PDML 格式（XML 结构，最详细）
		args = append(args, "-T", "pdml")
	case "ek":
		// ElasticSearch JSON 格式
		args = append(args, "-T", "ek")
	default:
		// 默认使用 text 格式（人类可读）
		args = append(args, "-V")
	}

	cmd = exec.Command(tsharkPath, args...)

	// 设置进程组，以便后续可以一起终止所有子进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 获取标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to create stdout pipe for details parser",
			zap.Int64("taskID", taskID),
			zap.Error(err))
		return
	}

	// 获取标准错误（必须在 Start 之前）
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to create stderr pipe for details parser",
			zap.Int64("taskID", taskID),
			zap.Error(err))
	} else {
		// 启动一个协程读取stderr
		go func() {
			buf := make([]byte, 4096)
			for {
				n, readErr := stderrPipe.Read(buf)
				if n > 0 {
					line := string(buf[:n])
					// 实时打印 stderr 内容
					logger.Warn("Details parser stderr output",
						zap.Int64("taskID", taskID),
						zap.String("output", line))
				}
				if readErr != nil {
					if readErr != io.EOF {
						logger.Warn("Failed to read stderr of tshark-detail", zap.Error(readErr))
					}
					break
				}
			}
		}()
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start details parser",
			zap.Int64("taskID", taskID),
			zap.Error(err))
		return
	}

	logger.Info("Details parser started, waiting for data from FIFO",
		zap.Int64("taskID", taskID),
		zap.Int("pid", cmd.Process.Pid))

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
	err = gorm.Repo.CreateProcess(&process)
	if err != nil {
		logger.Error("Failed to save tshark-detail process",
			zap.Int64("taskID", taskID),
			zap.Error(err))
	}

	// 异步读取和保存详情
	go func() {
		// 确保在 goroutine 结束时回收进程状态，避免僵尸进程
		defer func() {
			if err := cmd.Wait(); err != nil {
				logger.Error("Details parser process exited error",
					zap.Int64("taskID", taskID),
					zap.Int("pid", cmd.Process.Pid),
					zap.Error(err))
			} else {
				logger.Info("Details parser process exited normally",
					zap.Int64("taskID", taskID),
					zap.Int("pid", cmd.Process.Pid))
			}
		}()
		defer func(stdout io.ReadCloser) {
			err := stdout.Close()
			if err != nil {
				logger.Error("Failed to close stdout pipe",
					zap.Int64("taskID", taskID),
					zap.Error(err))
			}
		}(stdout)

		scanner := bufio.NewScanner(stdout)
		// 增加缓冲区以支持大输出
		scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

		packetIndex := int64(0)
		currentDetail := strings.Builder{}
		var currentFrameNumber int64 // 当前数据包的帧号

		for scanner.Scan() {
			line := scanner.Text()

			// 检测新数据包的开始（根据不同格式）
			isNewPacket := false
			nextFrameNumber := int64(0) // 新数据包的帧号（将在下一个循环中使用）

			switch strings.ToLower(detailFormat) {
			case "json":
				if line == "{" || strings.HasPrefix(strings.TrimSpace(line), "{\"_index\"") {
					isNewPacket = true
					// JSON 格式的帧号需要从完整 JSON 中解析，这里先标记，稍后提取
				}
			case "pdml":
				if strings.Contains(line, "<packet>") {
					isNewPacket = true
					// PDML 格式：从 <field name="num" value="123"/> 提取帧号
					if strings.Contains(line, `name="num"`) {
						if frameNum := extractFrameNumberFromPDML(line); frameNum > 0 {
							nextFrameNumber = frameNum
						}
					}
				}
			default:
				// text 格式通过 "Frame" 开头判断
				if strings.HasPrefix(line, "Frame ") {
					isNewPacket = true
					// Text 格式：从 "Frame 123:" 提取帧号
					if frameNum := extractFrameNumberFromText(line); frameNum > 0 {
						nextFrameNumber = frameNum
					}
				}
			}

			if isNewPacket && currentDetail.Len() > 0 {
				// 保存上一个数据包的详情
				packetIndex++
				savePacketDetail(taskID, detailFormat, packetIndex, currentDetail.String(), currentFrameNumber)
				currentDetail.Reset()
			}

			// 如果是新数据包，将提取的帧号保存为下一个数据包的帧号
			if isNewPacket {
				if nextFrameNumber > 0 {
					currentFrameNumber = nextFrameNumber
				} else {
					// 如果没有从当前行提取到帧号，尝试从后续行提取
					currentFrameNumber = extractFrameNumberFromLine(line, detailFormat)
				}
			}

			currentDetail.WriteString(line)
			currentDetail.WriteString("\n")
		}

		// 保存最后一个数据包
		if currentDetail.Len() > 0 {
			packetIndex++
			savePacketDetail(taskID, detailFormat, packetIndex, currentDetail.String(), currentFrameNumber)
		}

		if err := scanner.Err(); err != nil {
			logger.Error("Details scanner error",
				zap.Int64("taskID", taskID),
				zap.Error(err))
		}

		// 更新进程状态
		process.Alive = false
		err := gorm.Repo.UpdateProcess(&process)
		if err != nil {
			logger.Error("Failed to update tshark-detail process alive",
				zap.Int64("taskID", taskID),
				zap.Error(err))
		}

		logger.Info("Packet details parsing completed",
			zap.Int64("taskID", taskID),
			zap.Int64("packetCount", packetIndex))
	}()
}

// updatePacketContent 更新单个数据包的 Content 字段
func updatePacketContent(taskID int64, frameNumber int64, content string) error {
	// 压缩数据包内容
	compressedContent, err := utils.CompressString(content)
	if err != nil {
		logger.Error("Failed to compress packet content for caching",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber),
			zap.Error(err))
		return fmt.Errorf("failed to compress packet content: %v", err)
	}

	return savePacketContent(taskID, frameNumber, compressedContent)
}

// updatePacketContentWithCompressed 更新单个数据包的 Content 字段（直接使用已压缩的内容）
func updatePacketContentWithCompressed(taskID int64, frameNumber int64, compressedContent string) error {
	return savePacketContent(taskID, frameNumber, compressedContent)
}

// savePacketContent 保存数据包详情到数据库（内部辅助函数）
func savePacketContent(taskID int64, frameNumber int64, compressedContent string) error {
	// 查询对应的 Packet 记录
	packets, _, err := gorm.Repo.ListPacketsByTaskIDAndFrameNumber(taskID, frameNumber, frameNumber, 1, 1)
	if err != nil {
		return fmt.Errorf("failed to query packet: %v", err)
	}

	if len(packets) == 0 {
		// 数据包尚未写入，放入缓存等待重试
		cachePacketDetail(taskID, frameNumber, compressedContent)
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

// savePacketDetail 保存单个数据包的详情（内部辅助函数）
func savePacketDetail(taskID int64, detailFormat string, packetIndex int64, detailContent string, currentFrameNumber int64) {
	// 确定使用的帧号：优先使用当前数据包提取的帧号，否则从内容中提取或使用序号
	frameNumber := currentFrameNumber
	if frameNumber <= 0 {
		// 对于 JSON 格式，尝试从完整内容中提取帧号
		if strings.ToLower(detailFormat) == "json" {
			frameNumber = extractFrameNumberFromJSON(detailContent)
		}
		// 如果还是没提取到，使用序号
		if frameNumber <= 0 {
			frameNumber = packetIndex
		}
	}

	if err := updatePacketContent(taskID, frameNumber, detailContent); err != nil {
		logger.Error("Failed to update packet content",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber),
			zap.Error(err))
	}
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

// StopCapture 停止抓包任务
// taskGroupId: 任务组ID，如果为空字符串则忽略
// taskId: 单个任务ID，如果为空字符串则忽略
// 优先级：如果两者都提供，优先停止单个任务
func StopCapture(taskGroupId, taskId string) error {
	// 参数验证
	if taskGroupId == "" && taskId == "" {
		return fmt.Errorf("taskGroupId or taskId must be provided")
	}

	// 优先处理单个任务停止
	if taskId != "" {
		taskID, err := strconv.ParseInt(taskId, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid taskId: %v", err)
		}
		return stopSingleTask(taskID)
	}

	// 处理任务组停止
	taskGroupID, err := strconv.ParseInt(taskGroupId, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid taskGroupId: %v", err)
	}
	return stopTaskGroup(taskGroupID)
}

// stopSingleTask 停止单个任务
func stopSingleTask(taskID int64) error {
	logger.Info("Stopping single task", zap.Int64("taskID", taskID))

	// 1. 获取任务信息
	task, err := gorm.Repo.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %v", err)
	}

	if task == nil {
		return fmt.Errorf("task %d not found", taskID)
	}

	// 2. 检查任务状态
	if task.Status == "stopped" || task.Status == "failed" {
		logger.Info("Task already stopped", zap.Int64("taskID", taskID), zap.String("status", task.Status))
		return nil
	}

	// 3. 获取任务相关的所有进程
	processes, err := gorm.Repo.ListProcessesByTaskID(taskID)
	if err != nil {
		logger.Warn("Failed to list processes", zap.Int64("taskID", taskID), zap.Error(err))
		// 继续执行，尝试直接更新任务状态
	} else {
		// 4. 终止所有存活的进程
		for _, proc := range processes {
			if proc.Alive {
				killProcess(proc)
			}
		}
	}

	// 5. 更新任务状态
	task.Status = "stopped"
	task.StopAt = new(time.Now())
	task.Message = "Manually stopped by user"
	if err := gorm.Repo.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	// 6. 清空并更新所有缓存的详情（执行最后一次更新）
	flushAllCachedDetailsForTask(taskID)

	logger.Info("Task stopped successfully", zap.Int64("taskID", taskID))
	return nil
}

// stopTaskGroup 停止任务组中的所有任务
func stopTaskGroup(taskGroupID int64) error {
	logger.Info("Stopping task group", zap.Int64("taskGroupID", taskGroupID))

	// 1. 获取任务组信息
	taskGroup, err := gorm.Repo.GetTaskGroupByID(taskGroupID)
	if err != nil {
		return fmt.Errorf("failed to get task group: %v", err)
	}

	if taskGroup == nil {
		return fmt.Errorf("task group %d not found", taskGroupID)
	}

	// 2. 获取任务组下的所有任务（分页获取所有）
	page := 1
	pageSize := 100
	totalStopped := 0

	for {
		tasks, total, err := gorm.Repo.ListTasksByTaskGroupID(taskGroupID, page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %v", err)
		}

		if len(tasks) == 0 {
			break
		}

		// 3. 停止每个任务
		for _, task := range tasks {
			if task.Status == "stopped" || task.Status == "failed" {
				continue
			}

			if err := stopSingleTask(task.ID); err != nil {
				logger.Error("Failed to stop task in group",
					zap.Int64("taskID", task.ID),
					zap.Error(err))
				// 继续处理其他任务
			} else {
				totalStopped++
			}
		}

		// 检查是否已获取所有任务
		if int64(page*pageSize) >= total {
			break
		}
		page++
	}

	// 4. 更新任务组状态
	taskGroup.StopAt = new(time.Now())
	if err := gorm.Repo.UpdateTaskGroup(taskGroup); err != nil {
		logger.Warn("Failed to update task group status", zap.Error(err))
	}

	logger.Info("Task group stopped successfully",
		zap.Int64("taskGroupID", taskGroupID),
		zap.Int("totalStopped", totalStopped))

	return nil
}

// killProcess 终止进程及其子进程
func killProcess(proc *gorm.Process) {
	logger.Info("Killing process",
		zap.Int64("pid", proc.Pid),
		zap.Int64("ppid", proc.Ppid),
		zap.String("type", proc.Type))

	// 1. 首先尝试终止整个进程组（因为设置了Setpgid）
	if proc.Pid > 0 {
		// 使用 -Pid 向整个进程组发送信号
		if err := syscall.Kill(-int(proc.Pid), syscall.SIGTERM); err != nil {
			logger.Warn("Failed to kill process group",
				zap.Int64("pid", proc.Pid),
				zap.Error(err))

			// 如果失败，尝试单独终止进程
			if err := syscall.Kill(int(proc.Pid), syscall.SIGTERM); err != nil {
				logger.Warn("Failed to kill single process",
					zap.Int64("pid", proc.Pid),
					zap.Error(err))
			}
		}
	}

	// 2. 等待一小段时间让进程优雅退出
	time.Sleep(500 * time.Millisecond)

	// 3. 检查进程是否仍然存在，如果存在则强制杀死
	if isProcessAlive(int(proc.Pid)) {
		logger.Info("Process still alive, sending SIGKILL", zap.Int64("pid", proc.Pid))
		if proc.Pid > 0 {
			syscall.Kill(-int(proc.Pid), syscall.SIGKILL)
			syscall.Kill(int(proc.Pid), syscall.SIGKILL)
		}
	}

	// 4. 更新进程状态
	proc.Alive = false
	if err := gorm.Repo.UpdateProcess(proc); err != nil {
		logger.Warn("Failed to update process status",
			zap.Int64("pid", proc.Pid),
			zap.Error(err))
	}
}

// isProcessAlive 检查进程是否仍然存活
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	// 发送信号0来检查进程是否存在
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}

	// 如果错误是ESRCH，说明进程不存在
	if errors.Is(err, syscall.ESRCH) {
		return false
	}

	// 如果是EPERM，说明进程存在但没有权限（仍然存活）
	if errors.Is(err, syscall.EPERM) {
		return true
	}

	return false
}

// GetPacketDetail 获取单个数据包的详情（解压缩后返回）
func GetPacketDetail(taskID int64, frameNumber int64) (string, error) {
	// 查询数据包记录
	packets, _, err := gorm.Repo.ListPacketsByTaskIDAndFrameNumber(taskID, frameNumber, frameNumber, 1, 1)
	if err != nil {
		return "", fmt.Errorf("failed to query packet: %v", err)
	}

	if len(packets) == 0 {
		return "", fmt.Errorf("packet not found: taskID=%d, frameNumber=%d", taskID, frameNumber)
	}

	packet := packets[0]

	// 如果 Content 为空，说明没有详情数据
	if packet.Content == "" {
		return "", fmt.Errorf("packet detail is empty: taskID=%d, frameNumber=%d", taskID, frameNumber)
	}

	// 解压缩数据包内容
	decompressedContent, err := utils.DecompressString(packet.Content)
	if err != nil {
		logger.Error("Failed to decompress packet content",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber),
			zap.Error(err))
		return "", fmt.Errorf("failed to decompress packet content: %v", err)
	}

	return decompressedContent, nil
}

// extractFrameNumberFromText 从 text 格式的行中提取帧号
// 示例: "Frame 123: 1234 bytes on wire..."
func extractFrameNumberFromText(line string) int64 {
	if !strings.HasPrefix(line, "Frame ") {
		return 0
	}

	// 移除 "Frame " 前缀
	rest := strings.TrimPrefix(line, "Frame ")

	// 查找冒号前的数字
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return 0
	}

	frameNumStr := strings.TrimSpace(rest[:colonIdx])
	if frameNum, err := strconv.ParseInt(frameNumStr, 10, 64); err == nil {
		return frameNum
	}

	return 0
}

// extractFrameNumberFromPDML 从 PDML 格式的行中提取帧号
// 示例: <field name="num" value="123" show="123"/>
func extractFrameNumberFromPDML(line string) int64 {
	// 查找 name="num" 属性
	if !strings.Contains(line, `name="num"`) {
		return 0
	}

	// 查找 value="123" 属性
	valueIdx := strings.Index(line, `value="`)
	if valueIdx == -1 {
		return 0
	}

	rest := line[valueIdx+7:] // 跳过 'value="'
	quoteIdx := strings.Index(rest, `"`)
	if quoteIdx == -1 {
		return 0
	}

	frameNumStr := rest[:quoteIdx]
	if frameNum, err := strconv.ParseInt(frameNumStr, 10, 64); err == nil {
		return frameNum
	}

	return 0
}

// extractFrameNumberFromJSON 从 JSON 格式中提取帧号
// 需要在完整的 JSON 对象中查找 "_source.layers.frame.number" 或类似字段
func extractFrameNumberFromJSON(jsonContent string) int64 {
	// 尝试查找常见的帧号字段
	patterns := []string{
		`"num":"`,          // tshark JSON 格式
		`"number":"`,       // 另一种格式
		`"frame.number":"`, // 完整字段名
	}

	for _, pattern := range patterns {
		idx := strings.Index(jsonContent, pattern)
		if idx == -1 {
			continue
		}

		rest := jsonContent[idx+len(pattern):]
		quoteIdx := strings.Index(rest, `"`)
		if quoteIdx == -1 {
			continue
		}

		frameNumStr := rest[:quoteIdx]
		if frameNum, err := strconv.ParseInt(frameNumStr, 10, 64); err == nil {
			return frameNum
		}
	}

	return 0
}

// extractFrameNumberFromLine 根据格式从行中提取帧号
func extractFrameNumberFromLine(line string, detailFormat string) int64 {
	switch strings.ToLower(detailFormat) {
	case "json":
		// JSON 格式需要从完整内容中提取，这里返回 0
		return 0
	case "pdml":
		return extractFrameNumberFromPDML(line)
	default:
		// text 格式
		return extractFrameNumberFromText(line)
	}
}
