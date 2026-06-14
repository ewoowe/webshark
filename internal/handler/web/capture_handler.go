package web

import (
	"strconv"
	"webshark/internal/logger"
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StartCapture 开始抓包（Gin handler）
func StartCapture(c *gin.Context) {
	var req service.CaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, req)
		return
	}
	if req.DetailFormat == "" {
		req.DetailFormat = "normal"
	}

	taskInfo, err := service.StartCapture(req)
	if err != nil {
		logger.Error("Failed to start capture", zap.Any("CaptureRequest", req), zap.Error(err))
		InternalErrorWithMsg(c, taskInfo, "Failed to start capture")
		return
	}

	SuccessWithMsg(c, taskInfo, "Success start capture")
}

// StopCapture 停止抓包（Gin handler）
func StopCapture(c *gin.Context) {
	taskGroupId := c.Query("taskGroupId")
	taskId := c.Query("taskId")

	// 参数验证：至少提供一个
	if taskGroupId == "" && taskId == "" {
		BadRequest(c, "Missing taskId or taskGroupId parameter")
		return
	}

	err := service.StopCapture(taskGroupId, taskId)
	if err != nil {
		logger.Error("Failed stop capture",
			zap.String("taskGroupId", taskGroupId),
			zap.String("taskId", taskId),
			zap.Error(err))
		InternalErrorWithMsg(c, nil, "Failed stop capture")
		return
	}

	SuccessWithMsg(c, nil, "Success stop capture")
}

// GetPacketDetail 获取单个数据包的详情
func GetPacketDetail(c *gin.Context) {
	// 解析路径参数
	taskIDStr := c.Query("taskId")
	frameNumberStr := c.Query("frameNumber")

	if taskIDStr == "" || frameNumberStr == "" {
		BadRequest(c, "Missing taskId or frameNumber parameter")
		return
	}

	// 转换为 int64
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		BadRequest(c, "Invalid taskId parameter")
		return
	}

	frameNumber, err := strconv.ParseInt(frameNumberStr, 10, 64)
	if err != nil {
		BadRequest(c, "Invalid frameNumber parameter")
		return
	}

	// 调用服务层获取详情
	detail, err := service.GetPacketDetail(taskID, frameNumber)
	if err != nil {
		logger.Error("Failed get packet detail",
			zap.Int64("taskID", taskID),
			zap.Int64("frameNumber", frameNumber),
			zap.Error(err))
		InternalErrorWithMsg(c, nil, "Failed get packet detail")
		return
	}

	SuccessWithMsg(c, detail, "Success get packet detail")
}
