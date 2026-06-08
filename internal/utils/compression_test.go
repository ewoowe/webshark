package utils

import (
	"testing"
)

func TestCompressDecompressString(t *testing.T) {
	original := `{"layers":{"frame":{"frame.number":"1","frame.time":"2024-01-01 12:00:00"},"ip":{"ip.src":"192.168.1.1","ip.dst":"192.168.1.2"}}}`

	// 压缩
	compressed, err := CompressString(original)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	t.Logf("原始大小: %d bytes, 压缩后大小: %d bytes", len(original), len(compressed))
	t.Logf("压缩率: %.2f%%", float64(len(compressed))/float64(len(original))*100)

	// 解压缩
	decompressed, err := DecompressString(compressed)
	if err != nil {
		t.Fatalf("解压缩失败: %v", err)
	}

	// 验证
	if decompressed != original {
		t.Errorf("解压后的内容与原始内容不匹配")
		t.Logf("原始: %s", original)
		t.Logf("解压: %s", decompressed)
	}
}

func TestCompressDecompressJSON(t *testing.T) {
	type TestData struct {
		Name   string            `json:"name"`
		Age    int               `json:"age"`
		Extras map[string]string `json:"extras"`
	}

	original := TestData{
		Name: "test",
		Age:  25,
		Extras: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// 压缩
	compressed, err := CompressJSON(original)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	t.Logf("压缩后的 base64: %s", compressed)

	// 解压缩
	var result TestData
	err = DecompressToJSON(compressed, &result)
	if err != nil {
		t.Fatalf("解压缩失败: %v", err)
	}

	// 验证
	if result.Name != original.Name {
		t.Errorf("Name 不匹配: 期望 %s, 得到 %s", original.Name, result.Name)
	}
	if result.Age != original.Age {
		t.Errorf("Age 不匹配: 期望 %d, 得到 %d", original.Age, result.Age)
	}
	if len(result.Extras) != len(original.Extras) {
		t.Errorf("Extras 长度不匹配: 期望 %d, 得到 %d", len(original.Extras), len(result.Extras))
	}
}

func TestCompressLargeData(t *testing.T) {
	// 测试较大数据的压缩
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	compressed, err := CompressZstd(largeData)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	t.Logf("原始大小: %d bytes, 压缩后大小: %d bytes", len(largeData), len(compressed))
	t.Logf("压缩率: %.2f%%", float64(len(compressed))/float64(len(largeData))*100)

	// 解压缩
	decompressed, err := DecompressZstd(compressed)
	if err != nil {
		t.Fatalf("解压缩失败: %v", err)
	}

	// 验证
	if len(decompressed) != len(largeData) {
		t.Errorf("解压后长度不匹配: 期望 %d, 得到 %d", len(largeData), len(decompressed))
	}

	for i := range largeData {
		if decompressed[i] != largeData[i] {
			t.Errorf("数据在位置 %d 不匹配", i)
			break
		}
	}
}

func BenchmarkCompressString(b *testing.B) {
	data := `{"layers":{"frame":{"frame.number":"1","frame.time":"2024-01-01 12:00:00"},"ip":{"ip.src":"192.168.1.1","ip.dst":"192.168.1.2"},"tcp":{"tcp.srcport":"80","tcp.dstport":"8080"}}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompressString(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompressString(b *testing.B) {
	data := `{"layers":{"frame":{"frame.number":"1","frame.time":"2024-01-01 12:00:00"},"ip":{"ip.src":"192.168.1.1","ip.dst":"192.168.1.2"},"tcp":{"tcp.srcport":"80","tcp.dstport":"8080"}}}`

	compressed, _ := CompressString(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecompressString(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}
