// Package config defines imgup's configuration model and its persistence in a
// TOML file at ~/.config/imgup/config.toml. It supports multiple named profiles
// so a single install can target several buckets or, later, several backends.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// File is the on-disk representation of the whole config.
type File struct {
	// DefaultProfile names the profile used when none is given on the command line.
	DefaultProfile string `toml:"default_profile"`
	// Profiles is keyed by profile name.
	Profiles map[string]*Profile `toml:"profiles"`
}

// Profile holds the settings for a single upload target.
//
// SecretAccessKey is intentionally omitted from the file in the common case:
// on macOS the secret lives in the Keychain and UseKeychain is set to true.
// It may still be stored inline for portability (Linux/CI), in which case the
// config file should be mode 0600.
type Profile struct {
	// Backend selects the storage implementation. Only "r2" exists today; the
	// field is present so future backends can be added without a schema break.
	Backend string `toml:"backend"`

	Endpoint    string `toml:"endpoint"`
	Bucket      string `toml:"bucket"`
	PublicBase  string `toml:"public_base"`
	KeyTemplate string `toml:"key_template"`

	AccessKeyID string `toml:"access_key_id"`
	// SecretAccessKey is the inline secret. Empty when UseKeychain is true.
	SecretAccessKey string `toml:"secret_access_key,omitempty"`
	// UseKeychain, when true, tells imgup to read the secret from the OS keyring
	// (service "imgup", account = profile name) instead of from this file.
	UseKeychain bool `toml:"use_keychain,omitempty"`
}

const (
	// DefaultKeyTemplate matches the historical PicList naming convention.
	DefaultKeyTemplate = "{year}/{month}/{md5}.{ext}"
	// DefaultProfileName is used when default_profile is unset.
	DefaultProfileName = "default"
	// DefaultBackend is the only backend shipped today.
	DefaultBackend = "r2"
)

// Path returns the resolved config file path, honoring $IMGUP_CONFIG and then
// $XDG_CONFIG_HOME, falling back to ~/.config/imgup/config.toml.
func Path() (string, error) {
	if p := os.Getenv("IMGUP_CONFIG"); p != "" {
		return p, nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "imgup", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "imgup", "config.toml"), nil
}

// Load reads the config file. A missing file yields an empty (non-nil) File so
// callers can still layer env/flag values on top.
func Load(path string) (*File, error) {
	f := &File{Profiles: map[string]*Profile{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(data, f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]*Profile{}
	}
	return f, nil
}

// Save writes the config file atomically with 0600 permissions, creating parent
// directories as needed.
func Save(path string, f *File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.toml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := toml.NewEncoder(tmp).Encode(f); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// Profile returns the named profile, or the default profile when name is empty.
func (f *File) Profile(name string) (*Profile, string, bool) {
	if name == "" {
		name = f.DefaultProfile
	}
	if name == "" {
		name = DefaultProfileName
	}
	p, ok := f.Profiles[name]
	return p, name, ok
}
