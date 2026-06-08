package utils

// 压缩工具函数使用示例
// 此文件仅用于演示，不会被实际编译

/*
import (
	"encoding/json"
	"webshark/internal/gorm"
)

// 示例1: 压缩字符串（适用于存储 JSON、XML 等文本）
func exampleCompressString() {
	// 假设这是要存储的详细数据包内容
	detailJSON := `{"layers":{"frame":{"frame.number":"1"},"ip":{"ip.src":"192.168.1.1"}}}`

	// 压缩并转换为 base64 字符串
	compressed, err := CompressString(detailJSON)
	if err != nil {
		// 处理错误
		return
	}

	// 将 compressed 存储到数据库的 content 字段
	// packet.Content = compressed

	// 从数据库读取后解压缩
	decompressed, err := DecompressString(compressed)
	if err != nil {
		// 处理错误
		return
	}

	// decompressed 现在是原始 JSON 字符串
	println(decompressed)
}

// 示例2: 压缩结构体对象（推荐用于 Packet Details）
func exampleCompressObject() {
	// 假设这是数据包的详细信息
	type PacketDetail struct {
		Layers    map[string]interface{} `json:"layers"`
		Protocols []string              `json:"protocols"`
		RawData   string                `json:"raw_data"`
	}

	detail := PacketDetail{
		Layers: map[string]interface{}{
			"frame": map[string]string{"frame.number": "1"},
			"ip":    map[string]string{"ip.src": "192.168.1.1"},
		},
		Protocols: []string{"ETH", "IP", "TCP"},
		RawData:   "...",
	}

	// 直接压缩为 base64 字符串
	compressed, err := CompressJSON(detail)
	if err != nil {
		// 处理错误
		return
	}

	// 存储到数据库
	// packet.Content = compressed
}

// 示例3: 从数据库读取并解压缩到结构体
func exampleDecompressToObject() {
	// 从数据库读取的压缩内容
	compressedContent := "..." // 从 packet.Content 获取

	// 定义目标结构体
	var detail map[string]interface{}

	// 解压缩并解析
	err := DecompressToJSON(compressedContent, &detail)
	if err != nil {
		// 处理错误
		return
	}

	// 现在可以使用 detail 了
	println(detail)
}

// 示例4: 在保存 Packet 时使用压缩
func exampleSavePacketWithCompression() {
	// 创建数据包对象
	packet := &gorm.Packet{
		TaskID:      1,
		No:          1,
		FrameNumber: 1,
		Timestamp:   1234567890,
		Src:         "192.168.1.1",
		Dst:         "192.168.1.2",
		Protocol:    "TCP",
		Length:      1500,
		Info:        "Standard TCP packet",
	}

	// 准备详细数据
	detailData := map[string]interface{}{
		"layers": map[string]interface{}{
			"frame": map[string]string{
				"frame.number": "1",
				"frame.time":   "2024-01-01 12:00:00",
			},
			"ip": map[string]string{
				"ip.src": "192.168.1.1",
				"ip.dst": "192.168.1.2",
			},
		},
	}

	// 压缩详细数据
	compressedDetail, err := CompressJSON(detailData)
	if err != nil {
		// 处理错误
		return
	}

	// 设置压缩后的内容
	packet.Content = compressedDetail

	// 保存到数据库
	// db.Create(packet)
}

// 示例5: 从数据库读取并解压缩 Packet 详情
func exampleLoadPacketWithDecompression() {
	// 从数据库查询
	// var packet gorm.Packet
	// db.First(&packet, 1)

	packet := &gorm.Packet{
		Content: "...", // 从数据库加载的压缩内容
	}

	// 解压缩详情
	var detailData map[string]interface{}
	err := DecompressToJSON(packet.Content, &detailData)
	if err != nil {
		// 处理错误
		return
	}

	// 现在可以使用 detailData 了
	layers := detailData["layers"]
	println(layers)
}

// 示例6: 批量压缩（适用于大量数据包）
func exampleBatchCompression() {
	packets := []map[string]interface{}{
		{"frame": 1, "src": "192.168.1.1", "detail": "..."},
		{"frame": 2, "src": "192.168.1.2", "detail": "..."},
		// ... 更多数据包
	}

	for i, pkt := range packets {
		// 压缩每个数据包的详情
		compressed, err := CompressJSON(pkt)
		if err != nil {
			// 处理错误
			continue
		}

		// 存储压缩后的数据
		packets[i]["compressed_detail"] = compressed
	}
}

// 性能提示：
// 1. zstd 编码器/解码器已在 init() 中全局初始化，可重复使用
// 2. SpeedFastest 级别适合实时抓包场景，平衡速度和压缩率
// 3. 如果需要更高的压缩率，可以使用 zstd.SpeedDefault 或 zstd.SpeedBetterCompression
// 4. Base64 编码会增加约 33% 的大小，但便于存储和传输
// 5. 如果数据库支持 blob 类型，可以直接存储二进制压缩数据，跳过 base64 编码
*/
