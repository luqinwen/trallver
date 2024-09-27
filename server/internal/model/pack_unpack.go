package model

// 打包固定字段 (服务端使用)
func PackFixedFields(timeout uint16, count uint8, threshold uint8) uint32 {
    var packed uint32
    packed |= uint32(timeout)           // Timeout占用低16位
    packed |= uint32(count) << 16       // Count占用17-24位
    packed |= uint32(threshold) << 24   // Threshold占用25-32位
    return packed
}

// 解包探测结果字段 (服务端使用)
// 从 uint16 中解包 PacketLoss 和 Threshold
func UnpackResultFields(packed uint16) (uint8, uint8) {
    packetLoss := uint8(packed & 0xFF)        // 提取低8位的 PacketLoss
    threshold := uint8((packed >> 8) & 0xFF)  // 提取高8位的 Threshold
    return packetLoss, threshold
}

