package gorm

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository 通用仓储基类，提供泛型 CRUD 操作
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository 创建通用仓储实例
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// Create 创建记录
func (r *BaseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// BatchCreate 批量创建记录
func (r *BaseRepository[T]) BatchCreate(entities []*T, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100 // 默认每批 100 条
	}
	return r.db.CreateInBatches(entities, batchSize).Error
}

// GetByID 根据 ID 获取记录
func (r *BaseRepository[T]) GetByID(id int64) (*T, error) {
	var entity T
	err := r.db.First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record not found: %d", id)
		}
		return nil, err
	}
	return &entity, nil
}

// Update 更新记录
func (r *BaseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

// Delete 删除记录
func (r *BaseRepository[T]) Delete(id int64) error {
	var entity T
	return r.db.Delete(&entity, id).Error
}

// Exists 检查记录是否存在
func (r *BaseRepository[T]) Exists(id int64) (bool, error) {
	var count int64
	var entity T
	err := r.db.Model(&entity).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count 获取总数
func (r *BaseRepository[T]) Count() (int64, error) {
	var entity T
	var count int64
	err := r.db.Model(&entity).Count(&count).Error
	return count, err
}

// CountByCondition 根据条件获取总数
func (r *BaseRepository[T]) CountByCondition(query string, args ...interface{}) (int64, error) {
	var entity T
	var count int64
	err := r.db.Model(&entity).Where(query, args...).Count(&count).Error
	return count, err
}

// List 分页列表查询
func (r *BaseRepository[T]) List(page, pageSize int, orderBy string) ([]*T, int64, error) {
	var entities []*T
	var total int64
	var entity T

	// 计算总数
	if err := r.db.Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 处理分页参数
	offset := (page - 1) * pageSize
	if page <= 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 默认排序
	if orderBy == "" {
		orderBy = "created_at DESC"
	}

	// 分页查询
	err := r.db.Offset(offset).Limit(pageSize).Order(orderBy).Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// ListByCondition 根据条件分页查询
func (r *BaseRepository[T]) ListByCondition(query string, args []interface{}, page, pageSize int, orderBy string) ([]*T, int64, error) {
	var entities []*T
	var total int64
	var entity T

	// 构建查询
	dbQuery := r.db.Where(query, args...)

	// 计算总数
	if err := dbQuery.Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 处理分页参数
	offset := (page - 1) * pageSize
	if page <= 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 默认排序
	if orderBy == "" {
		orderBy = "created_at DESC"
	}

	// 分页查询
	err := dbQuery.Offset(offset).Limit(pageSize).Order(orderBy).Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// UpdateField 更新单个字段
func (r *BaseRepository[T]) UpdateField(id int64, field string, value interface{}) error {
	var entity T
	return r.db.Model(&entity).Where("id = ?", id).Update(field, value).Error
}

// DeleteByCondition 根据条件删除
func (r *BaseRepository[T]) DeleteByCondition(query string, args ...interface{}) error {
	var entity T
	return r.db.Where(query, args...).Delete(&entity).Error
}

// FindByCondition 根据条件查询（不分页）
func (r *BaseRepository[T]) FindByCondition(query string, args []interface{}, orderBy string) ([]*T, error) {
	var entities []*T

	dbQuery := r.db.Where(query, args...)

	if orderBy == "" {
		orderBy = "id ASC"
	}

	err := dbQuery.Order(orderBy).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return entities, nil
}

// FirstByCondition 根据条件查询第一条记录
func (r *BaseRepository[T]) FirstByCondition(query string, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.Where(query, args...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record not found")
		}
		return nil, err
	}
	return &entity, nil
}
