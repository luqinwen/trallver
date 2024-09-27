package model

// 打包 PacketLoss 和 Threshold 到 Packed 字段（16位）
func PackResultFields(packetLoss uint8, threshold uint8) uint16 {
    var packed uint16
    packed |= uint16(packetLoss)        // PacketLoss 占用低8位
    packed |= uint16(threshold) << 8    // Threshold 占用高8位
    return packed
}

// 解包固定字段
func UnpackFixedFields(packed uint32) (uint16, uint8, uint8) {
    timeout := uint16(packed & 0xFFFF)            // 提取低16位的 Timeout
    count := uint8((packed >> 16) & 0xFF)         // 提取中间8位的 Count
    threshold := uint8((packed >> 24) & 0xFF)     // 提取高8位的 Threshold
    return timeout, count, threshold
}


