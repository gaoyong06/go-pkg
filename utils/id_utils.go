package utils

import (
	"crypto/rand"
	"encoding/hex"
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

