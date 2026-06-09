package entity

import "fmt"

// ApiCode API 响应状态码枚举
type ApiCode int

const (
	Success ApiCode = iota // 成功
	Failure                // 失败
)

// ApiResponse 基础响应结构
type ApiResponse[T any] struct {
	Code ApiCode `json:"code"`
	Msg  string  `json:"msg"`
	Data T       `json:"data"`
}

// IsSuccess 判断请求是否成功
func (r *ApiResponse[T]) IsSuccess() bool {
	return r.Code == Success
}

// GetData 安全获取数据
func (r *ApiResponse[T]) GetData() (T, error) {
	var zero T
	if !r.IsSuccess() {
		return zero, fmt.Errorf("请求失败: %s", r.Msg)
	}
	return r.Data, nil
}

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int `form:"page" json:"page"`         // 页码，从1开始
	PageSize int `form:"pageSize" json:"pageSize"` // 每页数量
}

// Validate 验证分页参数
func (p *PageRequest) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
}

// PageResponse 分页响应数据
type PageResponse[T any] struct {
	Items     []T   `json:"items"`     // 数据列表
	Total     int64 `json:"total"`     // 总记录数
	Page      int   `json:"page"`      // 当前页码
	PageSize  int   `json:"pageSize"`  // 每页数量
	TotalPage int64 `json:"totalPage"` // 总页数
}

// NewPageResponse 创建分页响应
func NewPageResponse[T any](items []T, total int64, page, pageSize int) *PageResponse[T] {
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)
	return &PageResponse[T]{
		Items:     items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}
}
