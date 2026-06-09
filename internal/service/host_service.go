package service

import (
	"fmt"
	"webshark/internal/entity"
	"webshark/internal/gorm"
)

// CreateHostRequest 创建主机请求
type CreateHostRequest struct {
	HostName string `json:"hostName" binding:"required" label:"主机名称"`
	IP       string `json:"ip" binding:"required" label:"IP地址"`
	UserName string `json:"userName" binding:"required" label:"用户名"`
	Password string `json:"password" binding:"required" label:"密码"`
	OS       string `json:"os" label:"操作系统"`
}

// UpdateHostRequest 更新主机请求
type UpdateHostRequest struct {
	ID int64 `json:"id" binding:"required" label:"主机ID"`
	CreateHostRequest
}

// ListHostsRequest 获取主机列表请求
type ListHostsRequest struct {
	entity.PageRequest
}

// SearchHostsRequest 搜索主机请求
type SearchHostsRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	entity.PageRequest
}

// CreateHost 创建主机
func CreateHost(req *CreateHostRequest) (*gorm.Host, error) {
	host := &gorm.Host{
		HostName: req.HostName,
		IP:       req.IP,
		UserName: req.UserName,
		Password: req.Password, // 注意：实际应用中应该加密存储
		OS:       req.OS,
	}

	if err := gorm.Repo.CreateHost(host); err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return host, nil
}

// GetHostByID 根据 ID 获取主机
func GetHostByID(id int64) (*gorm.Host, error) {
	host, err := gorm.Repo.GetHostByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	return host, nil
}

// ListHosts 获取主机列表
func ListHosts(req *ListHostsRequest) (*entity.PageResponse[*gorm.Host], error) {
	page := req.Page
	pageSize := req.PageSize

	hosts, total, err := gorm.Repo.ListHosts(page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}

	return entity.NewPageResponse(hosts, total, page, pageSize), nil
}

// SearchHosts 搜索主机
func SearchHosts(req *SearchHostsRequest) (*entity.PageResponse[*gorm.Host], error) {
	if req.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	page := req.Page
	pageSize := req.PageSize

	hosts, total, err := gorm.Repo.SearchHosts(req.Keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search hosts: %w", err)
	}

	return entity.NewPageResponse(hosts, total, page, pageSize), nil
}

// UpdateHost 更新主机
func UpdateHost(req *UpdateHostRequest) (*gorm.Host, error) {
	// 先获取现有记录
	host, err := gorm.Repo.GetHostByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	// 更新字段
	if req.HostName != "" {
		host.HostName = req.HostName
	}
	if req.IP != "" {
		host.IP = req.IP
	}
	if req.UserName != "" {
		host.UserName = req.UserName
	}
	if req.Password != "" {
		host.Password = req.Password // 注意：实际应用中应该加密存储
	}
	if req.OS != "" {
		host.OS = req.OS
	}

	if err := gorm.Repo.UpdateHost(host); err != nil {
		return nil, fmt.Errorf("failed to update host: %w", err)
	}

	return host, nil
}

// DeleteHost 删除主机
func DeleteHost(id int64) error {
	if err := gorm.Repo.DeleteHost(id); err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}

	return nil
}
