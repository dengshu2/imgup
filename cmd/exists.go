package cmd

import (
	"fmt"
	"time"

	"github.com/dengshu2/imgup/internal/naming"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(existsCmd)
}

var existsCmd = &cobra.Command{
	Use:   "exists <file>",
	Short: "Check whether a file's content is already in the bucket",
	Long: `exists computes the object key a file would map to (by content md5) and checks
the bucket. If present it prints the public URL and exits 0; otherwise it exits
non-zero. Useful for dedup checks before deciding to upload.`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, up, err := resolvedUploader(cmd.Context())
		if err != nil {
			return err
		}
		content, origName, err := readSource(args[0])
		if err != nil {
			return err
		}
		obj, err := naming.Resolve(r.KeyTemplate, content, origName, time.Now())
		if err != nil {
			return err
		}
		ok, err := up.Exists(cmd.Context(), obj.Key)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("not found: %s", obj.Key)
		}
		fmt.Println(joinURL(r.PublicBase, obj.Key))
		return nil
	},
}
