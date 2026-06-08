# Service 层互相调用完整示例

## 📋 概述

在 Go 项目中，Service 层之间经常需要互相调用。本文通过**实际可运行的代码示例**展示如何实现。

---

## 🎯 核心原则

### ✅ 推荐：依赖注入（Dependency Injection）

通过构造函数注入需要的 Service 实例，这是最常用、最推荐的方式。

---

## 💡 完整示例

### 场景描述

假设我们有一个**任务管理系统**，包含以下 Service：
- **HostService**: 管理主机
- **TaskService**: 管理任务
- **TaskGroupService**: 管理任务组
- **CaptureService**: 抓包综合服务（需要调用其他 Service）

---

### 示例 1：简单的 Service 间调用

**文件**: `task_group_service.go`

```go
package service

import (
	"fmt"
	"time"
	"webshark/internal/entity"
	"webshark/internal/gorm"
)

// TaskGroupService TaskGroup 服务
type TaskGroupService struct {
	repo *gorm.WebSharkRepository
	
	// ✅ 注入其他 Service，实现跨服务调用
	taskSvc *TaskService
}

// NewTaskGroupService 创建 TaskGroup 服务实例
func NewTaskGroupService(repo *gorm.WebSharkRepository, taskSvc *TaskService) *TaskGroupService {
	return &TaskGroupService{
		repo:    repo,
		taskSvc: taskSvc, // ✅ 通过构造函数注入 TaskService
	}
}

// CreateTaskGroupRequest 创建任务组请求
type CreateTaskGroupRequest struct {
	GroupName string `json:"groupName" binding:"required"`
}

// CreateTaskGroup 创建任务组
func (s *TaskGroupService) CreateTaskGroup(req *CreateTaskGroupRequest) (*gorm.TaskGroup, error) {
	taskGroup := &gorm.TaskGroup{
		TaskGroupName: req.GroupName,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateTaskGroup(taskGroup); err != nil {
		return nil, fmt.Errorf("failed to create task group: %w", err)
	}

	return taskGroup, nil
}

// GetTaskGroupWithTasksRequest 获取任务组及其所有任务
type GetTaskGroupWithTasksRequest struct {
	ID       int64 `form:"id" json:"id" binding:"required"`
	Page     int   `form:"page" json:"page"`
	PageSize int   `form:"pageSize" json:"pageSize"`
}

// TaskGroupWithTasksResponse 任务组及任务响应
type TaskGroupWithTasksResponse struct {
	TaskGroup *gorm.TaskGroup                  `json:"taskGroup"`
	Tasks     *entity.PageResponse[*gorm.Task] `json:"tasks"`
}

// GetTaskGroupWithTasks 获取任务组及其所有任务
func (s *TaskGroupService) GetTaskGroupWithTasks(req *GetTaskGroupWithTasksRequest) (*TaskGroupWithTasksResponse, error) {
	// ✅ 步骤 1：获取任务组信息
	taskGroup, err := s.repo.GetTaskGroupByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task group: %w", err)
	}

	// ✅ 步骤 2：调用 TaskService 获取该任务组下的所有任务
	tasksResp, err := s.taskSvc.ListTasksByTaskGroupID(&ListTasksByTaskGroupIDRequest{
		TaskGroupID: req.ID,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return &TaskGroupWithTasksResponse{
		TaskGroup: taskGroup,
		Tasks:     tasksResp,
	}, nil
}

// DeleteTaskGroupWithTasksRequest 删除任务组及其所有任务
type DeleteTaskGroupWithTasksRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// DeleteTaskGroupWithTasks 删除任务组及其所有任务
func (s *TaskGroupService) DeleteTaskGroupWithTasks(req *DeleteTaskGroupWithTasksRequest) error {
	// ✅ 先获取任务组下的所有任务
	tasks, _, err := s.repo.ListTasksByTaskGroupID(req.ID, 1, 1000)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// ✅ 逐个删除任务
	for _, task := range tasks {
		if err := s.repo.DeleteTask(task.ID); err != nil {
			return fmt.Errorf("failed to delete task %d: %w", task.ID, err)
		}
	}

	// ✅ 最后删除任务组
	if err := s.repo.DeleteTaskGroup(req.ID); err != nil {
		return fmt.Errorf("failed to delete task group: %w", err)
	}

	return nil
}
```

---

### 示例 2：复杂的业务流程编排

**文件**: `capture_service.go`

```go
package service

import (
	"fmt"
	"time"
	"webshark/internal/gorm"
)

// CaptureService 抓包综合服务（演示多 Service 协作）
type CaptureService struct {
	repo *gorm.WebSharkRepository
	
	// ✅ 注入多个 Service，实现复杂的业务流程
	hostSvc      *HostService
	taskSvc      *TaskService
	taskGroupSvc *TaskGroupService
}

// NewCaptureService 创建抓包服务实例
func NewCaptureService(
	repo *gorm.WebSharkRepository,
	hostSvc *HostService,
	taskSvc *TaskService,
	taskGroupSvc *TaskGroupService,
) *CaptureService {
	return &CaptureService{
		repo:         repo,
		hostSvc:      hostSvc,
		taskSvc:      taskSvc,
		taskGroupSvc: taskGroupSvc,
	}
}

// StartCaptureRequest 开始抓包请求
type StartCaptureRequest struct {
	HostName      string   `json:"hostName" binding:"required"`
	TaskGroupName string   `json:"taskGroupName"`
	TaskName      string   `json:"taskName" binding:"required"`
	Interfaces    []string `json:"interfaces"`
	OnlyCapture   bool     `json:"onlyCapture"`
	ParseDetail   bool     `json:"parseDetail"`
	FilePath      string   `json:"filePath" binding:"required"`
	BpfFilter     string   `json:"bpfFilter"`
}

// StartCaptureResponse 开始抓包响应
type StartCaptureResponse struct {
	HostID      int64  `json:"hostId"`
	TaskGroupID int64  `json:"taskGroupId"`
	TaskID      int64  `json:"taskId"`
	Message     string `json:"message"`
}

// StartCapture 开始抓包（复杂业务流程）
func (s *CaptureService) StartCapture(req *StartCaptureRequest) (*StartCaptureResponse, error) {
	// ✅ 步骤 1：调用 HostService 获取或创建主机
	host, err := s.hostSvc.GetOrCreateHost(req.HostName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create host: %w", err)
	}

	// ✅ 步骤 2：如果指定了任务组，调用 TaskGroupService 创建任务组
	var taskGroupID int64
	if req.TaskGroupName != "" {
		taskGroup, err := s.taskGroupSvc.CreateTaskGroup(&CreateTaskGroupRequest{
			GroupName: req.TaskGroupName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create task group: %w", err)
		}
		taskGroupID = taskGroup.ID
	}

	// ✅ 步骤 3：调用 TaskService 创建任务
	task, err := s.taskSvc.CreateTask(&CreateTaskRequest{
		TaskName:    req.TaskName,
		HostID:      host.ID,
		Interfaces:  req.Interfaces,
		OnlyCapture: req.OnlyCapture,
		ParseDetail: req.ParseDetail,
		FilePath:    req.FilePath,
		BpfFilter:   req.BpfFilter,
		TaskGroupId: taskGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &StartCaptureResponse{
		HostID:      host.ID,
		TaskGroupID: taskGroupID,
		TaskID:      task.ID,
		Message:     "Capture started successfully",
	}, nil
}

// GetCaptureStatsRequest 获取抓包统计请求
type GetCaptureStatsRequest struct {
	HostID int64 `form:"hostId" json:"hostId"`
}

// CaptureStatsResponse 抓包统计响应
type CaptureStatsResponse struct {
	Host           *gorm.Host        `json:"host"`
	TaskGroups     []*gorm.TaskGroup `json:"taskGroups"`
	Tasks          []*gorm.Task      `json:"tasks"`
	TotalPackets   int64             `json:"totalPackets"`
	TotalProcesses int64             `json:"totalProcesses"`
}

// GetCaptureStats 获取抓包统计（聚合多个 Service 的数据）
func (s *CaptureService) GetCaptureStats(req *GetCaptureStatsRequest) (*CaptureStatsResponse, error) {
	// ✅ 调用 HostService 获取主机信息
	host, err := s.hostSvc.GetHostByID(req.HostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	// ✅ 直接调用 Repository 获取其他数据
	taskGroups, _, err := s.repo.ListTaskGroups(1, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list task groups: %w", err)
	}

	tasks, _, err := s.repo.ListTasksByHostID(req.HostID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// ✅ 统计数据包总数
	var totalPackets int64
	for _, task := range tasks {
		count, _ := s.repo.GetPacketCountByTaskID(task.ID)
		totalPackets += count
	}

	// ✅ 统计进程总数
	var totalProcesses int64
	for _, task := range tasks {
		processes, _ := s.repo.ListProcessesByTaskID(task.ID)
		totalProcesses += int64(len(processes))
	}

	return &CaptureStatsResponse{
		Host:           host,
		TaskGroups:     taskGroups,
		Tasks:          tasks,
		TotalPackets:   totalPackets,
		TotalProcesses: totalProcesses,
	}, nil
}
```

---

### 示例 3：在 main.go 中初始化

**关键：注意初始化顺序！**

```go
package main

import (
	"webshark/internal/config"
	"webshark/internal/gorm"
	"webshark/internal/handler/web"
	"webshark/internal/service"
)

func initServices(dbConfig *config.DBConfig) {
	// ✅ 步骤 1：创建 Repository
	repo := gorm.NewWebSharkRepository(dbConfig)
	
	// ✅ 步骤 2：创建基础 Service（没有依赖其他 Service）
	hostSvc := service.NewHostService(repo)
	taskSvc := service.NewTaskService(repo)
	
	// ✅ 步骤 3：创建依赖其他 Service 的 Service
	taskGroupSvc := service.NewTaskGroupService(repo, taskSvc)
	
	// ✅ 步骤 4：创建综合协调 Service
	captureSvc := service.NewCaptureService(repo, hostSvc, taskSvc, taskGroupSvc)
	
	// ✅ 步骤 5：创建 Handler
	hostHandler := web.NewHostHandler(hostSvc)
	taskHandler := web.NewTaskHandler(taskSvc)
	taskGroupHandler := web.NewTaskGroupHandler(taskGroupSvc)
	captureHandler := web.NewCaptureHandler(captureSvc)
	
	// ✅ 步骤 6：注册路由
	router := gin.Default()
	
	// 注册 Host 路由
	router.GET("/api/v1/hosts/:id", hostHandler.GetHost)
	router.POST("/api/v1/hosts", hostHandler.CreateHost)
	
	// 注册 Task 路由
	router.GET("/api/v1/tasks/:id", taskHandler.GetTask)
	router.POST("/api/v1/tasks", taskHandler.CreateTask)
	
	// 注册 TaskGroup 路由
	router.GET("/api/v1/task-groups/:id", taskGroupHandler.GetTaskGroup)
	router.GET("/api/v1/task-groups/:id/with-tasks", taskGroupHandler.GetTaskGroupWithTasks)
	router.POST("/api/v1/task-groups", taskGroupHandler.CreateTaskGroup)
	
	// 注册 Capture 路由
	router.POST("/api/v1/captures/start", captureHandler.StartCapture)
	router.GET("/api/v1/captures/stats", captureHandler.GetCaptureStats)
	
	// 启动服务器
	router.Run(":8080")
}

func main() {
	// 加载配置
	dbConfig := config.LoadDBConfig()
	
	// 初始化并启动
	initServices(dbConfig)
}
```

---

## 🔍 关键点解析

### 1. 依赖注入的方向

```
CaptureService (高层业务)
    ↓ 依赖
TaskGroupService, TaskService, HostService (中层业务)
    ↓ 依赖
WebSharkRepository (数据访问层)
    ↓ 依赖
Database
```

**规则**：依赖方向应该是**从上到下**，不能反向依赖！

---

### 2. 初始化顺序

```go
// ❌ 错误顺序 - taskSvc 还未初始化就使用
taskGroupSvc := NewTaskGroupService(repo, taskSvc)
taskSvc := NewTaskService(repo)

// ✅ 正确顺序 - 先初始化被依赖的
taskSvc := NewTaskService(repo)                    // 先
taskGroupSvc := NewTaskGroupService(repo, taskSvc) // 后
```

---

### 3. 避免循环依赖

```go
// ❌ 循环依赖 - 编译错误！
type ServiceA struct {
    svcB *ServiceB
}

type ServiceB struct {
    svcA *ServiceA  // ❌ A 依赖 B，B 又依赖 A
}

// ✅ 解决方案 - 提取公共逻辑到 ServiceC
type ServiceC struct {
    // 公共逻辑
}

type ServiceA struct {
    svcC *ServiceC  // ✅ 都依赖 C，不互相依赖
}

type ServiceB struct {
    svcC *ServiceC
}
```

---

## 📊 三种常见方式对比

| 方式 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| **依赖注入** | 需要复用业务逻辑 | 清晰、可测试、易维护 | 需要注意初始化顺序 |
| **通过 Repository** | 简单的数据查询/聚合 | 松耦合、无依赖 | 无法复用业务逻辑 |
| **协调 Service** | 复杂的业务流程编排 | 单一职责、流程清晰 | 增加一层抽象 |

---

## 🎓 最佳实践

### ✅ Do's

1. **优先使用依赖注入**
   ```go
   func NewMyService(repo *Repo, otherSvc *OtherService) *MyService {
       return &MyService{repo: repo, otherSvc: otherSvc}
   }
   ```

2. **保持 Service 单一职责**
   ```go
   // ✅ HostService 只负责主机相关
   type HostService struct { repo *Repo }
   
   // ✅ TaskService 只负责任务相关
   type TaskService struct { repo *Repo }
   ```

3. **文档化依赖关系**
   ```go
   type CaptureService struct {
       // 依赖: HostService, TaskService, TaskGroupService
       hostSvc *HostService
       taskSvc *TaskService
   }
   ```

### ❌ Don'ts

1. **不要循环依赖**
   ```go
   // ❌ A 依赖 B，B 又依赖 A
   ```

2. **不要在 Service 中直接创建其他 Service**
   ```go
   // ❌ 错误
   func NewServiceA() *ServiceA {
       svcB := NewServiceB()  // 不应该在这里创建
       return &ServiceA{svcB: svcB}
   }
   
   // ✅ 正确
   func NewServiceA(svcB *ServiceB) *ServiceA {
       return &ServiceA{svcB: svcB}
   }
   ```

3. **不要过度分层**
   ```go
   // ❌ 简单查询不需要调用其他 Service
   func (s *TaskService) GetTask(id int64) (*Task, error) {
       return s.repo.GetTaskByID(id)  // ✅ 直接调用 Repository
   }
   ```

---

## 🧪 如何测试

### 单元测试示例

```go
package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// Mock TaskService
type MockTaskService struct{}

func (m *MockTaskService) ListTasksByTaskGroupID(req *ListTasksByTaskGroupIDRequest) (*entity.PageResponse[*gorm.Task], error) {
	// 返回模拟数据
	return &entity.PageResponse[*gorm.Task]{
		Data:  []*gorm.Task{{ID: 1, TaskName: "test"}},
		Total: 1,
	}, nil
}

// 测试 TaskGroupService
func TestTaskGroupService_GetTaskGroupWithTasks(t *testing.T) {
	// 准备
	mockTaskSvc := &MockTaskService{}
	taskGroupSvc := NewTaskGroupService(mockRepo, mockTaskSvc)
	
	// 执行
	req := &GetTaskGroupWithTasksRequest{
		ID:       1,
		Page:     1,
		PageSize: 10,
	}
	resp, err := taskGroupSvc.GetTaskGroupWithTasks(req)
	
	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Tasks.Data))
}
```

---

## 📖 总结

### 核心要点

1. **依赖注入是最常用的方式**
   - 通过构造函数注入
   - 清晰的依赖关系
   - 易于测试和维护

2. **注意初始化顺序**
   - 先初始化被依赖的 Service
   - 再初始化依赖它的 Service

3. **避免循环依赖**
   - 如果发现循环依赖，重新设计 Service 边界
   - 可以提取公共逻辑到第三个 Service

4. **选择合适的方案**
   - 简单数据查询 → 直接用 Repository
   - 复用业务逻辑 → 依赖注入
   - 复杂流程编排 → 协调 Service

---

## 🔗 相关文档

- [SERVICE_DEPENDENCY_GUIDE.md](./SERVICE_DEPENDENCY_GUIDE.md) - 详细的理论说明
- [REFACTORING_SUMMARY.md](../gorm/REFACTORING_SUMMARY.md) - Repository 泛型重构
- [TASK_HANDLER_GUIDE.md](../handler/web/TASK_HANDLER_GUIDE.md) - Handler 层使用指南
