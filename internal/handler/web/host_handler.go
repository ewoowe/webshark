package web

import (
	"webshark/internal/service"

	"github.com/gin-gonic/gin"
)

// HostHandler Host 处理器
type HostHandler struct {
	*BaseHandler[*service.HostService]
}

// NewHostHandler 创建 Host 处理器
func NewHostHandler(hostSvc *service.HostService) *HostHandler {
	return &HostHandler{
		BaseHandler: NewBaseHandler[*service.HostService](hostSvc),
	}
}

// CreateHost 创建主机
func (h *HostHandler) CreateHost(c *gin.Context) {
	var req service.CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	host, err := h.GetService().CreateHost(&req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, host, "Host created successfully")
}

// GetHost 获取单个主机
func (h *HostHandler) GetHost(c *gin.Context) {
	id, ok := h.ParseIntParam(c, "id")
	if !ok {
		return
	}

	host, err := h.GetService().GetHostByID(id)
	if err != nil {
		h.NotFound(c, err)
		return
	}

	h.Success(c, host)
}

// ListHosts 获取主机列表
func (h *HostHandler) ListHosts(c *gin.Context) {
	page, pageSize := h.ParsePageParams(c)

	req := &service.ListHostsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.GetService().ListHosts(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.Success(c, resp)
}

// SearchHosts 搜索主机
func (h *HostHandler) SearchHosts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		h.BadRequest(c, "Missing search keyword")
		return
	}

	page, pageSize := h.ParsePageParams(c)

	req := &service.SearchHostsRequest{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.GetService().SearchHosts(req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.Success(c, resp)
}

// UpdateHost 更新主机
func (h *HostHandler) UpdateHost(c *gin.Context) {
	var req service.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	host, err := h.GetService().UpdateHost(&req)
	if err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, host, "Host updated successfully")
}

// DeleteHost 删除主机
func (h *HostHandler) DeleteHost(c *gin.Context) {
	id, ok := h.ParseIntParam(c, "id")
	if !ok {
		return
	}

	if err := h.GetService().DeleteHost(id); err != nil {
		h.InternalError(c, err)
		return
	}

	h.SuccessWithMsg(c, nil, "Host deleted successfully")
}
