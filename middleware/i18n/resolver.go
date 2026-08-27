package i18n

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

// ResolverConfig 描述一个服务实际提供的语言集合和默认语言。
// 语言标签使用 BCP 47 格式，例如 en、zh-CN、zh-TW。
type ResolverConfig struct {
	DefaultLanguage    string
	SupportedLanguages []string
}

// Resolver 将 Accept-Language 解析为服务支持的规范化语言标签。
// Resolver 创建后只读，可以安全地在多个请求之间复用。
type Resolver struct {
	defaultTag language.Tag
	supported  []language.Tag
	matcher    language.Matcher
}

// NewResolver 创建请求语言解析器。
func NewResolver(config ResolverConfig) (*Resolver, error) {
	defaultTag, err := parseTag(config.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid default language: %w", err)
	}

	supported := make([]language.Tag, 0, len(config.SupportedLanguages)+1)
	seen := make(map[language.Tag]struct{}, len(config.SupportedLanguages)+1)
	appendTag := func(raw string) error {
		tag, err := parseTag(raw)
		if err != nil {
			return err
		}
		if _, ok := seen[tag]; ok {
			return nil
		}
		seen[tag] = struct{}{}
		supported = append(supported, tag)
		return nil
	}

	for _, raw := range config.SupportedLanguages {
		if err := appendTag(raw); err != nil {
			return nil, fmt.Errorf("invalid supported language %q: %w", raw, err)
		}
	}
	if len(supported) == 0 {
		supported = append(supported, defaultTag)
		seen[defaultTag] = struct{}{}
	}
	if _, ok := seen[defaultTag]; !ok {
		supported = append([]language.Tag{defaultTag}, supported...)
	}

	return &Resolver{
		defaultTag: defaultTag,
		supported:  supported,
		matcher:    language.NewMatcher(supported),
	}, nil
}

// MustNewResolver 创建解析器；配置无效时直接 panic，适合服务启动阶段使用。
func MustNewResolver(config ResolverConfig) *Resolver {
	resolver, err := NewResolver(config)
	if err != nil {
		panic(err)
	}
	return resolver
}

// DefaultLanguage 返回规范化的默认语言标签。
func (r *Resolver) DefaultLanguage() string {
	if r == nil {
		return ""
	}
	return r.defaultTag.String()
}

// SupportedLanguages 返回解析器支持的语言标签副本。
func (r *Resolver) SupportedLanguages() []string {
	if r == nil {
		return nil
	}
	result := make([]string, 0, len(r.supported))
	for _, tag := range r.supported {
		result = append(result, tag.String())
	}
	return result
}

// Resolve 根据 Accept-Language 解析最匹配的服务语言。
// 无法解析或没有匹配项时返回配置的默认语言。
func (r *Resolver) Resolve(acceptLanguage string) string {
	if r == nil {
		return ""
	}
	acceptLanguage = strings.ReplaceAll(strings.TrimSpace(acceptLanguage), "_", "-")
	if acceptLanguage == "" {
		return r.DefaultLanguage()
	}

	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return r.DefaultLanguage()
	}
	_, index, confidence := r.matcher.Match(tags...)
	if confidence == language.No {
		return r.DefaultLanguage()
	}
	if index < 0 || index >= len(r.supported) {
		return r.DefaultLanguage()
	}
	return r.supported[index].String()
}

// Normalize 将单个语言标签规范化为 BCP 47 字符串；Accept-Language 应使用 Resolve。
func Normalize(raw string) string {
	tag, err := parseTag(raw)
	if err != nil {
		return ""
	}
	return tag.String()
}

func parseTag(raw string) (language.Tag, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "_", "-"))
	if raw == "" || raw == "*" || strings.ContainsAny(raw, ",;") {
		return language.Und, fmt.Errorf("language tag must be a single BCP 47 tag")
	}
	tag, err := language.Parse(raw)
	if err != nil || tag == language.Und {
		if err == nil {
			err = fmt.Errorf("unknown language tag")
		}
		return language.Und, err
	}
	return tag, nil
}
