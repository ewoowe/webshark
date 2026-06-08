package gorm

import (
	"time"
)

// GetAllModels 返回所有需要迁移的模型列表
// 新增模型时，只需在此处添加即可
func GetAllModels() []interface{} {
	return []interface{}{
		&Host{},
		&Task{},
		&TaskGroup{},
		&Packet{},
		&Process{},
	}
}

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
	StreamID        int8      `gorm:"column:stream_id;type:tinyint;default:0;comment:抓包流ID，同一个任务组内唯一" json:"streamId"`
	HostID          int64     `gorm:"column:host_id;type:bigint;not null;index:idx_host_id;comment:抓包任务使用的主机ID" json:"hostId"`
	Interfaces      []string  `gorm:"column:interfaces;type:varchar(256);comment:抓包接口列表，为空时则抓取any" json:"interfaces"`
	OnlyCapture     bool      `gorm:"column:only_capture;type:boolean;default:false;comment:是否只抓包，不解析，如果是那么就没有实时解析内容的展示" json:"onlyCapture"`
	ParseDetail     bool      `gorm:"column:parse_detail;type:boolean;default:false;comment:是否解析抓包文件详细内容，如果是那么就只有一条条概览信息" json:"parseDetail"`
	DetailFormat    string    `gorm:"column:detail_format;type:varchar(64);default:json;comment:详细内容格式，normal，json，pdml，ek等" json:"detailFormat"`
	FilePath        string    `gorm:"column:file_path;type:varchar(256);not null;comment:抓包文件保存路径" json:"filePath"`
	FifoPath        string    `gorm:"column:detail_fifo;type:varchar(256);comment:抓包流量FIFO文件路径，用于详细解析" json:"detailFifo"`
	BpfFilter       string    `gorm:"column:bpf_filter;type:varchar(256);comment:BPF过滤条件" json:"bpfFilter"`
	WiresharkFilter string    `gorm:"column:wireshark_filter;type:varchar(256);comment:Wireshark过滤条件" json:"wiresharkFilter"`
	FullCommand     string    `gorm:"column:full_command;type:varchar(512);comment:完整的命令行" json:"fullCommand"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:任务开始时间" json:"createdAt"`
	StopAt          time.Time `gorm:"column:stop_at;type:timestamp;comment:任务停止时间" json:"stopAt"`
	TaskGroupId     int64     `gorm:"column:task_group_id;type:bigint;comment:任务组ID" json:"taskGroupId"`
	Status          string    `gorm:"column:status;type:varchar(64);default:running;comment:任务状态，created，running，stopped，failed" json:"status"`
	Message         string    `gorm:"column:message;type:varchar(256);comment:任务状态下的额外消息" json:"message"`
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
	Content     string `gorm:"column:content;type:mediumtext;comment:数据包详情(zstd压缩+base64编码)" json:"-"`
}

// TableName 获取表名
func (Packet) TableName() string {
	return "packet"
}

// Process 进程信息，记录单个Task下运行中的进程信息
type Process struct {
	ID      int64  `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TaskID  int64  `gorm:"column:task_id;type:bigint;not null;index:idx_task_id;comment:任务ID" json:"taskId"`
	Pid     int64  `gorm:"column:pid;type:bigint;not null;comment:进程ID" json:"pid"`
	Ppid    int64  `gorm:"column:ppid;type:bigint;not null;comment:父进程ID" json:"ppid"`
	Type    string `gorm:"column:type;type:varchar(64);not null;comment:进程类型, sshpass, tshark, tee" json:"type"`
	Command string `gorm:"column:command;type:varchar(64);not null;comment:进程命令行" json:"command"`
	Alive   bool   `gorm:"column:alive;type:boolean;not null;default:true;comment:进程是否存活" json:"alive"`
}

// TableName 获取表名
func (Process) TableName() string {
	return "process"
}
