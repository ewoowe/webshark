package utils

import (
	"encoding/base64"
	"encoding/json"

	"github.com/klauspost/compress/zstd"
)

var (
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
)

func init() {
	// 初始化 zstd 编码器和解码器
	var err error
	// 使用 SpeedFastest 级别以获得最快的压缩速度，适合实时抓包场景
	zstdEncoder, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		panic("failed to create zstd encoder: " + err.Error())
	}

	zstdDecoder, err = zstd.NewReader(nil)
	if err != nil {
		panic("failed to create zstd decoder: " + err.Error())
	}
}

// CompressZstd 使用 zstd 压缩数据
// 输入：原始字节数组
// 输出：压缩后的字节数组
func CompressZstd(data []byte) ([]byte, error) {
	return zstdEncoder.EncodeAll(data, nil), nil
}

// DecompressZstd 使用 zstd 解压缩数据
// 输入：压缩的字节数组
// 输出：解压后的原始字节数组
func DecompressZstd(data []byte) ([]byte, error) {
	return zstdDecoder.DecodeAll(data, nil)
}

// CompressString 压缩字符串
// 适用于压缩 JSON、XML 等文本内容
// 返回 base64 编码的字符串，方便存储到数据库
func CompressString(content string) (string, error) {
	compressed, err := CompressZstd([]byte(content))
	if err != nil {
		return "", err
	}
	// 将二进制压缩数据转换为 base64 字符串，便于存储
	return base64.StdEncoding.EncodeToString(compressed), nil
}

// DecompressString 解压缩字符串
// 适用于从数据库读取压缩的文本内容
func DecompressString(compressedBase64 string) (string, error) {
	// 先解码 base64
	compressed, err := base64.StdEncoding.DecodeString(compressedBase64)
	if err != nil {
		return "", err
	}
	// 再解压缩
	decompressed, err := DecompressZstd(compressed)
	if err != nil {
		return "", err
	}
	return string(decompressed), nil
}

// CompressJSON 压缩 JSON 对象
// 直接将任意结构体压缩为 base64 字符串
func CompressJSON(obj interface{}) (string, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return CompressString(string(jsonData))
}

// DecompressToJSON 解压缩并解析为 JSON 对象
// 从 base64 字符串解压缩并反序列化到目标结构体
func DecompressToJSON(compressedBase64 string, target interface{}) error {
	jsonStr, err := DecompressString(compressedBase64)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonStr), target)
}
