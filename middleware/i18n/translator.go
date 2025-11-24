// Package i18n 提供国际化（i18n）翻译服务
package i18n

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// BundleTranslator 基于 go-i18n Bundle 的翻译器实现
type BundleTranslator struct {
	bundle *i18n.Bundle
	mutex  sync.RWMutex
}

// NewBundleTranslator 创建 Bundle 翻译器
// configDir: 配置文件目录（如 "." 表示当前目录）
// 会自动加载 configDir/i18n/{lang}/*.json 文件
func NewBundleTranslator(configDir string) (*BundleTranslator, error) {
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 加载所有语言的翻译文件
	i18nDir := filepath.Join(configDir, "i18n")
	entries, err := ioutil.ReadDir(i18nDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				lang := entry.Name()
				// 加载该语言目录下的所有 JSON 文件
				langDir := filepath.Join(i18nDir, lang)
				files, err := ioutil.ReadDir(langDir)
				if err == nil {
					for _, file := range files {
						if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
							filePath := filepath.Join(langDir, file.Name())
							_, err := bundle.LoadMessageFile(filePath)
							if err != nil {
								// 忽略加载错误，继续加载其他文件
								continue
							}
						}
					}
				}
			}
		}
	}

	return &BundleTranslator{
		bundle: bundle,
	}, nil
}

// Translate 实现 Translator 接口
func (t *BundleTranslator) Translate(ctx context.Context, key string, templateData map[string]interface{}) string {
	lang := Language(ctx)
	
	t.mutex.RLock()
	bundle := t.bundle
	t.mutex.RUnlock()
	
	if bundle == nil {
		return key
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	config := &i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	}

	translated, err := localizer.Localize(config)
	if err != nil {
		// 如果翻译失败，尝试使用默认语言
		if lang != "zh-CN" {
			ctx = WithLanguage(ctx, "zh-CN")
			return t.Translate(ctx, key, templateData)
		}
		// 如果默认语言也失败，返回 key
		return key
	}

	return translated
}

// TranslateWithDefault 带默认值的翻译函数
func (t *BundleTranslator) TranslateWithDefault(ctx context.Context, key string, defaultMessage string, templateData map[string]interface{}) string {
	lang := Language(ctx)
	
	t.mutex.RLock()
	bundle := t.bundle
	t.mutex.RUnlock()
	
	if bundle == nil {
		return defaultMessage
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	config := &i18n.LocalizeConfig{
		MessageID: key,
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: defaultMessage,
		},
		TemplateData: templateData,
	}

	translated, err := localizer.Localize(config)
	if err != nil {
		return defaultMessage
	}

	return translated
}

