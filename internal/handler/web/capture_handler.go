package web

import (
	"net/http"
	"webshark/internal/entity"
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
		logger.Error("开启抓包失败", zap.Any("CaptureRequest", req), zap.Error(err))
		InternalErrorWithMsg(c, err, "开启抓包失败")
		return
	}

	SuccessWithMsg(c, taskInfo, "开启抓包成功")
}

// StopCapture 停止抓包（Gin handler）
func StopCapture(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, entity.ApiResponse[any]{
			Code: entity.Failure,
			Msg:  "Missing session_id",
		})
		return
	}

	err := service.StopCapture("", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ApiResponse[any]{
			Code: entity.Failure,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.ApiResponse[any]{
		Code: entity.Success,
		Msg:  "Capture stopped",
	})
}
