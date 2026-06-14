package gorm

import (
	"fmt"
	"sync"
	"time"
	"webshark/internal/config"
	"webshark/internal/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

// Repo 全局Repo
var (
	Repo *WebSharkRepository
	once sync.Once
)

// WebSharkRepository 接入数据仓库
type WebSharkRepository struct {
	db     *gorm.DB
	config *config.DBConfig
	// 预创建的泛型 repository 实例，避免重复创建
	hostRepo      *BaseRepository[Host]
	taskRepo      *BaseRepository[Task]
	taskGroupRepo *BaseRepository[TaskGroup]
	packetRepo    *BaseRepository[Packet]
	processRepo   *BaseRepository[Process]
}

func InitWebSharkRepository(dbConfig *config.DBConfig) (*WebSharkRepository, error) {
	var initErr error
	once.Do(func() {
		Repo, initErr = initWebSharkRepository(dbConfig)
		if initErr != nil {
			initErr = fmt.Errorf("failed to init WebSharkRepository: %w", initErr)
		}
	})
	return Repo, initErr
}

// initWebSharkRepository 初始化WebShark数据仓库
func initWebSharkRepository(dbConfig *config.DBConfig) (*WebSharkRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	// 从全局配置读取 SQL 日志开关
	enableSQLLog := false
	if globalCfg, err := config.GetConfig(); err == nil && globalCfg != nil {
		enableSQLLog = globalCfg.Output.EnableSQLLog
	}

	// 配置 GORM 日志：集成到项目的 zap 日志框架
	gormConfig := gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  gormlogger.Warn, // 只输出警告和错误（包括慢 SQL）
	}

	// 如果开启 SQL 日志，设置为 Info 级别
	if enableSQLLog {
		gormConfig.LogLevel = gormlogger.Info
	}

	gormLog := logger.NewGormLogger(logger.GetLogger(), gormConfig)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		logger.Error("failed to connect database", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get sql.DB", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

	repo := &WebSharkRepository{
		db:     db,
		config: dbConfig,
		// 预创建所有 repository 实例
		hostRepo:      NewBaseRepository[Host](db),
		taskRepo:      NewBaseRepository[Task](db),
		taskGroupRepo: NewBaseRepository[TaskGroup](db),
		packetRepo:    NewBaseRepository[Packet](db),
		processRepo:   NewBaseRepository[Process](db),
	}

	// 自动迁移表结构
	if err := repo.autoMigrate(); err != nil {
		logger.Error("failed to migrate database", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	Repo = repo

	return repo, nil
}

// autoMigrate 自动迁移表结构
func (r *WebSharkRepository) autoMigrate() error {
	// 使用 GetAllModels() 自动获取所有模型，无需手动维护
	models := GetAllModels()
	return r.db.AutoMigrate(models...)
}

// Close 关闭数据库连接
func (r *WebSharkRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ==================== Host 增删改查 ====================

// CreateHost 创建主机记录
func (r *WebSharkRepository) CreateHost(host *Host) error {
	return r.hostRepo.Create(host)
}

// GetHostByID 根据 ID 获取主机
func (r *WebSharkRepository) GetHostByID(id int64) (*Host, error) {
	return r.hostRepo.GetByID(id)
}

// GetHostByIP 根据 IP 获取主机（特殊查询）
func (r *WebSharkRepository) GetHostByIP(ip string) (*Host, error) {
	return r.hostRepo.FirstByCondition("ip = ?", ip)
}

// ListHosts 获取主机列表（支持分页）
func (r *WebSharkRepository) ListHosts(page, pageSize int) ([]*Host, int64, error) {
	return r.hostRepo.List(page, pageSize, "created_at DESC")
}

// SearchHosts 搜索主机（根据 hostname、IP 或用户名）
func (r *WebSharkRepository) SearchHosts(keyword string, page, pageSize int) ([]*Host, int64, error) {
	query := "host_name LIKE ? OR ip LIKE ? OR user_name LIKE ?"
	args := []interface{}{"%" + keyword + "%", "%" + keyword + "%", "%" + keyword + "%"}
	return r.hostRepo.ListByCondition(query, args, page, pageSize, "created_at DESC")
}

// UpdateHost 更新主机信息
func (r *WebSharkRepository) UpdateHost(host *Host) error {
	return r.hostRepo.Update(host)
}

// UpdateHostPassword 更新主机密码
func (r *WebSharkRepository) UpdateHostPassword(id int64, newPassword string) error {
	return r.hostRepo.UpdateField(id, "password", newPassword)
}

// DeleteHost 删除主机（物理删除）
func (r *WebSharkRepository) DeleteHost(id int64) error {
	return r.hostRepo.Delete(id)
}

// DeleteHostByIP 根据 IP 删除主机
func (r *WebSharkRepository) DeleteHostByIP(ip string) error {
	return r.hostRepo.DeleteByCondition("ip = ?", ip)
}

// HostExists 检查主机是否存在
func (r *WebSharkRepository) HostExists(id int64) (bool, error) {
	return r.hostRepo.Exists(id)
}

// GetHostCount 获取主机总数
func (r *WebSharkRepository) GetHostCount() (int64, error) {
	return r.hostRepo.Count()
}

// ==================== Task 增删改查 ====================

// CreateTask 创建任务记录
func (r *WebSharkRepository) CreateTask(task *Task) error {
	return r.taskRepo.Create(task)
}

// GetTaskByID 根据 ID 获取任务
func (r *WebSharkRepository) GetTaskByID(id int64) (*Task, error) {
	return r.taskRepo.GetByID(id)
}

// ListTasks 获取任务列表（支持分页）
func (r *WebSharkRepository) ListTasks(page, pageSize int) ([]*Task, int64, error) {
	return r.taskRepo.List(page, pageSize, "id DESC")
}

// ListTasksByHostID 根据主机 ID 获取任务列表
func (r *WebSharkRepository) ListTasksByHostID(hostID int64, page, pageSize int) ([]*Task, int64, error) {
	return r.taskRepo.ListByCondition("host_id = ?", []interface{}{hostID}, page, pageSize, "id DESC")
}

// ListTasksByTaskGroupID 根据任务组 ID 获取任务列表
func (r *WebSharkRepository) ListTasksByTaskGroupID(taskGroupID int64, page, pageSize int) ([]*Task, int64, error) {
	return r.taskRepo.ListByCondition("task_group_id = ?", []interface{}{taskGroupID}, page, pageSize, "id DESC")
}

// UpdateTask 更新任务信息
func (r *WebSharkRepository) UpdateTask(task *Task) error {
	return r.taskRepo.Update(task)
}

// UpdateTaskFields 更新任务指定字段
func (r *WebSharkRepository) UpdateTaskFields(id int64, updates map[string]interface{}) error {
	return r.taskRepo.UpdateFields(id, updates)
}

// AppendTaskMessage 追加任务消息（使用 SQL CONCAT 原子追加）
func (r *WebSharkRepository) AppendTaskMessage(id int64, message string) error {
	return r.taskRepo.AppendField(id, "message", message)
}

// UpdateTaskStatusIfNot 条件更新任务状态：仅当当前状态不等于指定状态时更新
// 返回受影响的行数，可用于判断是否实际执行了更新
func (r *WebSharkRepository) UpdateTaskStatusIfNot(id int64, notStatus string, newStatus string) (int64, error) {
	result := r.db.Model(&Task{}).
		Where("id = ? AND status != ?", id, notStatus).
		Update("status", newStatus)
	return result.RowsAffected, result.Error
}

// UpdateTaskStopTime 更新任务停止时间
func (r *WebSharkRepository) UpdateTaskStopTime(id int64, stopAt time.Time) error {
	return r.taskRepo.UpdateField(id, "stop_at", stopAt)
}

// DeleteTask 删除任务
func (r *WebSharkRepository) DeleteTask(id int64) error {
	return r.taskRepo.Delete(id)
}

// TaskExists 检查任务是否存在
func (r *WebSharkRepository) TaskExists(id int64) (bool, error) {
	return r.taskRepo.Exists(id)
}

// GetTaskCount 获取任务总数
func (r *WebSharkRepository) GetTaskCount() (int64, error) {
	return r.taskRepo.Count()
}

// ==================== TaskGroup 增删改查 ====================

// CreateTaskGroup 创建任务组
func (r *WebSharkRepository) CreateTaskGroup(taskGroup *TaskGroup) error {
	return r.taskGroupRepo.Create(taskGroup)
}

// GetTaskGroupByID 根据 ID 获取任务组
func (r *WebSharkRepository) GetTaskGroupByID(id int64) (*TaskGroup, error) {
	return r.taskGroupRepo.GetByID(id)
}

// ListTaskGroups 获取任务组列表（支持分页）
func (r *WebSharkRepository) ListTaskGroups(page, pageSize int) ([]*TaskGroup, int64, error) {
	return r.taskGroupRepo.List(page, pageSize, "created_at DESC")
}

// UpdateTaskGroup 更新任务组信息
func (r *WebSharkRepository) UpdateTaskGroup(taskGroup *TaskGroup) error {
	return r.taskGroupRepo.Update(taskGroup)
}

// UpdateTaskGroupStopTime 更新任务组停止时间
func (r *WebSharkRepository) UpdateTaskGroupStopTime(id int64, stopAt time.Time) error {
	return r.taskGroupRepo.UpdateField(id, "stop_at", stopAt)
}

// DeleteTaskGroup 删除任务组
func (r *WebSharkRepository) DeleteTaskGroup(id int64) error {
	return r.taskGroupRepo.Delete(id)
}

// TaskGroupExists 检查任务组是否存在
func (r *WebSharkRepository) TaskGroupExists(id int64) (bool, error) {
	return r.taskGroupRepo.Exists(id)
}

// GetTaskGroupCount 获取任务组总数
func (r *WebSharkRepository) GetTaskGroupCount() (int64, error) {
	return r.taskGroupRepo.Count()
}

// ==================== Packet 增删改查 ====================

// CreatePacket 创建数据包记录
func (r *WebSharkRepository) CreatePacket(packet *Packet) error {
	return r.packetRepo.Create(packet)
}

// BatchCreatePackets 批量创建数据包记录
func (r *WebSharkRepository) BatchCreatePackets(packets []*Packet) error {
	return r.packetRepo.BatchCreate(packets, 100)
}

// GetPacketByID 根据 ID 获取数据包
func (r *WebSharkRepository) GetPacketByID(id int64) (*Packet, error) {
	return r.packetRepo.GetByID(id)
}

// ListPacketsByTaskID 根据任务 ID 获取数据包列表（支持分页）
func (r *WebSharkRepository) ListPacketsByTaskID(taskID int64, page, pageSize int) ([]*Packet, int64, error) {
	return r.packetRepo.ListByCondition("task_id = ?", []interface{}{taskID}, page, pageSize, "frame_number ASC")
}

// ListPacketsByProtocol 根据协议类型获取数据包列表
func (r *WebSharkRepository) ListPacketsByProtocol(taskID int64, protocol string, page, pageSize int) ([]*Packet, int64, error) {
	query := "task_id = ? AND protocol = ?"
	args := []interface{}{taskID, protocol}
	return r.packetRepo.ListByCondition(query, args, page, pageSize, "frame_number ASC")
}

// SearchPackets 搜索数据包（根据源地址、目的地址或协议）
func (r *WebSharkRepository) SearchPackets(taskID int64, keyword string, page, pageSize int) ([]*Packet, int64, error) {
	query := "task_id = ? AND (src LIKE ? OR dst LIKE ? OR protocol LIKE ?)"
	args := []interface{}{taskID, "%" + keyword + "%", "%" + keyword + "%", "%" + keyword + "%"}
	return r.packetRepo.ListByCondition(query, args, page, pageSize, "frame_number ASC")
}

// DeletePacketsByTaskID 根据任务 ID 删除所有数据包
func (r *WebSharkRepository) DeletePacketsByTaskID(taskID int64) error {
	return r.packetRepo.DeleteByCondition("task_id = ?", taskID)
}

// UpdatePacket 更新数据包信息
func (r *WebSharkRepository) UpdatePacket(packet *Packet) error {
	return r.packetRepo.Update(packet)
}

// GetPacketCountByTaskID 获取任务的数据包总数
func (r *WebSharkRepository) GetPacketCountByTaskID(taskID int64) (int64, error) {
	return r.packetRepo.CountByCondition("task_id = ?", taskID)
}

// ListPacketsByTaskIDAndFrameNumber 根据任务 ID 和帧号范围获取数据包列表
func (r *WebSharkRepository) ListPacketsByTaskIDAndFrameNumber(taskID int64, startFrameNumber, endFrameNumber int64, page, pageSize int) ([]*Packet, int64, error) {
	query := "task_id = ? AND frame_number >= ? AND frame_number <= ?"
	args := []interface{}{taskID, startFrameNumber, endFrameNumber}
	return r.packetRepo.ListByCondition(query, args, page, pageSize, "frame_number ASC")
}

// UpdatePacketContent 更新数据包的 Content 字段（通过 taskID 和 frameNumber 定位）
// 返回受影响的行数，如果为 0 表示记录不存在
func (r *WebSharkRepository) UpdatePacketContent(taskID int64, frameNumber int64, content string) (int64, error) {
	result := r.db.Model(&Packet{}).
		Where("task_id = ? AND frame_number = ?", taskID, frameNumber).
		Update("content", content)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetMaxPacketNoByTaskGroupID 获取任务组内最大的包序号
func (r *WebSharkRepository) GetMaxPacketNoByTaskGroupID(taskGroupID int64) (int64, error) {
	var maxNo int64
	// 通过子查询找到任务组内所有任务的包的最大 No 值
	err := r.db.Model(&Packet{}).
		Select("COALESCE(MAX(no), 0)").
		Where("task_id IN (SELECT id FROM task WHERE task_group_id = ?)", taskGroupID).
		Scan(&maxNo).Error
	if err != nil {
		return 0, err
	}
	return maxNo, nil
}

// AllocatePacketNoInTaskGroup 在任务组内分配唯一的包序号（使用数据库锁保证原子性）
func (r *WebSharkRepository) AllocatePacketNoInTaskGroup(taskGroupID int64) (int64, error) {
	var allocatedNo int64

	// 使用事务和行级锁确保原子性
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 查询当前最大值并锁定相关行
		var maxNo int64
		err := tx.Model(&Packet{}).
			Select("COALESCE(MAX(no), 0)").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("task_id IN (SELECT id FROM task WHERE task_group_id = ?)", taskGroupID).
			Scan(&maxNo).Error
		if err != nil {
			return err
		}

		allocatedNo = maxNo + 1
		return nil
	})

	if err != nil {
		return 0, err
	}

	return allocatedNo, nil
}

// ==================== Process 增删改查 ====================

// CreateProcess 创建进程记录
func (r *WebSharkRepository) CreateProcess(process *Process) error {
	return r.processRepo.Create(process)
}

// BatchCreateProcesses 批量创建进程记录
func (r *WebSharkRepository) BatchCreateProcesses(processes []*Process) error {
	return r.processRepo.BatchCreate(processes, 50)
}

// GetProcessByID 根据 ID 获取进程
func (r *WebSharkRepository) GetProcessByID(id int64) (*Process, error) {
	return r.processRepo.GetByID(id)
}

// ListProcessesByTaskID 根据任务 ID 获取进程列表
func (r *WebSharkRepository) ListProcessesByTaskID(taskID int64) ([]*Process, error) {
	return r.processRepo.FindByCondition("task_id = ?", []interface{}{taskID}, "id ASC")
}

// GetProcessByPid 根据 PID 获取进程
func (r *WebSharkRepository) GetProcessByPid(pid int64) (*Process, error) {
	return r.processRepo.FirstByCondition("pid = ?", pid)
}

// UpdateProcess 更新进程信息
func (r *WebSharkRepository) UpdateProcess(process *Process) error {
	return r.processRepo.Update(process)
}

// UpdateProcessFields 更新进程指定字段
func (r *WebSharkRepository) UpdateProcessFields(id int64, updates map[string]interface{}) error {
	return r.processRepo.UpdateFields(id, updates)
}

// AppendProcessMessage 追加进程消息（使用 SQL CONCAT 原子追加）
func (r *WebSharkRepository) AppendProcessMessage(id int64, message string) error {
	return r.processRepo.AppendField(id, "message", message)
}

// DeleteProcess 删除进程记录
func (r *WebSharkRepository) DeleteProcess(id int64) error {
	return r.processRepo.Delete(id)
}

// DeleteProcessesByTaskID 根据任务 ID 删除所有进程记录
func (r *WebSharkRepository) DeleteProcessesByTaskID(taskID int64) error {
	return r.processRepo.DeleteByCondition("task_id = ?", taskID)
}

// ProcessExists 检查进程是否存在
func (r *WebSharkRepository) ProcessExists(id int64) (bool, error) {
	return r.processRepo.Exists(id)
}

// GetProcessCountByTaskID 获取任务的进程总数
func (r *WebSharkRepository) GetProcessCountByTaskID(taskID int64) (int64, error) {
	return r.processRepo.CountByCondition("task_id = ?", taskID)
}
