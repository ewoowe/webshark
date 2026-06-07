package gorm

import (
	"time"
)

// Host 远程主机信息
type Host struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	HostName  string    `gorm:"column:host_name;type:varchar(64);not null;index:idx_host_name;comment:主机名称" json:"hostName"`
	IP        string    `gorm:"column:ip;type:varchar(64);not null;index:idx_ip;comment:主机IP地址" json:"ip"`
	UserName  string    `gorm:"column:user_name;type:varchar(64);not null;comment:用户名" json:"userName"`
	Password  string    `gorm:"column:password;type:varchar(256);not null;comment:密码" json:"password"`
	OS        string    `gorm:"column:os;type:varchar(64);comment:操作系统，以实际检测为主" json:"os"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
}

// TableName 指定表名
func (Host) TableName() string {
	return "host"
}

// Task 抓包任务信息
type Task struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TaskName        string    `gorm:"column:task_name;type:varchar(64);not null;index:idx_task_name;comment:任务名称或者任务组名称" json:"taskName"`
	HostID          int64     `gorm:"column:host_id;type:bigint;not null;index:idx_host_id;comment:抓包任务使用的主机ID" json:"hostId"`
	Interfaces      []string  `gorm:"column:interfaces;type:varchar(256);comment:抓包接口列表，为空时则抓取any" json:"interfaces"`
	OnlyCapture     bool      `gorm:"column:only_capture;type:boolean;default:false;comment:是否只抓包，不解析，如果是那么就没有实时解析内容的展示" json:"onlyCapture"`
	ParseDetail     bool      `gorm:"column:parse_detail;type:boolean;default:false;comment:是否解析抓包文件详细内容，如果是那么就只有一条条概览信息" json:"parseDetail"`
	DetailFormat    string    `gorm:"column:detail_format;type:varchar(64);default:json;comment:详细内容格式，normal，json，pdml，ek等" json:"detailFormat"`
	FilePath        string    `gorm:"column:file_path;type:varchar(256);not null;comment:抓包文件路径" json:"filePath"`
	BpfFilter       string    `gorm:"column:bpf_filter;type:varchar(256);comment:BPF过滤条件" json:"bpfFilter"`
	WiresharkFilter string    `gorm:"column:wireshark_filter;type:varchar(256);comment:Wireshark过滤条件" json:"wiresharkFilter"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:任务开始时间" json:"createdAt"`
	StopAt          time.Time `gorm:"column:stop_at;type:timestamp;comment:任务停止时间" json:"stopAt"`
	TaskGroupId     int64     `gorm:"column:task_group_id;type:bigint;comment:任务组ID" json:"taskGroupId"`
}

// TableName 获取表名
func (Task) TableName() string {
	return "task"
}

// TaskGroup 任务组信息
type TaskGroup struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TaskGroupName string    `gorm:"column:task_name;type:varchar(64);not null;index:idx_task_name;comment:任务组名称" json:"taskGroupName"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:任务组开始时间" json:"createdAt"`
	StopAt        time.Time `gorm:"column:stop_at;type:timestamp;comment:任务组停止时间" json:"stopAt"`
}

// TableName 获取表名
func (TaskGroup) TableName() string {
	return "task_group"
}

// Packet 数据包表
type Packet struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TaskID      int64  `gorm:"column:task_id;type:bigint;not null;index:idx_task_id;comment:任务ID" json:"taskId"`
	No          int64  `gorm:"column:no;type:bigint;not null;comment:任务组内全局序号" json:"no"`
	FrameNumber int64  `gorm:"column:frame_number;type:bigint;not null;comment:任务内数据包序号" json:"frameNumber"`
	Timestamp   int64  `gorm:"column:timestamp;type:bigint;not null;comment:纳秒级UNIX时间戳" json:"timestamp"`
	Src         string `gorm:"column:src;type:varchar(64);not null;comment:源地址" json:"src"`
	Dst         string `gorm:"column:dst;type:varchar(64);not null;comment:目的地址" json:"dst"`
	Protocol    string `gorm:"column:protocol;type:varchar(64);not null;comment:协议" json:"protocol"`
	Length      int64  `gorm:"column:length;type:bigint;not null;comment:数据包长度" json:"length"`
	Info        string `gorm:"column:info;type:varchar(256);comment:信息" json:"info"`
	Content     string `gorm:"column:content;type:blob;comment:解析了的数据包详情内容（zstd压缩算法输出的二进制数据）" json:"content"`
}
