package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// GenerateSessionID 生成会话 ID
func GenerateSessionID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateEventID 生成事件 ID
func GenerateEventID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// UUIDToString 将 uuid.UUID 转为字符串，Nil 返回空字符串（用于写入 DB/API 等可选 UUID 字段）
func UUIDToString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

