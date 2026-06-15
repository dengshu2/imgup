package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	in := &File{
		DefaultProfile: "default",
		Profiles: map[string]*Profile{
			"default": {
				Backend:     "r2",
				Endpoint:    "https://acct.r2.cloudflarestorage.com",
				Bucket:      "img",
				PublicBase:  "https://img.example.com",
				KeyTemplate: DefaultKeyTemplate,
				AccessKeyID: "akid",
				UseKeychain: true,
			},
		},
	}
	if err := Save(path, in); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode = %#o, want 0600", perm)
	}

	out, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	p, name, ok := out.Profile("")
	if !ok || name != "default" {
		t.Fatalf("Profile() = %q, ok=%v", name, ok)
	}
	if p.Bucket != "img" || !p.UseKeychain || p.PublicBase != "https://img.example.com" {
		t.Errorf("round-trip mismatch: %+v", p)
	}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	f, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Profiles) != 0 {
		t.Errorf("expected empty profiles, got %d", len(f.Profiles))
	}
}

func TestResolvePrecedence(t *testing.T) {
	prof := &Profile{
		Endpoint:        "file-endpoint",
		Bucket:          "file-bucket",
		PublicBase:      "file-base",
		AccessKeyID:     "file-akid",
		SecretAccessKey: "file-secret",
	}

	t.Setenv("IMGUP_BUCKET", "env-bucket")
	t.Setenv("IMGUP_SECRET_ACCESS_KEY", "env-secret")

	r, err := Resolve("default", prof, Overrides{Endpoint: "flag-endpoint"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Endpoint != "flag-endpoint" {
		t.Errorf("endpoint: flag should win, got %q", r.Endpoint)
	}
	if r.Bucket != "env-bucket" {
		t.Errorf("bucket: env should win over file, got %q", r.Bucket)
	}
	if r.PublicBase != "file-base" {
		t.Errorf("public base: file should be used, got %q", r.PublicBase)
	}
	if r.SecretAccessKey != "env-secret" || r.SecretSource != "env" {
		t.Errorf("secret: env should win, got %q from %q", r.SecretAccessKey, r.SecretSource)
	}
}

func TestValidateMissingField(t *testing.T) {
	r := &Resolved{Endpoint: "e", Bucket: "b", PublicBase: "p", AccessKeyID: "a"}
	if err := r.Validate(); err == nil {
		t.Error("expected error for missing secret_access_key")
	}
}
