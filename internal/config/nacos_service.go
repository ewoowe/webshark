package config

import (
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// NacosConfigService Nacos配置服务接口
type NacosConfigService interface {
	// GetConfig 获取配置
	GetConfig(dataId, group string) (string, error)
	// PublishConfig 发布配置
	PublishConfig(dataId, group, content string) error
	// DeleteConfig 删除配置
	DeleteConfig(dataId, group string) error
	// ListenConfig 监听配置变化
	ListenConfig(dataId, group string, listener func(namespace, group, dataId, data string)) error
	// CancelListenConfig 取消监听
	CancelListenConfig(dataId, group string) error
}

// nacosConfigService Nacos配置服务实现
type nacosConfigService struct{}

// NewNacosConfigService 创建 Nacos配置服务实例
func NewNacosConfigService() NacosConfigService {
	return &nacosConfigService{}
}

// GetConfig 获取配置
func (s *nacosConfigService) GetConfig(dataId, group string) (string, error) {
	content, err := GetNacosConfigClient().GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return "", err
	}
	return content, nil
}

// PublishConfig 发布配置
func (s *nacosConfigService) PublishConfig(dataId, group, content string) error {
	success, err := GetNacosConfigClient().PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})
	if err != nil {
		return err
	}
	if !success {
		return nil
	}
	return nil
}

// DeleteConfig 删除配置
func (s *nacosConfigService) DeleteConfig(dataId, group string) error {
	success, err := GetNacosConfigClient().DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return err
	}
	if !success {
		return nil
	}
	return nil
}

// ListenConfig 监听配置变化
func (s *nacosConfigService) ListenConfig(dataId, group string, listener func(namespace, group, dataId, data string)) error {
	err := GetNacosConfigClient().ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			listener(namespace, group, dataId, data)
		},
	})
	return err
}

// CancelListenConfig 取消监听
func (s *nacosConfigService) CancelListenConfig(dataId, group string) error {
	err := GetNacosConfigClient().CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	return err
}
