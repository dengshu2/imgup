package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// SkillAsset holds the embedded companion SKILL.md, injected from main via
// go:embed. It lets `imgup skill install` work with no network or extra files.
var SkillAsset string

// skillName is the folder the skill is installed under in each agent dir.
const skillName = "imgup"

var flagSkillTo []string

func init() {
	skillInstallCmd.Flags().StringArrayVar(&flagSkillTo, "to", nil,
		"target agent skills directory (repeatable; default: auto-detected)")
	skillCmd.AddCommand(skillInstallCmd, skillPrintCmd)
	rootCmd.AddCommand(skillCmd)
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage imgup's bundled agent skill",
	Long: `imgup bundles a companion "skill" — a short instruction file that teaches an
AI coding agent (Claude Code, Cursor, ...) when and how to call imgup. The skill
is embedded in this binary, so "imgup skill install" needs no network or extra
files.`,
}

var skillPrintCmd = &cobra.Command{
	Use:           "print",
	Short:         "Print the bundled SKILL.md to stdout",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), SkillAsset)
		return err
	},
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the bundled skill into agent skills directories",
	Long: `Writes the bundled SKILL.md to <dir>/imgup/SKILL.md for each target agent
skills directory. With no --to flags, imgup installs into every well-known agent
skills directory that exists on this machine.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runSkillInstall,
}

func runSkillInstall(_ *cobra.Command, _ []string) error {
	if SkillAsset == "" {
		return fmt.Errorf("this build has no embedded skill asset")
	}

	targets := flagSkillTo
	if len(targets) == 0 {
		detected, err := defaultSkillDirs()
		if err != nil {
			return err
		}
		if len(detected) == 0 {
			return fmt.Errorf("no known agent skills directory found; pass --to <dir>")
		}
		targets = detected
	}

	for _, base := range targets {
		dir := filepath.Join(base, skillName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		dst := filepath.Join(dir, "SKILL.md")
		// Drop any pre-existing symlink so we write a real, self-contained file.
		_ = os.Remove(dst)
		if err := os.WriteFile(dst, []byte(SkillAsset), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "installed skill -> %s\n", dst)
	}
	return nil
}

// defaultSkillDirs returns the well-known agent skills directories that exist
// on this machine, so a bare `imgup skill install` fans out to all of them.
func defaultSkillDirs() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	candidates := []string{
		filepath.Join(home, ".agents", "skills"), // shared source used by the `skills` tool
		filepath.Join(home, ".claude", "skills"), // Claude Code
	}
	var found []string
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			found = append(found, c)
		}
	}
	return found, nil
}
