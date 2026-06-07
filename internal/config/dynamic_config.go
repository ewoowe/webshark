package config

import (
	"encoding/json"
	"sync"
)

// ConfigChangeListener 配置变更监听器
type ConfigChangeListener func(configType string, content string)

// DynamicConfigManager 动态配置管理器
type DynamicConfigManager struct {
	listeners map[string]map[string]ConfigChangeListener // dataId -> group -> listener
	mu        sync.RWMutex
	service   NacosConfigService
}

// NewDynamicConfigManager 创建动态配置管理器
func NewDynamicConfigManager() *DynamicConfigManager {
	return &DynamicConfigManager{
		listeners: make(map[string]map[string]ConfigChangeListener),
		service:   NewNacosConfigService(),
	}
}

// Start 启动动态配置监听
func (m *DynamicConfigManager) Start() error {
	// 这里可以添加需要监听的配置列表
	// 例如：监听应用配置、业务配置等
	configs := []struct {
		dataId string
		group  string
	}{
		{"app-config.yaml", "DEFAULT_GROUP"},
		{"business-config.yaml", "DEFAULT_GROUP"},
	}

	for _, cfg := range configs {
		err := m.service.ListenConfig(cfg.dataId, cfg.group, m.createListener(cfg.dataId, cfg.group))
		if err != nil {
			return err
		}
	}
	return nil
}

// createListener 创建监听器回调
func (m *DynamicConfigManager) createListener(dataId, group string) func(namespace, group, dataId, data string) {
	return func(namespace, group, dataId, data string) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		// 触发所有注册的监听器
		if listeners, ok := m.listeners[dataId]; ok {
			for _, listener := range listeners {
				go listener(dataId, data)
			}
		}
	}
}

// RegisterListener 注册配置变更监听器
func (m *DynamicConfigManager) RegisterListener(dataId, group string, listener ConfigChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.listeners[dataId]; !ok {
		m.listeners[dataId] = make(map[string]ConfigChangeListener)
	}
	m.listeners[dataId][group] = listener
}

// UnregisterListener 注销配置变更监听器
func (m *DynamicConfigManager) UnregisterListener(dataId, group string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if listeners, ok := m.listeners[dataId]; ok {
		delete(listeners, group)
		if len(listeners) == 0 {
			delete(m.listeners, dataId)
		}
	}
}

// GetConfig 获取配置
func (m *DynamicConfigManager) GetConfig(dataId, group string) (string, error) {
	return m.service.GetConfig(dataId, group)
}

// PublishConfig 发布配置
func (m *DynamicConfigManager) PublishConfig(dataId, group, content string) error {
	return m.service.PublishConfig(dataId, group, content)
}

// GetJsonConfig 获取 JSON 格式配置并解析
func (m *DynamicConfigManager) GetJsonConfig(dataId, group string, target interface{}) error {
	content, err := m.service.GetConfig(dataId, group)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(content), target)
}

// GetYamlConfig 获取 YAML 格式配置并解析
func (m *DynamicConfigManager) GetYamlConfig(dataId, group string, target interface{}) error {
	content, err := m.service.GetConfig(dataId, group)
	if err != nil {
		return err
	}
	// 使用 viper 解析 YAML
	return json.Unmarshal([]byte(content), target)
}
