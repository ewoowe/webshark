package service

import (
	"fmt"
	"time"
	"webshark/internal/entity"
	"webshark/internal/gorm"
)

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	TaskName        string   `json:"taskName" binding:"required" label:"任务名称"`
	HostID          int64    `json:"hostId" binding:"required" label:"主机ID"`
	Interfaces      []string `json:"interfaces"`
	OnlyCapture     bool     `json:"onlyCapture"`
	ParseDetail     bool     `json:"parseDetail"`
	DetailFormat    string   `json:"detailFormat"`
	BpfFilter       string   `json:"bpfFilter"`
	WiresharkFilter string   `json:"wiresharkFilter"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	ID              int64    `json:"id" binding:"required" label:"任务ID"`
	TaskName        string   `json:"taskName"`
	Interfaces      []string `json:"interfaces"`
	OnlyCapture     *bool    `json:"onlyCapture"`
	ParseDetail     *bool    `json:"parseDetail"`
	DetailFormat    string   `json:"detailFormat"`
	BpfFilter       string   `json:"bpfFilter"`
	WiresharkFilter string   `json:"wiresharkFilter"`
}

// CreateTask 创建任务
func CreateTask(req *CreateTaskRequest) (*gorm.Task, error) {
	task := &gorm.Task{
		TaskName:        req.TaskName,
		HostID:          req.HostID,
		Interfaces:      req.Interfaces,
		OnlyCapture:     req.OnlyCapture,
		ParseDetail:     req.ParseDetail,
		DetailFormat:    req.DetailFormat,
		BpfFilter:       req.BpfFilter,
		WiresharkFilter: req.WiresharkFilter,
	}

	if err := gorm.Repo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetTaskByID 根据 ID 获取任务
func GetTaskByID(id int64) (*gorm.Task, error) {
	task, err := gorm.Repo.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// ListTasksRequest 获取任务列表请求
type ListTasksRequest struct {
	entity.PageRequest
}

// ListTasks 获取任务列表
func ListTasks(req *ListTasksRequest) (*entity.PageResponse[*gorm.Task], error) {
	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	tasks, total, err := gorm.Repo.ListTasks(page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return entity.NewPageResponse(tasks, total, page, pageSize), nil
}

// ListTasksByHostIDRequest 根据主机 ID 获取任务列表请求
type ListTasksByHostIDRequest struct {
	HostID int64 `form:"hostId" json:"hostId" binding:"required" label:"主机ID"`
	entity.PageRequest
}

// ListTasksByHostID 根据主机 ID 获取任务列表
func ListTasksByHostID(req *ListTasksByHostIDRequest) (*entity.PageResponse[*gorm.Task], error) {
	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	tasks, total, err := gorm.Repo.ListTasksByHostID(req.HostID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by host: %w", err)
	}

	return entity.NewPageResponse(tasks, total, page, pageSize), nil
}

// ListTasksByTaskGroupIDRequest 根据任务组 ID 获取任务列表请求
type ListTasksByTaskGroupIDRequest struct {
	TaskGroupID int64 `form:"taskGroupId" json:"taskGroupId" binding:"required" label:"任务组ID"`
	entity.PageRequest
}

// ListTasksByTaskGroupID 根据任务组 ID 获取任务列表
func ListTasksByTaskGroupID(req *ListTasksByTaskGroupIDRequest) (*entity.PageResponse[*gorm.Task], error) {
	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	tasks, total, err := gorm.Repo.ListTasksByTaskGroupID(req.TaskGroupID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by task group: %w", err)
	}

	return entity.NewPageResponse(tasks, total, page, pageSize), nil
}

// UpdateTask 更新任务
func UpdateTask(req *UpdateTaskRequest) (*gorm.Task, error) {
	// 先获取现有记录
	task, err := gorm.Repo.GetTaskByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// 更新字段
	if req.TaskName != "" {
		task.TaskName = req.TaskName
	}
	if req.Interfaces != nil {
		task.Interfaces = gorm.StringArray(req.Interfaces)
	}
	if req.OnlyCapture != nil {
		task.OnlyCapture = *req.OnlyCapture
	}
	if req.ParseDetail != nil {
		task.ParseDetail = *req.ParseDetail
	}
	if req.DetailFormat != "" {
		task.DetailFormat = req.DetailFormat
	}
	if req.BpfFilter != "" {
		task.BpfFilter = req.BpfFilter
	}
	if req.WiresharkFilter != "" {
		task.WiresharkFilter = req.WiresharkFilter
	}

	if err := gorm.Repo.UpdateTask(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// StopTask 停止任务（更新停止时间）
func StopTask(id int64) error {
	if err := gorm.Repo.UpdateTaskStopTime(id, time.Now()); err != nil {
		return fmt.Errorf("failed to stop task: %w", err)
	}

	return nil
}

// DeleteTask 删除任务
func DeleteTask(id int64) error {
	if err := gorm.Repo.DeleteTask(id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}
