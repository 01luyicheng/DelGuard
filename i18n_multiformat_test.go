package main

import (
	"path/filepath"
	"testing"
)

func TestLoadLanguagePacks_MultiFormat(t *testing.T) {
	// 重新加载语言包目录（包含我们新增的多格式样例）
	dir := filepath.Join("config", "languages")
	if err := LoadLanguagePacks(dir); err != nil {
		t.Fatalf("LoadLanguagePacks failed: %v", err)
	}

	// de-DE (properties)
	SetLocale("de-DE")
	out := T("删除 %s ? [y/N/a/q]: ")
	if want := "DelGuard: Datei %s löschen? [y/N/a/q]: "; out != want {
		t.Fatalf("de-DE translation mismatch: got %q, want %q", out, want)
	}

	// fr-FR (ini)
	SetLocale("fr-FR")
	out = T("删除 %s ? [y/N/a/q]: ")
	if want := "DelGuard: Supprimer %s ? [y/N/a/q]: "; out != want {
		t.Fatalf("fr-FR translation mismatch: got %q, want %q", out, want)
	}

	// es-ES (jsonc)
	SetLocale("es-ES")
	out = T("删除 %s ? [y/N/a/q]: ")
	if want := "DelGuard: ¿Eliminar %s? [y/N/a/q]: "; out != want {
		t.Fatalf("es-ES translation mismatch: got %q, want %q", out, want)
	}
}

func TestI18n_FallbackToEnglish(t *testing.T) {
	// 设置为一个不存在的语言，应该回退到 en-US
	SetLocale("xx-YY")
	out := T("删除 %s ? [y/N/a/q]: ")
	if want := "DelGuard: Delete %s ? [y/N/a/q]: "; out != want {
		t.Fatalf("fallback translation mismatch: got %q, want %q", out, want)
	}
}
