# Service 层互相调用指南

## 📋 概述

在 Go 项目中，Service 层之间经常需要互相调用。本文介绍几种常见的实现方式。

---

## 🎯 三种常见方式

### 方式 1：依赖注入（推荐）⭐

通过构造函数注入需要的 Service 实例。

#### ✅ 优点
- **清晰的依赖关系**：一眼就能看出依赖哪些 Service
- **易于测试**：可以方便地 Mock 依赖的 Service
- **无循环依赖**：编译期就能发现循环依赖问题
- **符合 SOLID 原则**：依赖倒置原则

#### ❌ 缺点
- 初始化时需要手动管理依赖顺序
- Service 较多时，构造函数参数可能很长

#### 示例代码

```go
// task_group_service.go
type TaskGroupService struct {
    repo    *gorm.WebSharkRepository
    taskSvc *TaskService  // ✅ 注入 TaskService
}

func NewTaskGroupService(repo *gorm.WebSharkRepository, taskSvc *TaskService) *TaskGroupService {
    return &TaskGroupService{
        repo:    repo,
        taskSvc: taskSvc,
    }
}

// 使用示例
func (s *TaskGroupService) GetTaskGroupWithTasks(id int64) (*TaskGroupWithTasksResponse, error) {
    // ✅ 直接调用其他 Service 的方法
    tasksResp, err := s.taskSvc.ListTasksByTaskGroupID(&ListTasksByTaskGroupIDRequest{
        TaskGroupID: id,
        Page:        1,
        PageSize:    10,
    })
    if err != nil {
        return nil, err
    }
    
    return &TaskGroupWithTasksResponse{
        Tasks: tasksResp,
    }, nil
}
```

---

### 方式 2：通过 Repository 层共享数据

Service 之间不直接调用，而是通过共享的 Repository 层访问数据。

#### ✅ 优点
- **松耦合**：Service 之间没有直接依赖
- **职责清晰**：每个 Service 只负责自己的业务逻辑
- **易于维护**：修改一个 Service 不影响其他 Service

#### ❌ 缺点
- **无法复用业务逻辑**：如果另一个 Service 有复杂的业务逻辑，无法直接调用
- **可能重复代码**：多个 Service 可能需要实现相似的逻辑

#### 示例代码

```go
// capture_service.go
type CaptureService struct {
    repo *gorm.WebSharkRepository
}

func (s *CaptureService) GetCaptureStats(hostID int64) (*CaptureStatsResponse, error) {
    // ✅ 直接调用 Repository，而不是调用其他 Service
    host, _ := s.repo.GetHostByID(hostID)
    tasks, _, _ := s.repo.ListTasksByHostID(hostID, 1, 100)
    
    return &CaptureStatsResponse{
        Host:  host,
        Tasks: tasks,
    }, nil
}
```

---

### 方式 3：创建综合协调 Service

创建一个高层 Service 来协调多个底层 Service。

#### ✅ 优点
- **单一职责**：每个底层 Service 保持专注
- **流程清晰**：复杂业务流程在一个地方管理
- **易于扩展**：新增流程只需修改协调 Service

#### ❌ 缺点
- **增加一层抽象**：多了一层需要维护的代码
- **可能过度设计**：简单场景不需要这么复杂

#### 示例代码

```go
// orchestration_service.go
type OrchestrationService struct {
    hostSvc      *HostService
    taskSvc      *TaskService
    taskGroupSvc *TaskGroupService
}

func NewOrchestrationService(
    hostSvc *HostService,
    taskSvc *TaskService,
    taskGroupSvc *TaskGroupService,
) *OrchestrationService {
    return &OrchestrationService{
        hostSvc:      hostSvc,
        taskSvc:      taskSvc,
        taskGroupSvc: taskGroupSvc,
    }
}

// StartCapture 复杂的业务流程
func (s *OrchestrationService) StartCapture(req *StartCaptureRequest) error {
    // 步骤 1：调用 HostService
    host, err := s.hostSvc.GetOrCreateHost(req.HostName)
    if err != nil {
        return err
    }
    
    // 步骤 2：调用 TaskGroupService
    taskGroup, err := s.taskGroupSvc.CreateTaskGroup(&CreateTaskGroupRequest{
        HostID: host.ID,
    })
    if err != nil {
        return err
    }
    
    // 步骤 3：调用 TaskService
    _, err = s.taskSvc.CreateTask(&CreateTaskRequest{
        HostID:      host.ID,
        TaskGroupId: taskGroup.ID,
    })
    if err != nil {
        return err
    }
    
    return nil
}
```

---

## 🔧 在 main.go 中初始化

### 关键：注意初始化顺序！

```go
func initServices(dbConfig *config.DBConfig) {
    // 步骤 1：创建 Repository
    repo := gorm.NewWebSharkRepository(dbConfig)
    
    // 步骤 2：创建基础 Service（没有依赖其他 Service）
    hostSvc := service.NewHostService(repo)
    taskSvc := service.NewTaskService(repo)
    
    // 步骤 3：创建依赖其他 Service 的 Service
    taskGroupSvc := service.NewTaskGroupService(repo, taskSvc)
    
    // 步骤 4：创建综合协调 Service
    captureSvc := service.NewCaptureService(repo, hostSvc, taskSvc, taskGroupSvc)
    
    // 步骤 5：创建 Handler
    hostHandler := web.NewHostHandler(hostSvc)
    taskHandler := web.NewTaskHandler(taskSvc)
    taskGroupHandler := web.NewTaskGroupHandler(taskGroupSvc)
    captureHandler := web.NewCaptureHandler(captureSvc)
    
    // 步骤 6：注册路由
    router := gin.Default()
    hostHandler.RegisterRoutes(router)
    taskHandler.RegisterRoutes(router)
    // ...
}
```

---

## ⚠️ 常见问题

### 1. 循环依赖

❌ **错误示例**：
```go
// Service A 依赖 Service B
type ServiceA struct {
    svcB *ServiceB
}

// Service B 依赖 Service A
type ServiceB struct {
    svcA *ServiceA  // ❌ 循环依赖！编译错误
}
```

✅ **解决方案**：
- 提取公共逻辑到第三个 Service C
- 或者通过 Repository 层共享数据

---

### 2. 初始化顺序错误

❌ **错误示例**：
```go
taskGroupSvc := service.NewTaskGroupService(repo, taskSvc)  // taskSvc 还未初始化
taskSvc := service.NewTaskService(repo)
```

✅ **正确示例**：
```go
taskSvc := service.NewTaskService(repo)                     // 先初始化被依赖的
taskGroupSvc := service.NewTaskGroupService(repo, taskSvc)  // 再初始化依赖它的
```

---

### 3. 测试时如何 Mock

```go
// 创建 Mock Service
type MockTaskService struct{}

func (m *MockTaskService) ListTasksByTaskGroupID(req *ListTasksByTaskGroupIDRequest) (*entity.PageResponse[*gorm.Task], error) {
    // 返回模拟数据
    return &entity.PageResponse[*gorm.Task]{
        Data: []*gorm.Task{{ID: 1, TaskName: "test"}},
        Total: 1,
    }, nil
}

// 在测试中使用
func TestTaskGroupService(t *testing.T) {
    mockTaskSvc := &MockTaskService{}
    taskGroupSvc := NewTaskGroupService(mockRepo, mockTaskSvc)
    
    // 测试代码...
}
```

---

## 📊 选择建议

| 场景 | 推荐方式 |
|------|---------|
| 简单的数据查询/聚合 | 方式 2：通过 Repository |
| 需要复用业务逻辑 | 方式 1：依赖注入 |
| 复杂的业务流程编排 | 方式 3：协调 Service |
| 微服务架构 | 方式 1 + 方式 3 组合 |

---

## 🎓 最佳实践

1. **优先使用依赖注入**：清晰、可测试、易维护
2. **避免循环依赖**：如果发现循环依赖，重新设计 Service 边界
3. **保持 Service 单一职责**：每个 Service 只负责一个领域
4. **不要过度分层**：简单场景直接用 Repository 即可
5. **文档化依赖关系**：在注释中说明为什么需要这个依赖
6. **使用接口解耦**（高级）：对于大型项目，可以用接口定义 Service 契约

---

## 📖 参考示例

查看以下文件获取完整示例：
- `task_group_service_example.go` - TaskGroupService 调用 TaskService
- `capture_service_example.go` - CaptureService 协调多个 Service
- `SERVICE_DEPENDENCY_EXAMPLE.md` - 本文档
