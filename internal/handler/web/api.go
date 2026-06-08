package web

import (
	"net/http"
	"webshark/internal/entity"
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// 这里面放一些单独的不成组的API，或者一些需要wrapper的API

// GetInterfaces 获取远程网卡列表（Gin handler）
func GetInterfaces(c *gin.Context) {
	host := c.Query("host")
	username := c.Query("username")
	password := c.Query("password")

	if host == "" || username == "" || password == "" {
		c.JSON(http.StatusBadRequest, entity.ApiResponse[[]service.NetworkInterface]{
			Code: entity.Failure,
			Msg:  "Missing required parameters",
		})
		return
	}

	interfaces, err := service.GetRemoteInterfaces(host, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ApiResponse[[]service.NetworkInterface]{
			Code: entity.Failure,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.ApiResponse[[]service.NetworkInterface]{
		Code: entity.Success,
		Data: &interfaces,
		Msg:  "Success",
	})
}

// 全局变量和初始化函数
var globalHostHandler *HostHandler
var globalTaskHandler *TaskHandler

// InitHostHandler 初始化全局 HostHandler
func InitHostHandler(svc *service.HostService) {
	globalHostHandler = NewHostHandler(svc)
}

// InitTaskHandler 初始化全局 TaskHandler
func InitTaskHandler(svc *service.TaskService) {
	globalTaskHandler = NewTaskHandler(svc)
}

// HostHandler 方法Wrapper

func CreateHost(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.CreateHost(c)
}

func GetHost(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.GetHost(c)
}

func ListHosts(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.ListHosts(c)
}

func SearchHosts(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.SearchHosts(c)
}

func UpdateHost(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.UpdateHost(c)
}

func DeleteHost(c *gin.Context) {
	if !CheckHandler(c, globalHostHandler) {
		return
	}
	globalHostHandler.DeleteHost(c)
}
