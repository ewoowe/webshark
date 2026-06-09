package web

import (
	"webshark/internal/entity"
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateHost 创建主机
func CreateHost(c *gin.Context) {
	var req service.CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, &req)
		return
	}

	host, err := service.CreateHost(&req)
	if err != nil {
		InternalError(c, err)
		return
	}

	SuccessWithMsg(c, host, "主机创建成功")
}

// GetHost 获取单个主机
func GetHost(c *gin.Context) {
	id, ok := ParseIntParam(c, "id")
	if !ok {
		return
	}

	host, err := service.GetHostByID(id)
	if err != nil {
		NotFound(c, err)
		return
	}

	Success(c, host)
}

// ListHosts 获取主机列表
func ListHosts(c *gin.Context) {
	page, pageSize := ParsePageParams(c)

	req := &service.ListHostsRequest{
		PageRequest: entity.PageRequest{
			Page:     page,
			PageSize: pageSize,
		},
	}

	resp, err := service.ListHosts(req)
	if err != nil {
		InternalError(c, err)
		return
	}

	Success(c, resp)
}

// SearchHosts 搜索主机
func SearchHosts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		BadRequest(c, "缺少搜索关键字")
		return
	}

	page, pageSize := ParsePageParams(c)

	req := &service.SearchHostsRequest{
		Keyword: keyword,
		PageRequest: entity.PageRequest{
			Page:     page,
			PageSize: pageSize,
		},
	}

	resp, err := service.SearchHosts(req)
	if err != nil {
		InternalError(c, err)
		return
	}

	Success(c, resp)
}

// UpdateHost 更新主机
func UpdateHost(c *gin.Context) {
	var req service.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, &req)
		return
	}

	host, err := service.UpdateHost(&req)
	if err != nil {
		InternalError(c, err)
		return
	}

	SuccessWithMsg(c, host, "主机更新成功")
}

// DeleteHost 删除主机
func DeleteHost(c *gin.Context) {
	id, ok := ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := service.DeleteHost(id); err != nil {
		InternalError(c, err)
		return
	}

	SuccessWithMsg(c, nil, "主机删除成功")
}
