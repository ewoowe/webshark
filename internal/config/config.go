package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"webshark/internal/logger"

	"github.com/spf13/viper"
)

var (
	cfg  *viper.Viper
	once sync.Once
)

// Config 配置结构体
type Config struct {
	App      AppConfig     `mapstructure:"app"`
	Server   ServerConfig  `mapstructure:"web"`
	Log      LogConfig     `mapstructure:"log"`
	Nacos    NacosConfig   `mapstructure:"tnacos"`
	Database *DBConfig     `mapstructure:"database"`
	Output   OutputConfig  `mapstructure:"output"`
	Capture  CaptureConfig `mapstructure:"capture"`
}

// NacosConfig Nacos 配置
type NacosConfig struct {
	ServerAddr string `mapstructure:"server_addr"`
	ServerPort int    `mapstructure:"server_port"`
	Namespace  string `mapstructure:"namespace"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// OutputConfig 输出控制配置
type OutputConfig struct {
	EnableAccessTimeline bool `mapstructure:"enable_access_timeline"` // 是否输出接入时间线详情
	EnableTimeslotAccess bool `mapstructure:"enable_timeslot_access"` // 是否输出时刻槽接入情况
	EnableSQLLog         bool `mapstructure:"enable_sql_log"`         // 是否输出 SQL 日志
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	URL  string `mapstructure:"url"` // WebSocket URL 路径
}

// LogConfig 日志配置
type LogConfig struct {
	Level        string           `mapstructure:"level"`
	Format       string           `mapstructure:"format"`
	OutputPaths  []string         `mapstructure:"output_paths"`
	RotateConfig *LogRotateConfig `mapstructure:"rotate"`
}

// LogRotateConfig 日志切割配置
type LogRotateConfig struct {
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个文件最大大小 (MB)
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 日志文件保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩旧日志
}

// CaptureConfig 抓包配置
type CaptureConfig struct {
	TsharkPath                string `mapstructure:"tshark"`               // tshark 命令路径
	PcapDir                   string `mapstructure:"pcap"`                 // PCAP 文件存储目录
	DetailFlushTimeoutSeconds int    `mapstructure:"detail_flush_timeout"` // 包详情空闲刷新超时（秒），0 表示使用默认值 10
}

// InitConfig 初始化配置
func InitConfig(configPath string) (*Config, error) {
	var initErr error
	once.Do(func() {
		cfg = viper.New()

		// 设置配置文件名（不带扩展名）
		if configPath != "" {
			cfg.SetConfigFile(configPath)
		} else {
			// 默认查找路径
			cfg.SetConfigName("config")
			cfg.AddConfigPath(".")
			cfg.AddConfigPath("./config/")
			cfg.AddConfigPath("/etc/webshark/")
		}

		// 设置配置文件类型
		cfg.SetConfigType("yaml")

		// 自动环境变量
		cfg.AutomaticEnv()
		cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// 设置默认值
		setDefaults()

		// 读取配置
		if err := cfg.ReadInConfig(); err != nil {
			if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
				initErr = fmt.Errorf("读取配置文件失败：%w", err)
				logger.Error(initErr.Error())
				return
			}
			// 配置文件不存在时使用默认值
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	// 将配置解码到结构体
	var config Config
	if err := cfg.Unmarshal(&config); err != nil {
		unmarshalErr := fmt.Errorf("解码配置文件失败：%w", err)
		logger.Error(unmarshalErr.Error())
		return nil, unmarshalErr
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用默认配置
	cfg.SetDefault("app.name", "WebShark")
	cfg.SetDefault("app.version", "v1")

	// 服务器默认配置
	cfg.SetDefault("web.host", "0.0.0.0")
	cfg.SetDefault("web.port", 38081)

	// 日志默认配置
	cfg.SetDefault("log.level", "info")
	cfg.SetDefault("log.format", "json")
	cfg.SetDefault("log.output_paths", []string{"stdout"})

	// 日志切割默认配置
	cfg.SetDefault("log.rotate.filename", "logs/webshark.log")
	cfg.SetDefault("log.rotate.max_size", 10)
	cfg.SetDefault("log.rotate.max_backups", 5)
	cfg.SetDefault("log.rotate.max_age", 30)
	cfg.SetDefault("log.rotate.compress", false)

	// 输出控制默认配置
	cfg.SetDefault("output.enable_sql_log", false) // SQL 日志默认关闭，避免性能影响

	// 抓包默认配置
	cfg.SetDefault("capture.pcap", "./pcaps")          // PCAP 文件存储目录
	cfg.SetDefault("capture.detail_flush_timeout", 10) // 包详情空闲刷新超时（秒）
}

// GetViper 获取 viper 实例（用于高级操作）
func GetViper() *viper.Viper {
	if cfg == nil {
		_, err := InitConfig("")
		if err != nil {
			return nil
		}
	}
	return cfg
}

// GetConfig 获取配置（便捷方法）
func GetConfig() (*Config, error) {
	if cfg == nil {
		return InitConfig("")
	}

	var config Config
	if err := cfg.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败：%w", err)
	}

	return &config, nil
}

// GetCapturePcapDir 获取 PCAP 文件存储目录
func GetCapturePcapDir() string {
	if cfg == nil {
		_, err := InitConfig("")
		if err != nil {
			return "."
		}
	}
	return cfg.GetString("capture.pcap")
}

// GetCaptureTsharkPath 获取 tshark 命令路径
func GetCaptureTsharkPath() string {
	if cfg == nil {
		_, err := InitConfig("")
		if err != nil {
			return "."
		}
	}
	return cfg.GetString("capture.tshark")
}
