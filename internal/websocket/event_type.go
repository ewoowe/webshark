package websocket

import (
	"sync"
)

// EventType 预定义的事件类型
type EventType struct {
	eventType  string                      // 事件类型名称
	logEnabled bool                        // 是否记录日志
	logFunc    func(event *WsEventMessage) // 如果需要记录日志并且实现了此方法，事件分发器将使用该方法打印日志事件，
	// 如果未实现该方法，事件分发器将使用默认的日志方法
}

// NewEventType 创建新的 EventType，logEnabled 默认为 true
func NewEventType(eventType string) EventType {
	return EventType{
		eventType:  eventType,
		logEnabled: true, // 默认启用日志
	}
}

// NewEventTypeWithLog 创建新的 EventType，并指定是否启用日志
func NewEventTypeWithLog(eventType string, logEnabled bool) EventType {
	return EventType{
		eventType:  eventType,
		logEnabled: logEnabled,
	}
}

// NewEventTypeWithLogFunc 创建新的 EventType，并指定自定义日志函数
func NewEventTypeWithLogFunc(eventType string, logEnabled bool, logFunc func(event *WsEventMessage)) EventType {
	return EventType{
		eventType:  eventType,
		logEnabled: logEnabled,
		logFunc:    logFunc,
	}
}

// 预定义的事件类型（使用 var 而非 const，因为结构体不能作为常量）
var (
	// EventTypeConnected 连接相关
	EventTypeConnected    = NewEventType("Connected")
	EventTypeDisconnected = NewEventType("Disconnected")

	// EventTypePing 心跳
	EventTypePing = NewEventTypeWithLog("Ping", false)
	EventTypePong = NewEventTypeWithLog("Pong", false)

	// EventTypeSubscribe 订阅
	EventTypeSubscribe   = NewEventType("Subscribe")
	EventTypeUnsubscribe = NewEventType("Unsubscribe")

	// EventTypeUnknownMessage 错误消息
	EventTypeUnknownMessage = NewEventType("UnknownMessage") // 未知消息类型
	EventTypeInvalidMessage = NewEventType("InvalidMessage") // 无效消息类型

)

// String 返回事件类型的字符串表示
func (e EventType) String() string {
	return e.eventType
}

// IsLogEnabled 返回是否启用日志
func (e EventType) IsLogEnabled() bool {
	return e.logEnabled
}

// Equals 比较两个 EventType 是否相等（基于 eventType 字段）
func (e EventType) Equals(other EventType) bool {
	return e.eventType == other.eventType
}

// Less 实现排序接口（基于 eventType 字段）
func (e EventType) Less(other EventType) bool {
	return e.eventType < other.eventType
}

// Hash 返回事件类型的哈希值（用于 map key）
func (e EventType) Hash() string {
	return e.eventType
}

// 预定义的有效事件类型集合（运行时初始化）
var validEventTypes map[string]EventType
var eventTypeMu sync.RWMutex

// init 初始化有效事件类型集合
func init() {
	validEventTypes = map[string]EventType{
		// 连接相关
		EventTypeConnected.eventType:    EventTypeConnected,
		EventTypeDisconnected.eventType: EventTypeDisconnected,

		// 心跳
		EventTypePing.eventType: EventTypePing,
		EventTypePong.eventType: EventTypePong,

		// 订阅
		EventTypeSubscribe.eventType:   EventTypeSubscribe,
		EventTypeUnsubscribe.eventType: EventTypeUnsubscribe,

		// 错误消息
		EventTypeUnknownMessage.eventType: EventTypeUnknownMessage,
		EventTypeInvalidMessage.eventType: EventTypeInvalidMessage,
	}
}

// IsValid 检查事件类型是否有效
func (e EventType) IsValid() bool {
	eventTypeMu.RLock()
	defer eventTypeMu.RUnlock()
	_, exists := validEventTypes[e.eventType]
	return exists
}

// RegisterEventType 注册新的事件类型（运行时动态添加）
func RegisterEventType(eventType EventType) {
	eventTypeMu.Lock()
	defer eventTypeMu.Unlock()
	validEventTypes[eventType.eventType] = eventType
}

// GetAllEventTypes 获取所有已注册的事件类型
func GetAllEventTypes() []EventType {
	eventTypeMu.RLock()
	defer eventTypeMu.RUnlock()

	types := make([]EventType, 0, len(validEventTypes))
	for _, et := range validEventTypes {
		types = append(types, et)
	}
	return types
}

// GetEventTypeByName 根据名称获取事件类型
func GetEventTypeByName(name string) EventType {
	eventTypeMu.RLock()
	defer eventTypeMu.RUnlock()

	if et, exists := validEventTypes[name]; exists {
		return et
	}
	return EventType{}
}
