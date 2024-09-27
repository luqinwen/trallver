package service

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"my_project/server/internal/model"
)

// ToBinaryTask 将 ProbeTask 结构体序列化为二进制数据
func ToBinaryTask(task *model.ProbeTask) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, task)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBinaryTask 从二进制数据反序列化为 ProbeTask 结构体
func FromBinaryTask(data []byte, task *model.ProbeTask) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, task)
	return err
}

// ToBinaryResult 将 ProbeResult 结构体序列化为二进制数据
func ToBinaryResult(result *model.ProbeResult) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, result)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBinaryResult 从二进制数据反序列化为 ProbeResult 结构体
func FromBinaryResult(data []byte, result *model.ProbeResult) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, result)
	return err
}

// Compress 压缩二进制数据
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress 解压缩数据
func Decompress(compressedData []byte) ([]byte, error) {
	buf := bytes.NewReader(compressedData)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	decompressedData, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	return decompressedData, nil
}
