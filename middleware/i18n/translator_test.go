package i18n

import (
	"testing"
	"testing/fstest"
)

func TestBundleTranslatorLoadsStandardCatalogNames(t *testing.T) {
	translator, err := NewBundleTranslatorFromFS(fstest.MapFS{
		"active.en.json":    &fstest.MapFile{Data: []byte(`{"hello":"Hello"}`)},
		"active.zh-CN.json": &fstest.MapFile{Data: []byte(`{"hello":"你好"}`)},
	}, "en")
	if err != nil {
		t.Fatal(err)
	}
	if got := translator.TranslateLanguage("zh-CN", "hello", "", nil); got != "你好" {
		t.Fatalf("translated text = %q, want 你好", got)
	}
	if got := translator.TranslateLanguage("de-DE", "hello", "", nil); got != "Hello" {
		t.Fatalf("fallback text = %q, want Hello", got)
	}
	if got := translator.TranslateLanguage("en", "missing", "Default", nil); got != "Default" {
		t.Fatalf("default text = %q, want Default", got)
	}
}

func TestBundleTranslatorRequiresDefaultCatalog(t *testing.T) {
	if _, err := NewBundleTranslatorFromFS(fstest.MapFS{
		"active.zh-CN.json": &fstest.MapFile{Data: []byte(`{"hello":"你好"}`)},
	}, "en"); err == nil {
		t.Fatal("expected missing default catalog error")
	}
}

func TestBundleTranslatorRejectsDuplicateCatalogLanguages(t *testing.T) {
	if _, err := NewBundleTranslatorFromFS(fstest.MapFS{
		"active.en.json":        &fstest.MapFile{Data: []byte(`{"hello":"Hello"}`)},
		"nested/active.en.json": &fstest.MapFile{Data: []byte(`{"hello":"Hi"}`)},
	}, "en"); err == nil {
		t.Fatal("expected duplicate language error")
	}
}
