package i18n

import (
	"context"
	"testing"
)

func TestResolverHonorsAcceptLanguageQuality(t *testing.T) {
	resolver, err := NewResolver(ResolverConfig{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "zh-CN", "zh-TW"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		acceptLanguage string
		want           string
	}{
		{name: "quality wins", acceptLanguage: "en;q=0.2,zh-CN;q=0.9", want: "zh-CN"},
		{name: "region match", acceptLanguage: "zh-TW,zh;q=0.8", want: "zh-TW"},
		{name: "unknown falls back", acceptLanguage: "de-DE,de;q=0.9", want: "en"},
		{name: "zero quality is ignored", acceptLanguage: "zh-CN;q=0,en", want: "en"},
		{name: "underscore is normalized", acceptLanguage: "zh_CN", want: "zh-CN"},
		{name: "wildcard uses default", acceptLanguage: "*", want: "en"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolver.Resolve(test.acceptLanguage); got != test.want {
				t.Fatalf("Resolve(%q) = %q, want %q", test.acceptLanguage, got, test.want)
			}
		})
	}
}

func TestNewResolverRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewResolver(ResolverConfig{DefaultLanguage: "not a language"}); err == nil {
		t.Fatal("expected invalid default language error")
	}
	if _, err := NewResolver(ResolverConfig{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"bad language"},
	}); err == nil {
		t.Fatal("expected invalid supported language error")
	}
}

func TestLanguageWithDefaultUsesServiceFallback(t *testing.T) {
	if got := LanguageWithDefault(context.Background(), "en"); got != "en" {
		t.Fatalf("LanguageWithDefault() = %q, want en", got)
	}
	ctx := WithLanguage(context.Background(), "zh-CN")
	if got := LanguageWithDefault(ctx, "en"); got != "zh-CN" {
		t.Fatalf("LanguageWithDefault() with context = %q, want zh-CN", got)
	}
}
