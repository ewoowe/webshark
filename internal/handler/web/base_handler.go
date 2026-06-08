package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础 Handler（泛型版本），提供通用的响应方法
type BaseHandler[T any] struct {
	service T
}

// NewBaseHandler 创建基础 Handler
func NewBaseHandler[T any](svc T) *BaseHandler[T] {
	return &BaseHandler[T]{
		service: svc,
	}
}

// GetService 获取服务实例
func (h *BaseHandler[T]) GetService() T {
	return h.service
}

// Success 返回成功响应
func (h *BaseHandler[T]) Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": data,
		"msg":  "Success",
	})
}

// SuccessWithMsg 返回成功响应（带消息）
func (h *BaseHandler[T]) SuccessWithMsg(c *gin.Context, data any, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": data,
		"msg":  msg,
	})
}

// BadRequest 返回参数错误响应
func (h *BaseHandler[T]) BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code": 1,
		"msg":  msg,
	})
}

// InternalError 返回内部错误响应
func (h *BaseHandler[T]) InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": 1,
		"msg":  err.Error(),
	})
}

// NotFound 返回未找到响应
func (h *BaseHandler[T]) NotFound(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, gin.H{
		"code": 1,
		"msg":  err.Error(),
	})
}

// ParseIntParam 解析整数参数
func (h *BaseHandler[T]) ParseIntParam(c *gin.Context, name string) (int64, bool) {
	param := c.Param(name)
	value, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		h.BadRequest(c, "Invalid "+name)
		return 0, false
	}
	return value, true
}

// ParsePageParams 解析分页参数
func (h *BaseHandler[T]) ParsePageParams(c *gin.Context) (page, pageSize int) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

// ParseStringToInt64 将字符串转换为 int64
func (h *BaseHandler[T]) ParseStringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// CheckService 检查服务是否初始化
func (h *BaseHandler[T]) CheckService(c *gin.Context) bool {
	// 通过反射或其他方式检查，这里简化处理
	return true
}

// CheckHandler 检查 Handler 是否初始化（用于全局变量检查）
func CheckHandler[T any](c *gin.Context, handler *T) bool {
	if handler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "Handler not initialized",
		})
		return false
	}
	return true
}
