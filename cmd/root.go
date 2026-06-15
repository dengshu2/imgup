package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dengshu2/imgup/internal/config"
	"github.com/dengshu2/imgup/internal/storage"
	"github.com/spf13/cobra"
)

// version is overridable at build time via -ldflags "-X .../cmd.version=...".
var version = "dev"

// Persistent flag values shared across commands.
var (
	flagProfile string
	flagConfig  string

	// Credential/setting overrides (highest precedence after init writes them).
	flagEndpoint    string
	flagBucket      string
	flagPublicBase  string
	flagKeyTemplate string
	flagAccessKeyID string
	flagSecretKey   string
)

var rootCmd = &cobra.Command{
	Use:   "imgup [flags] <file>...",
	Short: "Upload images to Cloudflare R2 and print their public URLs",
	Long: `imgup uploads images to a Cloudflare R2 (S3-compatible) bucket and prints the
resulting public URL to stdout. All diagnostics are written to stderr, so imgup
composes cleanly in scripts and pipelines.

Run "imgup init" to create a config, then "imgup <file>" to upload.`,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
	// RunE handles the default action (upload) when args don't match a
	// subcommand. With no args it prints help.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runUpload(cmd.Context(), args)
	},
}

// Execute is the program entry point.
func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "imgup: "+err.Error())
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagProfile, "profile", "", "config profile to use (default: the file's default_profile)")
	pf.StringVar(&flagConfig, "config", "", "path to config file (default: ~/.config/imgup/config.toml)")
	pf.StringVar(&flagEndpoint, "endpoint", "", "override R2 S3 endpoint")
	pf.StringVar(&flagBucket, "bucket", "", "override bucket name")
	pf.StringVar(&flagPublicBase, "public-base", "", "override public base URL")
	pf.StringVar(&flagKeyTemplate, "key-template", "", "override object key template")
	pf.StringVar(&flagAccessKeyID, "access-key-id", "", "override access key id")
	pf.StringVar(&flagSecretKey, "secret-access-key", "", "override secret access key")
}

// overrides collects the credential/setting flags into a config.Overrides.
func overrides() config.Overrides {
	return config.Overrides{
		Endpoint:        flagEndpoint,
		Bucket:          flagBucket,
		PublicBase:      flagPublicBase,
		KeyTemplate:     flagKeyTemplate,
		AccessKeyID:     flagAccessKeyID,
		SecretAccessKey: flagSecretKey,
	}
}

// loadResolved resolves the active profile through the full precedence chain.
func loadResolved() (*config.Resolved, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	f, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	prof, name, _ := f.Profile(flagProfile)
	return config.Resolve(name, prof, overrides())
}

// configPath returns the effective config path, honoring --config.
func configPath() (string, error) {
	if flagConfig != "" {
		return flagConfig, nil
	}
	return config.Path()
}

// newUploader builds the backend for a resolved profile.
func newUploader(ctx context.Context, r *config.Resolved) (storage.Uploader, error) {
	switch r.Backend {
	case "", config.DefaultBackend:
		return storage.NewR2(ctx, storage.R2Options{
			Endpoint:        r.Endpoint,
			Bucket:          r.Bucket,
			AccessKeyID:     r.AccessKeyID,
			SecretAccessKey: r.SecretAccessKey,
		})
	default:
		return nil, fmt.Errorf("unknown backend %q", r.Backend)
	}
}

// resolvedUploader resolves the active profile, validates it, and builds the
// backend — the shared preamble for upload/ls/rm/exists.
func resolvedUploader(ctx context.Context) (*config.Resolved, storage.Uploader, error) {
	r, err := loadResolved()
	if err != nil {
		return nil, nil, err
	}
	if err := r.Validate(); err != nil {
		return nil, nil, err
	}
	up, err := newUploader(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	return r, up, nil
}
