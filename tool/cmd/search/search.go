package search

import (
	"fmt"
	"kaffine-mod/kaffine"

	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [name]",
		Short: "Searches the managed catalogs for a function with the specified name",
		RunE: func(cmd *cobra.Command, args []string) error {
			fname := args[len(args)-1]
			res, err := kaffine.Fm.SearchFunctionDefintions(fname)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}

	return cmd
}
