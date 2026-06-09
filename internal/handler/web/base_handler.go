package web

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"webshark/internal/entity"
	"webshark/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Success 返回成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, entity.ApiResponse[any]{
		Code: entity.Success,
		Data: data,
		Msg:  "操作成功",
	})
}

// SuccessWithMsg 返回成功响应（带消息）
func SuccessWithMsg(c *gin.Context, data any, msg string) {
	c.JSON(http.StatusOK, entity.ApiResponse[any]{
		Code: entity.Success,
		Data: data,
		Msg:  msg,
	})
}

// BadRequest 返回参数错误响应
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, entity.ApiResponse[any]{
		Code: entity.Failure,
		Msg:  msg,
	})
}

// InternalError 返回内部错误响应
func InternalError(c *gin.Context, err error) {
	logger.Error("服务器内部错误", zap.Error(err))
	c.JSON(http.StatusInternalServerError, entity.ApiResponse[any]{
		Code: entity.Failure,
		Msg:  "服务器内部错误",
	})
}

// InternalErrorWithMsg 返回内部错误响应
func InternalErrorWithMsg(c *gin.Context, err error, msg string) {
	logger.Error("服务器内部错误", zap.String("msg", msg), zap.Error(err))
	c.JSON(http.StatusInternalServerError, entity.ApiResponse[any]{
		Code: entity.Failure,
		Msg:  msg,
	})
}

// NotFound 返回未找到响应
func NotFound(c *gin.Context, err error) {
	logger.Error("资源未找到", zap.Error(err))
	c.JSON(http.StatusNotFound, entity.ApiResponse[any]{
		Code: entity.Failure,
		Msg:  "接口请求的资源未找到",
	})
}

// ParseIntParam 解析整数参数
func ParseIntParam(c *gin.Context, name string) (int64, bool) {
	param := c.Param(name)
	value, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		BadRequest(c, "无效参数 "+name)
		return 0, false
	}
	return value, true
}

// ParsePageParams 解析分页参数, 返回页码和每页数量
func ParsePageParams(c *gin.Context) (page, pageSize int) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	var err error
	page, err = strconv.Atoi(pageStr)
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
func ParseStringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ValidationErrorWithStruct 返回详细的验证错误信息（支持 label 标签）
// 需要传入原始的结构体实例以通过反射获取 label 标签
func ValidationErrorWithStruct(c *gin.Context, err error, originalStruct any) {
	// 如果是验证错误，提取详细信息
	if validationErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		var messages []string
		for _, fieldErr := range validationErrs {
			// 通过反射获取字段的 label 标签
			label := getLabelFromStruct(originalStruct, fieldErr)

			var msg string
			switch fieldErr.Tag() {
			case "required":
				msg = fmt.Sprintf("%s 为必填字段", label)
			case "email":
				msg = fmt.Sprintf("%s 格式不正确", label)
			case "url":
				msg = fmt.Sprintf("%s 格式不正确", label)
			case "min":
				msg = fmt.Sprintf("%s 最小值为 %s", label, fieldErr.Param())
			case "max":
				msg = fmt.Sprintf("%s 最大值为 %s", label, fieldErr.Param())
			case "len":
				msg = fmt.Sprintf("%s 长度必须为 %s", label, fieldErr.Param())
			case "oneof":
				msg = fmt.Sprintf("%s 必须是以下值之一: %s", label, fieldErr.Param())
			default:
				msg = fmt.Sprintf("%s 验证失败: %s", label, fieldErr.Tag())
			}
			messages = append(messages, msg)
		}

		BadRequest(c, strings.Join(messages, "; "))
		return
	}

	// 非验证错误，返回通用错误
	BadRequest(c, "参数错误: "+err.Error())
}

// getLabelFromStruct 通过反射从结构体获取字段的 label 标签
func getLabelFromStruct(originalStruct any, fieldErr validator.FieldError) string {
	if originalStruct == nil {
		return fieldErr.Field()
	}

	// 获取结构体的类型
	v := reflect.ValueOf(originalStruct)
	t := v.Type()

	// 如果是指针，获取其底层结构
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// 确保是结构体类型
	if t.Kind() != reflect.Struct {
		return fieldErr.Field()
	}

	// 获取字段名
	fieldName := fieldErr.Field()

	// 查找字段
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	// 尝试获取 label 标签
	labelTag := field.Tag.Get("label")
	if labelTag != "" {
		return labelTag
	}

	// 如果没有 label 标签，返回字段名
	return fieldName
}
