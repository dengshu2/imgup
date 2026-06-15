package cmd

import (
	"fmt"
	"os"

	"github.com/dengshu2/imgup/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate the active profile against R2 with a real request",
	Long: `doctor resolves the active profile (flags > env > config file > keychain),
prints where each value came from, and performs an authenticated round-trip to
R2 to confirm the credentials and bucket are usable.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, err := configPath()
		if err != nil {
			return err
		}

		r, err := loadResolved()
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Config file: %s\n", path)
		fmt.Fprintf(os.Stderr, "Profile:     %s\n", r.Profile)
		fmt.Fprintf(os.Stderr, "Backend:     %s\n", r.Backend)
		fmt.Fprintf(os.Stderr, "Endpoint:    %s\n", display(r.Endpoint))
		fmt.Fprintf(os.Stderr, "Bucket:      %s\n", display(r.Bucket))
		fmt.Fprintf(os.Stderr, "Public base: %s\n", display(r.PublicBase))
		fmt.Fprintf(os.Stderr, "Key tmpl:    %s\n", r.KeyTemplate)
		fmt.Fprintf(os.Stderr, "Access key:  %s\n", display(r.AccessKeyID))
		fmt.Fprintf(os.Stderr, "Secret:      %s (from %s)\n", masked(r.SecretAccessKey), orNone(r.SecretSource))

		warnInsecureInlineSecret(path, r)

		if err := r.Validate(); err != nil {
			return err
		}

		up, err := newUploader(cmd.Context(), r)
		if err != nil {
			return err
		}

		fmt.Fprint(os.Stderr, "\nChecking R2 access... ")
		if err := up.Check(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, "FAILED")
			return err
		}
		fmt.Fprintln(os.Stderr, "OK")
		return nil
	},
}

// warnInsecureInlineSecret nudges the user when a plaintext secret sits in a
// config file whose permissions are looser than 0600.
func warnInsecureInlineSecret(path string, r *config.Resolved) {
	if r.SecretSource != "config" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Mode().Perm()&0o077 != 0 {
		fmt.Fprintf(os.Stderr, "WARNING: %s is group/world-accessible (%#o) and holds an inline secret; run: chmod 600 %s\n",
			path, info.Mode().Perm(), path)
	}
}

func display(s string) string {
	if s == "" {
		return "(unset)"
	}
	return s
}

func orNone(s string) string {
	if s == "" {
		return "unset"
	}
	return s
}

// masked shows only enough of a secret to tell two apart, never the whole thing.
func masked(s string) string {
	if s == "" {
		return "(unset)"
	}
	if len(s) <= 6 {
		return "******"
	}
	return s[:3] + "…" + s[len(s)-3:]
}
