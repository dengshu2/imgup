package config

import (
	"fmt"
	"os"
)

// Resolved is a fully materialized profile ready to hand to a backend. Every
// field has been resolved through the precedence chain.
type Resolved struct {
	Profile string

	Backend     string
	Endpoint    string
	Bucket      string
	PublicBase  string
	KeyTemplate string

	AccessKeyID     string
	SecretAccessKey string

	// SecretSource records where the secret came from, for doctor diagnostics.
	SecretSource string
}

// Overrides carries command-line flag values. Empty strings mean "not set" and
// fall through to the next source.
type Overrides struct {
	Endpoint        string
	Bucket          string
	PublicBase      string
	KeyTemplate     string
	AccessKeyID     string
	SecretAccessKey string
}

// firstNonEmpty returns the first non-empty argument.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Resolve merges, in descending precedence, command-line flags, environment
// variables, the config file profile, and finally the OS keyring (for the
// secret only). The profile may be nil when the file has no such entry, which
// is valid as long as flags/env supply the required fields.
func Resolve(profile string, p *Profile, ov Overrides) (*Resolved, error) {
	if p == nil {
		p = &Profile{}
	}

	r := &Resolved{
		Profile:     profile,
		Backend:     firstNonEmpty(os.Getenv("IMGUP_BACKEND"), p.Backend, DefaultBackend),
		Endpoint:    firstNonEmpty(ov.Endpoint, os.Getenv("IMGUP_ENDPOINT"), p.Endpoint),
		Bucket:      firstNonEmpty(ov.Bucket, os.Getenv("IMGUP_BUCKET"), p.Bucket),
		PublicBase:  firstNonEmpty(ov.PublicBase, os.Getenv("IMGUP_PUBLIC_BASE"), p.PublicBase),
		KeyTemplate: firstNonEmpty(ov.KeyTemplate, os.Getenv("IMGUP_KEY_TEMPLATE"), p.KeyTemplate, DefaultKeyTemplate),
		AccessKeyID: firstNonEmpty(ov.AccessKeyID, os.Getenv("IMGUP_ACCESS_KEY_ID"), p.AccessKeyID),
	}

	// Secret precedence: flag > env > inline file > keychain.
	switch {
	case ov.SecretAccessKey != "":
		r.SecretAccessKey, r.SecretSource = ov.SecretAccessKey, "flag"
	case os.Getenv("IMGUP_SECRET_ACCESS_KEY") != "":
		r.SecretAccessKey, r.SecretSource = os.Getenv("IMGUP_SECRET_ACCESS_KEY"), "env"
	case p.SecretAccessKey != "":
		r.SecretAccessKey, r.SecretSource = p.SecretAccessKey, "config"
	case p.UseKeychain:
		s, err := GetSecret(profile)
		if err != nil {
			return nil, fmt.Errorf("read secret for profile %q from keychain: %w", profile, err)
		}
		r.SecretAccessKey, r.SecretSource = s, "keychain"
	}

	return r, nil
}

// Validate reports the first missing required field, if any.
func (r *Resolved) Validate() error {
	missing := func(name, val string) error {
		if val == "" {
			return fmt.Errorf("missing required setting: %s (set it via flag, env, or `imgup init`)", name)
		}
		return nil
	}
	for _, check := range []struct{ name, val string }{
		{"endpoint", r.Endpoint},
		{"bucket", r.Bucket},
		{"public_base", r.PublicBase},
		{"access_key_id", r.AccessKeyID},
		{"secret_access_key", r.SecretAccessKey},
	} {
		if err := missing(check.name, check.val); err != nil {
			return err
		}
	}
	return nil
}
