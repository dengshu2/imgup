package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var flagRmYes bool

func init() {
	rmCmd.Flags().BoolVarP(&flagRmYes, "yes", "y", false, "skip the confirmation prompt")
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:           "rm <key|url>...",
	Short:         "Delete objects by key or public URL",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, up, err := resolvedUploader(cmd.Context())
		if err != nil {
			return err
		}

		keys := make([]string, len(args))
		for i, a := range args {
			keys[i] = keyFromArg(a, r.PublicBase)
		}

		if !flagRmYes && !confirmDelete(keys) {
			fmt.Fprintln(os.Stderr, "imgup: aborted")
			return nil
		}

		var failed int
		for _, k := range keys {
			if err := up.Delete(cmd.Context(), k); err != nil {
				fmt.Fprintf(os.Stderr, "imgup: %s: %v\n", k, err)
				failed++
				continue
			}
			fmt.Fprintf(os.Stderr, "imgup: deleted %s\n", k)
		}
		if failed > 0 {
			return fmt.Errorf("%d of %d delete(s) failed", failed, len(keys))
		}
		return nil
	},
}

// keyFromArg accepts either a raw object key or a full public URL and returns
// the object key.
func keyFromArg(arg, publicBase string) string {
	base := strings.TrimRight(publicBase, "/") + "/"
	if strings.HasPrefix(arg, base) {
		return strings.TrimPrefix(arg, base)
	}
	return strings.TrimLeft(arg, "/")
}

// confirmDelete prints the targets and asks for a y/N confirmation on stdin.
func confirmDelete(keys []string) bool {
	fmt.Fprintf(os.Stderr, "About to delete %d object(s):\n", len(keys))
	for _, k := range keys {
		fmt.Fprintf(os.Stderr, "  %s\n", k)
	}
	fmt.Fprint(os.Stderr, "Proceed? [y/N]: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}
