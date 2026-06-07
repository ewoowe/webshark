package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"webshark/internal/logger"

	"go.uber.org/zap"
)

// EventHandler 事件处理器接口
type EventHandler interface {
	// Handle 处理事件
	Handle(ctx context.Context, event *WsEventMessage) error
	// GetEventType 返回处理器支持的事件类型
	GetEventType() []EventType
}

// EventProcessor 事件处理器（函数类型实现）
type EventProcessor func(ctx context.Context, event *WsEventMessage) error

// Handle 实现 EventHandler 接口
func (p EventProcessor) Handle(ctx context.Context, event *WsEventMessage) error {
	return p(ctx, event)
}

// GetEventType 返回处理器支持的事件类型
func (p EventProcessor) GetEventType() []EventType {
	// 函数类型无法直接获取事件类型，需要通过注册时指定
	return []EventType{}
}

// processorInfo 处理器信息
type processorInfo struct {
	handler   EventHandler
	eventType EventType
	priority  int // 优先级，数值越大优先级越高
}

// EventDispatcher 事件分发器
type EventDispatcher struct {
	mu         sync.RWMutex
	processors map[string][]*processorInfo // 支持一个事件类型多个处理器
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	eventChan  chan *WsEventMessage
}

// DispatcherOption 分发器配置选项
type DispatcherOption func(*EventDispatcher)

// WithEventBufferSize 设置事件缓冲区大小
func WithEventBufferSize(size int) DispatcherOption {
	return func(d *EventDispatcher) {
		d.eventChan = make(chan *WsEventMessage, size)
	}
}

// NewEventDispatcher 创建新的事件分发器
func NewEventDispatcher(opts ...DispatcherOption) *EventDispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	dispatcher := &EventDispatcher{
		processors: make(map[string][]*processorInfo),
		ctx:        ctx,
		cancel:     cancel,
		eventChan:  make(chan *WsEventMessage, 256), // 默认缓冲 256
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(dispatcher)
	}

	return dispatcher
}

// Register 注册事件处理器
func (d *EventDispatcher) Register(eventType EventType, handler EventHandler, priority ...int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	p := 0
	if len(priority) > 0 {
		p = priority[0]
	}

	info := &processorInfo{
		handler:   handler,
		eventType: eventType,
		priority:  p,
	}

	// 添加到处理器列表
	d.processors[eventType.eventType] = append(d.processors[eventType.eventType], info)

	// 按优先级排序（优先级高的在前）
	procs := d.processors[eventType.eventType]
	for i := len(procs) - 1; i > 0; i-- {
		if procs[i].priority > procs[i-1].priority {
			procs[i], procs[i-1] = procs[i-1], procs[i]
		} else {
			break
		}
	}

	logger.Debug("注册事件处理器",
		zap.String("eventType", eventType.eventType),
		zap.Int("priority", p))
}

// RegisterFunc 注册事件处理函数（便捷方法）
func (d *EventDispatcher) RegisterFunc(eventType EventType, handler EventProcessor, priority ...int) {
	wrapper := &eventHandlerWrapper{
		eventType: eventType,
		handler:   handler,
	}
	d.Register(eventType, wrapper, priority...)
}

// Unregister 注销事件处理器
func (d *EventDispatcher) Unregister(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	procs, exists := d.processors[eventType.eventType]
	if !exists {
		return
	}

	// 移除指定的处理器
	newProcs := make([]*processorInfo, 0, len(procs))
	for _, p := range procs {
		if p.handler != handler {
			newProcs = append(newProcs, p)
		}
	}

	d.processors[eventType.eventType] = newProcs

	logger.Debug("注销事件处理器",
		zap.String("eventType", eventType.eventType))
}

// Dispatch 分发事件（非阻塞）
func (d *EventDispatcher) Dispatch(event *WsEventMessage) error {
	select {
	case <-d.ctx.Done():
		return fmt.Errorf("分发器已关闭")
	case d.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("事件队列已满")
	}
}

// DispatchSync 同步分发事件（阻塞）
func (d *EventDispatcher) DispatchSync(event *WsEventMessage) error {
	if d.ctx.Err() != nil {
		return fmt.Errorf("分发器已关闭")
	}

	d.eventChan <- event
	return nil
}

// Start 启动事件分发循环
func (d *EventDispatcher) Start() {
	logger.Info("事件分发器启动中...")
	d.wg.Add(1)
	go d.run()
}

// Stop 停止事件分发器
func (d *EventDispatcher) Stop() {
	logger.Info("事件分发器停止中...")
	d.cancel()
	d.wg.Wait()
	logger.Info("事件分发器已停止")
}

// run 事件分发主循环
func (d *EventDispatcher) run() {
	defer d.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("事件分发器 panic", zap.Any("panic", r))
		}
	}()
	logger.Info("事件分发器协程已启动")
	for {
		select {
		case <-d.ctx.Done():
			// 清空剩余事件
			for {
				select {
				case event := <-d.eventChan:
					logger.Debug("丢弃未处理事件", zap.String("eventType", event.EventType))
				default:
					return
				}
			}
		case event := <-d.eventChan:
			d.processEvent(event)
		}
	}
}

// processEvent 处理单个事件
func (d *EventDispatcher) processEvent(event *WsEventMessage) {
	// 打印事件原始数据
	eventType := GetEventTypeByName(event.EventType)
	if eventType.IsValid() {
		if eventType.IsLogEnabled() {
			if eventType.logFunc != nil {
				eventType.logFunc(event)
			} else {
				logger.Debug("收到事件",
					zap.String("eventType", event.EventType),
					zap.Int64("timestamp", event.Timestamp),
					zap.String("datetime", event.Datetime),
					zap.Reflect("params", json.RawMessage(event.Params)))
			}
		}
	} else {
		return
	}

	d.mu.RLock()
	procs, exists := d.processors[eventType.eventType]
	d.mu.RUnlock()

	if !exists || len(procs) == 0 {
		//logger.Debug("未找到事件处理器", zap.String("eventType", event.EventType))
		return
	}

	// 调用所有注册的处理器（按优先级顺序）
	for _, proc := range procs {
		select {
		case <-d.ctx.Done():
			return
		default:
			err := proc.handler.Handle(d.ctx, event)
			if err != nil {
				logger.Error("事件处理失败",
					zap.String("eventType", event.EventType),
					zap.Error(err))
			}
		}
	}
}

// GetProcessorsCount 获取注册的处理器数量
func (d *EventDispatcher) GetProcessorsCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	count := 0
	for _, procs := range d.processors {
		count += len(procs)
	}
	return count
}

// eventHandlerWrapper 函数类型的处理器包装器
type eventHandlerWrapper struct {
	eventType EventType
	handler   EventProcessor
}

// Handle 实现 EventHandler 接口
func (w *eventHandlerWrapper) Handle(ctx context.Context, event *WsEventMessage) error {
	return w.handler(ctx, event)
}

// GetEventType 返回事件类型
func (w *eventHandlerWrapper) GetEventType() []EventType {
	return []EventType{w.eventType}
}
