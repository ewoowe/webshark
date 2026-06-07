package config

import (
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/http_agent"
)

var (
	nacosClient  *nacos_client.NacosClient
	configClient config_client.IConfigClient
	ncOnce       sync.Once
)

// InitNacosConfig 初始化 Nacos配置客户端
func InitNacosConfig() (err error) {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	ncOnce.Do(func() {
		// 从 viper 中获取 Nacos配置
		serverConfigs := []constant.ServerConfig{
			{
				IpAddr: cfg.Nacos.ServerAddr,
				Port:   uint64(cfg.Nacos.ServerPort),
			},
		}
		clientConfig := constant.ClientConfig{
			NamespaceId:         cfg.Nacos.Namespace,
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogDir:              "logs/nacos",
			CacheDir:            "cache/nacos",
			LogLevel:            "info",
		}

		client := &nacos_client.NacosClient{}
		err = client.SetClientConfig(clientConfig)
		if err != nil {
			return
		}
		err = client.SetServerConfig(serverConfigs)
		if err != nil {
			return
		}
		// 设置 HTTP Agent
		httpAgent := &http_agent.HttpAgent{}
		err = client.SetHttpAgent(httpAgent)
		if err != nil {
			return
		}
		nacosClient = client

		// 创建配置客户端
		configClient, err = config_client.NewConfigClient(client)
		if err != nil {
			return
		}
	})
	return err
}

// GetNacosConfigClient 获取 Nacos配置客户端实例
func GetNacosConfigClient() config_client.IConfigClient {
	if configClient == nil {
		panic("Nacos配置客户端未初始化，请先调用 InitNacosConfig")
	}
	return configClient
}
