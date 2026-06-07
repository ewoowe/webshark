package gorm

import (
	"errors"
	"fmt"
	"time"
	"webshark/internal/config"
	"webshark/internal/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// WebSharkRepository 接入数据仓库
type WebSharkRepository struct {
	db     *gorm.DB
	config *config.DBConfig
}

// NewWebSharkRepository 创建WebShark数据仓库
func NewWebSharkRepository(dbConfig *config.DBConfig) (*WebSharkRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	// 从全局配置读取 SQL 日志开关
	enableSQLLog := false
	if globalCfg, err := config.GetConfig(); err == nil && globalCfg != nil {
		enableSQLLog = globalCfg.Output.EnableSQLLog
	}

	// 配置 GORM 日志：集成到项目的 zap 日志框架
	gormConfig := gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  gormlogger.Warn, // 只输出警告和错误（包括慢 SQL）
	}

	// 如果开启 SQL 日志，设置为 Info 级别
	if enableSQLLog {
		gormConfig.LogLevel = gormlogger.Info
	}

	gormLog := logger.NewGormLogger(logger.GetLogger(), gormConfig)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		logger.Error("failed to connect database", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get sql.DB", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

	repo := &WebSharkRepository{
		db:     db,
		config: dbConfig,
	}

	// 自动迁移表结构
	if err := repo.autoMigrate(); err != nil {
		logger.Error("failed to migrate database", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

// autoMigrate 自动迁移表结构
func (r *WebSharkRepository) autoMigrate() error {
	return r.db.AutoMigrate(
		&Host{},
	)
}

// Close 关闭数据库连接
func (r *WebSharkRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ==================== Host 增删改查 ====================

// CreateHost 创建主机记录
func (r *WebSharkRepository) CreateHost(host *Host) error {
	return r.db.Create(host).Error
}

// GetHostByID 根据 ID 获取主机
func (r *WebSharkRepository) GetHostByID(id int64) (*Host, error) {
	var host Host
	err := r.db.First(&host, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host not found: %d", id)
		}
		return nil, err
	}
	return &host, nil
}

// GetHostByIP 根据 IP 获取主机
func (r *WebSharkRepository) GetHostByIP(ip string) (*Host, error) {
	var host Host
	err := r.db.Where("ip = ?", ip).First(&host).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host not found with IP: %s", ip)
		}
		return nil, err
	}
	return &host, nil
}

// ListHosts 获取主机列表（支持分页）
func (r *WebSharkRepository) ListHosts(page, pageSize int) ([]*Host, int64, error) {
	var hosts []*Host
	var total int64

	// 计算总数
	if err := r.db.Model(&Host{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if page <= 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10 // 默认每页 10 条
	}

	err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&hosts).Error
	if err != nil {
		return nil, 0, err
	}

	return hosts, total, nil
}

// SearchHosts 搜索主机（根据 hostname 或 IP）
func (r *WebSharkRepository) SearchHosts(keyword string, page, pageSize int) ([]*Host, int64, error) {
	var hosts []*Host
	var total int64

	// 构建查询条件
	query := r.db.Where("host_name LIKE ? OR ip LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 计算总数
	if err := query.Model(&Host{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if page <= 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&hosts).Error
	if err != nil {
		return nil, 0, err
	}

	return hosts, total, nil
}

// UpdateHost 更新主机信息
func (r *WebSharkRepository) UpdateHost(host *Host) error {
	return r.db.Save(host).Error
}

// UpdateHostPassword 更新主机密码
func (r *WebSharkRepository) UpdateHostPassword(id int64, newPassword string) error {
	return r.db.Model(&Host{}).Where("id = ?", id).Update("password", newPassword).Error
}

// DeleteHost 删除主机（物理删除）
func (r *WebSharkRepository) DeleteHost(id int64) error {
	return r.db.Delete(&Host{}, id).Error
}

// DeleteHostByIP 根据 IP 删除主机
func (r *WebSharkRepository) DeleteHostByIP(ip string) error {
	return r.db.Where("ip = ?", ip).Delete(&Host{}).Error
}

// HostExists 检查主机是否存在
func (r *WebSharkRepository) HostExists(id int64) (bool, error) {
	var count int64
	err := r.db.Model(&Host{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetHostCount 获取主机总数
func (r *WebSharkRepository) GetHostCount() (int64, error) {
	var count int64
	err := r.db.Model(&Host{}).Count(&count).Error
	return count, err
}
