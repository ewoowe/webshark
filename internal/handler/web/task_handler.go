package web

import (
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// TaskHandler Task 处理器
type TaskHandler struct {
	*BaseHandler[*service.TaskService]
}

// NewTaskHandler 创建 Task 处理器
func NewTaskHandler(taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{
		BaseHandler: NewBaseHandler[*service.TaskService](taskSvc),
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	task, err := h.GetService().CreateTask(&req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, task, "Task created successfully")
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := h.ParseIntParam(c, "id")
	if !ok {
		return
	}

	task, err := h.GetService().GetTaskByID(id)
	if err != nil {
		h.NotFound(c, err)
		return
	}

	h.Success(c, task)
}

// ListTasks 获取任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, pageSize := h.ParsePageParams(c)

	req := &service.ListTasksRequest{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.GetService().ListTasks(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.Success(c, resp)
}

// ListTasksByHostID 根据主机 ID 获取任务列表
func (h *TaskHandler) ListTasksByHostID(c *gin.Context) {
	hostIDStr := c.Query("hostId")
	if hostIDStr == "" {
		h.BadRequest(c, "Missing hostId parameter")
		return
	}

	hostID, ok := h.ParseIntParamFromQuery(c, hostIDStr)
	if !ok {
		return
	}

	page, pageSize := h.ParsePageParams(c)

	req := &service.ListTasksByHostIDRequest{
		HostID:   hostID,
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.GetService().ListTasksByHostID(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.Success(c, resp)
}

// ListTasksByTaskGroupID 根据任务组 ID 获取任务列表
func (h *TaskHandler) ListTasksByTaskGroupID(c *gin.Context) {
	taskGroupIDStr := c.Query("taskGroupId")
	if taskGroupIDStr == "" {
		h.BadRequest(c, "Missing taskGroupId parameter")
		return
	}

	taskGroupID, ok := h.ParseIntParamFromQuery(c, taskGroupIDStr)
	if !ok {
		return
	}

	page, pageSize := h.ParsePageParams(c)

	req := &service.ListTasksByTaskGroupIDRequest{
		TaskGroupID: taskGroupID,
		Page:        page,
		PageSize:    pageSize,
	}

	resp, err := h.GetService().ListTasksByTaskGroupID(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.Success(c, resp)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	task, err := h.GetService().UpdateTask(&req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, task, "Task updated successfully")
}

// StopTask 停止任务
func (h *TaskHandler) StopTask(c *gin.Context) {
	id, ok := h.ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := h.GetService().StopTask(id); err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, nil, "Task stopped successfully")
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := h.ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := h.GetService().DeleteTask(id); err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, nil, "Task deleted successfully")
}

// ParseIntParamFromQuery 从查询参数解析整数
func (h *TaskHandler) ParseIntParamFromQuery(c *gin.Context, param string) (int64, bool) {
	value, err := h.ParseStringToInt64(param)
	if err != nil {
		h.BadRequest(c, "Invalid parameter")
		return 0, false
	}
	return value, true
}
