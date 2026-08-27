package errors

import (
	"fmt"
	"io/fs"
	"strconv"

	pkgI18n "github.com/gaoyong06/go-pkg/middleware/i18n"
)

// BundleErrorMessageLoader 使用标准 go-i18n catalog 加载错误消息。
type BundleErrorMessageLoader struct {
	translator *pkgI18n.BundleTranslator
}

// NewBundleErrorMessageLoaderFromFS 创建基于 fs.FS 的错误消息加载器。
// catalog 文件名必须使用 active.<locale>.json，消息 ID 使用数字错误码。
func NewBundleErrorMessageLoaderFromFS(fsys fs.FS, defaultLanguage string) (ErrorMessageLoader, error) {
	translator, err := pkgI18n.NewBundleTranslatorFromFS(fsys, defaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("create error catalog: %w", err)
	}
	return &BundleErrorMessageLoader{translator: translator}, nil
}

// GetMessage 根据语言和错误码获取消息，并按 Bundle 的默认语言回退。
func (l *BundleErrorMessageLoader) GetMessage(lang string, code int32) string {
	if l == nil || l.translator == nil {
		return strconv.FormatInt(int64(code), 10)
	}
	return l.translator.TranslateLanguage(lang, strconv.FormatInt(int64(code), 10), "", nil)
}
