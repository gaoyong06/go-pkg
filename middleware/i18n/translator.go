// Package i18n 提供基于 go-i18n 的翻译服务。
package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// BundleTranslator 持有一个只读的 go-i18n Bundle 和请求语言解析器。
// BundleTranslator 创建完成后不可变，可以安全地在多个请求之间复用。
type BundleTranslator struct {
	bundle   *goi18n.Bundle
	resolver *Resolver
}

// NewBundleTranslator 从 configDir/i18n 加载标准命名的 catalog 文件。
// configDir 通常为服务工作目录；资源文件名必须是 active.<locale>.json。
//
// Deprecated: 使用 NewBundleTranslatorWithDefault 显式传入服务默认语言。
func NewBundleTranslator(configDir string) (*BundleTranslator, error) {
	return NewBundleTranslatorWithDefault(configDir, "zh-CN")
}

// NewBundleTranslatorWithDefault 从 configDir/i18n 加载标准命名的 catalog 文件。
func NewBundleTranslatorWithDefault(configDir, defaultLanguage string) (*BundleTranslator, error) {
	root := filepath.Join(configDir, "i18n")
	if info, err := os.Stat(configDir); err == nil && info.IsDir() && filepath.Base(filepath.Clean(configDir)) == "i18n" {
		root = configDir
	}
	return newBundleTranslator(os.DirFS(root), defaultLanguage)
}

// NewBundleTranslatorFromFS 从嵌入式或其他 fs.FS 加载 catalog。
// fsys 中的文件可以位于任意子目录，但必须使用 active.<locale>.json 命名。
func NewBundleTranslatorFromFS(fsys fs.FS, defaultLanguage string) (*BundleTranslator, error) {
	return newBundleTranslator(fsys, defaultLanguage)
}

func newBundleTranslator(fsys fs.FS, defaultLanguage string) (*BundleTranslator, error) {
	if fsys == nil {
		return nil, fmt.Errorf("translation filesystem is nil")
	}
	defaultTag, err := parseTag(defaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid translator default language: %w", err)
	}

	bundle := goi18n.NewBundle(defaultTag)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	loadedTags := make([]string, 0)
	loadedTagSet := make(map[string]struct{})
	err = fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		baseName := filepath.Base(path)
		if entry.IsDir() || !strings.HasPrefix(baseName, "active.") || !strings.EqualFold(filepath.Ext(path), ".json") {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read translation file %q: %w", path, err)
		}
		messageFile, err := bundle.ParseMessageFileBytes(data, filepath.ToSlash(path))
		if err != nil {
			return fmt.Errorf("parse translation file %q: %w", path, err)
		}
		if messageFile.Tag == language.Und {
			return fmt.Errorf("translation file %q has an invalid language tag", path)
		}
		loadedTag := messageFile.Tag.String()
		if _, exists := loadedTagSet[loadedTag]; exists {
			return fmt.Errorf("duplicate translation language %q in file %q", loadedTag, path)
		}
		loadedTagSet[loadedTag] = struct{}{}
		loadedTags = append(loadedTags, loadedTag)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(loadedTags) == 0 {
		return nil, fmt.Errorf("no JSON translation files found")
	}
	if _, ok := loadedTagSet[defaultTag.String()]; !ok {
		return nil, fmt.Errorf("default translation language %q is not loaded", defaultTag.String())
	}

	resolver, err := NewResolver(ResolverConfig{
		DefaultLanguage:    defaultTag.String(),
		SupportedLanguages: loadedTags,
	})
	if err != nil {
		return nil, fmt.Errorf("build translator language resolver: %w", err)
	}
	return &BundleTranslator{bundle: bundle, resolver: resolver}, nil
}

// SupportedLanguages 返回 catalog 中已加载的语言标签。
func (t *BundleTranslator) SupportedLanguages() []string {
	if t == nil || t.resolver == nil {
		return nil
	}
	languages := t.resolver.SupportedLanguages()
	sort.Strings(languages)
	return languages
}

// ResolveLanguage 将 Accept-Language 解析为 catalog 中的语言标签。
func (t *BundleTranslator) ResolveLanguage(acceptLanguage string) string {
	if t == nil || t.resolver == nil {
		return ""
	}
	return t.resolver.Resolve(acceptLanguage)
}

// Translate 根据 context 中的语言翻译消息。
func (t *BundleTranslator) Translate(ctx context.Context, key string, templateData map[string]interface{}) string {
	if ctx == nil {
		ctx = context.Background()
	}
	return t.TranslateLanguage(Language(ctx), key, "", templateData)
}

// TranslateLanguage 使用指定语言翻译消息；缺失时返回 defaultMessage，未提供默认值则返回 key。
func (t *BundleTranslator) TranslateLanguage(lang, key, defaultMessage string, templateData map[string]interface{}) string {
	if t == nil || t.bundle == nil || t.resolver == nil || strings.TrimSpace(key) == "" {
		return fallbackMessage(key, defaultMessage)
	}
	if strings.TrimSpace(lang) == "" {
		lang = t.resolver.DefaultLanguage()
	}
	localizer := goi18n.NewLocalizer(t.bundle, lang)
	config := &goi18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	}
	if defaultMessage != "" {
		config.DefaultMessage = &goi18n.Message{ID: key, Other: defaultMessage}
	}
	translated, err := localizer.Localize(config)
	if err != nil || strings.TrimSpace(translated) == "" {
		return fallbackMessage(key, defaultMessage)
	}
	return translated
}

// TranslateWithDefault 是 TranslateLanguage 的 context 版本。
func (t *BundleTranslator) TranslateWithDefault(ctx context.Context, key, defaultMessage string, templateData map[string]interface{}) string {
	if ctx == nil {
		ctx = context.Background()
	}
	return t.TranslateLanguage(Language(ctx), key, defaultMessage, templateData)
}

func fallbackMessage(key, defaultMessage string) string {
	if defaultMessage != "" {
		return defaultMessage
	}
	return key
}
