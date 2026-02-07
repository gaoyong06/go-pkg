// Package errors 提供错误消息加载功能
package errors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// ErrorMessageLoader 错误消息加载接口
type ErrorMessageLoader interface {
	// GetMessage 根据语言和错误码获取错误消息
	GetMessage(lang string, code int32) string
}

// JSONErrorMessageLoader 从 JSON 文件加载错误消息
type JSONErrorMessageLoader struct {
	configDir string // i18n 配置目录，例如 "i18n"
	cache     map[string]map[string]string // lang -> code -> message
	mutex     sync.RWMutex
}

// NewJSONErrorMessageLoader 创建 JSON 错误消息加载器
// configDir: i18n 配置目录，例如 "i18n"
func NewJSONErrorMessageLoader(configDir string) ErrorMessageLoader {
	return &JSONErrorMessageLoader{
		configDir: configDir,
		cache:     make(map[string]map[string]string),
	}
}

// defaultMessages 常见错误码的默认文案（i18n 未加载或未配置时使用，便于排查）
var defaultMessages = map[int32]string{
	110501: "无效的 Token（未提供或无效的登录凭证，请重新登录）",
	110502: "Token 已失效（请重新登录）",
}

// getDefaultMessage 返回错误码的默认描述文案，便于问题排查
func getDefaultMessage(code int32) string {
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return fmt.Sprintf("错误码 %d（未找到对应文案，请检查服务 i18n 配置或联系开发）", code)
}

// GetMessage 获取错误消息
func (l *JSONErrorMessageLoader) GetMessage(lang string, code int32) string {
	// 尝试从缓存获取
	l.mutex.RLock()
	if langMessages, ok := l.cache[lang]; ok {
		if message, ok := langMessages[fmt.Sprintf("%d", code)]; ok {
			l.mutex.RUnlock()
			return message
		}
	}
	l.mutex.RUnlock()

	// 加载错误信息配置文件
	err := l.loadErrorMessages(lang)
	if err != nil {
		// 如果加载失败，尝试使用默认语言(zh-CN)
		if lang != "zh-CN" {
			return l.GetMessage("zh-CN", code)
		}
		return getDefaultMessage(code)
	}

	// 再次尝试从缓存获取
	l.mutex.RLock()
	if langMessages, ok := l.cache[lang]; ok {
		if message, ok := langMessages[fmt.Sprintf("%d", code)]; ok {
			l.mutex.RUnlock()
			return message
		}
	}
	l.mutex.RUnlock()

	// 如果仍然找不到，尝试使用默认语言
	if lang != "zh-CN" {
		return l.GetMessage("zh-CN", code)
	}

	return getDefaultMessage(code)
}

// ErrorMessageConfig 错误信息配置结构
type ErrorMessageConfig struct {
	Errors map[string]string `json:"errors"`
}

// loadErrorMessages 加载错误信息配置文件
func (l *JSONErrorMessageLoader) loadErrorMessages(lang string) error {
	// 检查缓存
	l.mutex.RLock()
	if _, ok := l.cache[lang]; ok {
		l.mutex.RUnlock()
		return nil
	}
	l.mutex.RUnlock()

	// 确定配置文件路径
	configDir := filepath.Join(l.configDir, lang)
	filePath := filepath.Join(configDir, "errors.json")

	// 检查文件是否存在
	if _, err := ioutil.ReadFile(filePath); err != nil {
		return fmt.Errorf("无法找到错误信息配置文件: %v", err)
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取错误信息配置失败: %v", err)
	}

	// 解析JSON
	var config ErrorMessageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析错误信息配置失败: %v", err)
	}

	// 缓存结果
	l.mutex.Lock()
	l.cache[lang] = config.Errors
	l.mutex.Unlock()

	return nil
}

