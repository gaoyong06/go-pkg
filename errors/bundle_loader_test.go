package errors

import (
	"testing"
	"testing/fstest"
)

func TestBundleErrorMessageLoaderUsesLanguageFallback(t *testing.T) {
	loader, err := NewBundleErrorMessageLoaderFromFS(fstest.MapFS{
		"active.en.json":    &fstest.MapFile{Data: []byte(`{"240001":"Invalid request"}`)},
		"active.zh-CN.json": &fstest.MapFile{Data: []byte(`{"240001":"请求无效"}`)},
	}, "en")
	if err != nil {
		t.Fatal(err)
	}
	if got := loader.GetMessage("zh-CN", 240001); got != "请求无效" {
		t.Fatalf("Chinese message = %q", got)
	}
	if got := loader.GetMessage("de-DE", 240001); got != "Invalid request" {
		t.Fatalf("fallback message = %q", got)
	}
}
