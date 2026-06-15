package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dengshu2/imgup/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create or update a config profile interactively",
	Long: `init prompts for the R2 connection settings and writes them to the config
file (default ~/.config/imgup/config.toml, mode 0600).

On macOS the secret access key is offered to the system Keychain by default, so
it is never written to disk. Choose inline storage for portability (Linux/CI).`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runInit,
}

func runInit(cmd *cobra.Command, _ []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	f, err := config.Load(path)
	if err != nil {
		return err
	}

	in := bufio.NewReader(os.Stdin)
	name := flagProfile
	if name == "" {
		name = config.DefaultProfileName
	}

	// Pre-fill from an existing profile so re-running init edits in place.
	prev := f.Profiles[name]
	if prev == nil {
		prev = &config.Profile{Backend: config.DefaultBackend, KeyTemplate: config.DefaultKeyTemplate}
	}

	fmt.Fprintf(os.Stderr, "Configuring profile %q in %s\n\n", name, path)

	p := &config.Profile{Backend: config.DefaultBackend}
	p.Endpoint = ask(in, "R2 S3 endpoint", prev.Endpoint)
	p.Bucket = ask(in, "Bucket name", prev.Bucket)
	p.PublicBase = ask(in, "Public base URL (custom domain)", prev.PublicBase)
	p.KeyTemplate = ask(in, "Key template", orDefault(prev.KeyTemplate, config.DefaultKeyTemplate))
	p.AccessKeyID = ask(in, "Access key ID", prev.AccessKeyID)
	secret := askSecret(in, "Secret access key (leave blank to keep existing)")

	// Decide where the secret lives.
	useKeychain := runtime.GOOS == "darwin"
	if secret != "" || !prev.UseKeychain {
		useKeychain = askYesNo(in, "Store secret in OS keychain (recommended)?", runtime.GOOS == "darwin")
	} else {
		useKeychain = prev.UseKeychain
	}

	if useKeychain {
		if secret != "" {
			if err := config.SetSecret(name, secret); err != nil {
				return fmt.Errorf("store secret in keychain: %w", err)
			}
			fmt.Fprintln(os.Stderr, "Secret saved to keychain.")
		}
		p.UseKeychain = true
	} else {
		if secret != "" {
			p.SecretAccessKey = secret
		} else {
			p.SecretAccessKey = prev.SecretAccessKey
		}
		fmt.Fprintln(os.Stderr, "WARNING: secret stored inline in the config file (mode 0600).")
	}

	f.Profiles[name] = p
	if f.DefaultProfile == "" {
		f.DefaultProfile = name
	}
	if err := config.Save(path, f); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\nSaved. Verify with: imgup doctor%s\n", profileFlagSuffix(name))
	return nil
}

// ask prompts for a value, returning def when the user enters nothing.
func ask(in *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", label)
	}
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// askSecret prompts without echoing a default value.
func askSecret(in *bufio.Reader, label string) string {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	line, _ := in.ReadString('\n')
	return strings.TrimSpace(line)
}

// askYesNo prompts for a yes/no answer with a default.
func askYesNo(in *bufio.Reader, label string, def bool) bool {
	hint := "y/N"
	if def {
		hint = "Y/n"
	}
	fmt.Fprintf(os.Stderr, "%s [%s]: ", label, hint)
	line, _ := in.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return def
	}
	return line == "y" || line == "yes"
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func profileFlagSuffix(name string) string {
	if name == config.DefaultProfileName {
		return ""
	}
	return " --profile " + name
}
