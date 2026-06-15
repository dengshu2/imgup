package naming

import (
	"testing"
	"time"
)

var fixedTime = time.Date(2026, time.June, 15, 10, 30, 0, 0, time.UTC)

func TestResolveDefaultTemplate(t *testing.T) {
	// PNG magic header so content sniffing agrees with the extension.
	content := []byte("\x89PNG\r\n\x1a\nrest-of-image")
	obj, err := Resolve("{year}/{month}/{md5}.{ext}", content, "shot.PNG", fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	wantKey := "2026/06/" + obj.MD5 + ".png"
	if obj.Key != wantKey {
		t.Errorf("Key = %q, want %q", obj.Key, wantKey)
	}
	if obj.Ext != "png" {
		t.Errorf("Ext = %q, want png", obj.Ext)
	}
	if len(obj.MD5) != 32 {
		t.Errorf("MD5 = %q, want 32 hex chars", obj.MD5)
	}
}

func TestResolveDeterministic(t *testing.T) {
	content := []byte("identical bytes")
	a, _ := Resolve("{md5}.{ext}", content, "a.jpg", fixedTime)
	b, _ := Resolve("{md5}.{ext}", content, "b.jpg", fixedTime)
	if a.MD5 != b.MD5 || a.Key != b.Key {
		t.Errorf("same content produced different keys: %q vs %q", a.Key, b.Key)
	}
}

func TestResolveExtFromContent(t *testing.T) {
	// GIF header, no filename extension available (stdin case).
	content := []byte("GIF89a----------")
	obj, err := Resolve("{md5}.{ext}", content, "", fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Ext != "gif" {
		t.Errorf("Ext = %q, want gif (sniffed)", obj.Ext)
	}
}

func TestResolveEmptyKeyErrors(t *testing.T) {
	if _, err := Resolve("", []byte("x"), "a.png", fixedTime); err == nil {
		t.Error("expected error for empty template, got nil")
	}
}
