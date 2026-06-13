package web

import (
	"webshark/internal/entity"
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateTask 创建任务
func CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, &req)
		return
	}

	task, err := service.CreateTask(&req)
	if err != nil {
		InternalError(c)
		return
	}

	SuccessWithMsg(c, task, "Task created successfully")
}

// GetTask 获取单个任务
func GetTask(c *gin.Context) {
	id, ok := ParseIntParam(c, "id")
	if !ok {
		return
	}

	task, err := service.GetTaskByID(id)
	if err != nil {
		NotFound(c)
		return
	}

	Success(c, task)
}

// ListTasks 获取任务列表
func ListTasks(c *gin.Context) {
	page, pageSize := ParsePageParams(c)

	req := &service.ListTasksRequest{
		PageRequest: entity.PageRequest{
			Page:     page,
			PageSize: pageSize,
		},
	}

	resp, err := service.ListTasks(req)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, resp)
}

// ListTasksByHostID 根据主机 ID 获取任务列表
func ListTasksByHostID(c *gin.Context) {
	hostIDStr := c.Query("hostId")
	if hostIDStr == "" {
		BadRequest(c, "Missing hostId parameter")
		return
	}

	hostID, ok := ParseIntParamFromQuery(c, hostIDStr)
	if !ok {
		return
	}

	page, pageSize := ParsePageParams(c)

	req := &service.ListTasksByHostIDRequest{
		HostID: hostID,
		PageRequest: entity.PageRequest{
			Page:     page,
			PageSize: pageSize,
		},
	}

	resp, err := service.ListTasksByHostID(req)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, resp)
}

// ListTasksByTaskGroupID 根据任务组 ID 获取任务列表
func ListTasksByTaskGroupID(c *gin.Context) {
	taskGroupIDStr := c.Query("taskGroupId")
	if taskGroupIDStr == "" {
		BadRequest(c, "Missing taskGroupId parameter")
		return
	}

	taskGroupID, ok := ParseIntParamFromQuery(c, taskGroupIDStr)
	if !ok {
		return
	}

	page, pageSize := ParsePageParams(c)

	req := &service.ListTasksByTaskGroupIDRequest{
		TaskGroupID: taskGroupID,
		PageRequest: entity.PageRequest{
			Page:     page,
			PageSize: pageSize,
		},
	}

	resp, err := service.ListTasksByTaskGroupID(req)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, resp)
}

// UpdateTask 更新任务
func UpdateTask(c *gin.Context) {
	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, &req)
		return
	}

	task, err := service.UpdateTask(&req)
	if err != nil {
		InternalError(c)
		return
	}

	SuccessWithMsg(c, task, "Task updated successfully")
}

// StopTask 停止任务
func StopTask(c *gin.Context) {
	id, ok := ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := service.StopTask(id); err != nil {
		InternalError(c)
		return
	}

	SuccessWithMsg(c, nil, "Task stopped successfully")
}

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	id, ok := ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := service.DeleteTask(id); err != nil {
		InternalError(c)
		return
	}

	SuccessWithMsg(c, nil, "Task deleted successfully")
}

// ParseIntParamFromQuery 从查询参数解析整数
func ParseIntParamFromQuery(c *gin.Context, param string) (int64, bool) {
	value, err := ParseStringToInt64(param)
	if err != nil {
		BadRequest(c, "Invalid parameter")
		return 0, false
	}
	return value, true
}
