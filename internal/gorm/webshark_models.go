package gorm

import (
	"time"
)

// Host 远程主机信息
type Host struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	HostName  string    `gorm:"column:host_name;type:varchar(64);not null;index:idx_host_name" json:"hostName"`
	IP        string    `gorm:"column:ip;type:varchar(64);not null;index:idx_ip" json:"ip"`
	UserName  string    `gorm:"column:user_name;type:varchar(64);not null" json:"userName"`
	Password  string    `gorm:"column:password;type:varchar(256);not null" json:"password"`
	OS        string    `gorm:"column:os;type:varchar(64)" json:"os"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName 指定表名
func (Host) TableName() string {
	return "host"
}
