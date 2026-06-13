package web

import (
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// 这里面放一些单独的不成组的API，或者一些需要wrapper的API

// GetInterfaces 获取远程网卡列表（Gin handler）
func GetInterfaces(c *gin.Context) {
	hostIdStr := c.Query("hostId")
	hostId, err := ParseStringToInt64(hostIdStr)
	if err != nil {
		BadRequest(c, "无效的主机")
		return
	}

	host, err := service.GetHostByID(hostId)
	if err != nil {
		NotFound(c)
		return
	}

	interfaces, err := service.GetRemoteInterfaces(host.IP, host.UserName, host.Password)
	if err != nil {
		InternalErrorWithMsg(c, nil, "获取远程网卡列表失败")
		return
	}

	SuccessWithMsg(c, interfaces, "获取远程网卡列表成功")
}
