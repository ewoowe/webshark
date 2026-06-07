package service

import (
	"fmt"
	"webshark/internal/entity"
	"webshark/internal/gorm"
)

// HostService Host 服务
type HostService struct {
	repo *gorm.WebSharkRepository
}

// NewHostService 创建 Host 服务实例
func NewHostService(repo *gorm.WebSharkRepository) *HostService {
	return &HostService{
		repo: repo,
	}
}

// CreateHostRequest 创建主机请求
type CreateHostRequest struct {
	HostName string `json:"hostName" binding:"required"`
	IP       string `json:"ip" binding:"required"`
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
	OS       string `json:"os"`
}

// UpdateHostRequest 更新主机请求
type UpdateHostRequest struct {
	ID       int64  `json:"id" binding:"required"`
	IP       string `json:"ip"`
	HostName string `json:"hostName"`
	UserName string `json:"userName"`
	Password string `json:"password"`
	OS       string `json:"os"`
}

// CreateHost 创建主机
func (s *HostService) CreateHost(req *CreateHostRequest) (*gorm.Host, error) {
	host := &gorm.Host{
		HostName: req.HostName,
		IP:       req.IP,
		UserName: req.UserName,
		Password: req.Password, // 注意：实际应用中应该加密存储
		OS:       req.OS,
	}

	if err := s.repo.CreateHost(host); err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return host, nil
}

// GetHostByID 根据 ID 获取主机
func (s *HostService) GetHostByID(id int64) (*gorm.Host, error) {
	host, err := s.repo.GetHostByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	return host, nil
}

// ListHostsRequest 获取主机列表请求
type ListHostsRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

// ListHosts 获取主机列表
func (s *HostService) ListHosts(req *ListHostsRequest) (*entity.PageResponse[*gorm.Host], error) {
	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	hosts, total, err := s.repo.ListHosts(page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}

	return entity.NewPageResponse(hosts, total, page, pageSize), nil
}

// SearchHostsRequest 搜索主机请求
type SearchHostsRequest struct {
	Keyword  string `form:"keyword" json:"keyword"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
}

// SearchHosts 搜索主机
func (s *HostService) SearchHosts(req *SearchHostsRequest) (*entity.PageResponse[*gorm.Host], error) {
	if req.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	hosts, total, err := s.repo.SearchHosts(req.Keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search hosts: %w", err)
	}

	return entity.NewPageResponse(hosts, total, page, pageSize), nil
}

// UpdateHost 更新主机
func (s *HostService) UpdateHost(req *UpdateHostRequest) (*gorm.Host, error) {
	// 先获取现有记录
	host, err := s.repo.GetHostByID(req.ID)
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

	if err := s.repo.UpdateHost(host); err != nil {
		return nil, fmt.Errorf("failed to update host: %w", err)
	}

	return host, nil
}

// DeleteHost 删除主机
func (s *HostService) DeleteHost(id int64) error {
	if err := s.repo.DeleteHost(id); err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}

	return nil
}
