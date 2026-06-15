package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagLsLong  bool
	flagLsURL   bool
	flagLsLimit int
)

func init() {
	lsCmd.Flags().BoolVarP(&flagLsLong, "long", "l", false, "long format: key, size, last-modified (tab-separated)")
	lsCmd.Flags().BoolVar(&flagLsURL, "url", false, "print public URLs instead of keys")
	lsCmd.Flags().IntVar(&flagLsLimit, "limit", 0, "max objects to list (0 = no limit)")
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:           "ls [prefix]",
	Short:         "List objects in the bucket, optionally filtered by key prefix",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, up, err := resolvedUploader(cmd.Context())
		if err != nil {
			return err
		}
		var prefix string
		if len(args) == 1 {
			prefix = args[0]
		}
		objs, err := up.List(cmd.Context(), prefix, flagLsLimit)
		if err != nil {
			return err
		}
		for _, o := range objs {
			switch {
			case flagLsLong:
				fmt.Printf("%s\t%d\t%s\n", o.Key, o.Size, o.LastModified.Format("2006-01-02 15:04:05"))
			case flagLsURL:
				fmt.Println(joinURL(r.PublicBase, o.Key))
			default:
				fmt.Println(o.Key)
			}
		}
		fmt.Fprintf(os.Stderr, "imgup: %d object(s)\n", len(objs))
		return nil
	},
}
