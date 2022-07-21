package search

import (
	"fmt"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [name]",
		Short: "Searches the managed catalogs for a function with the specified name",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			fname := args[len(args)-1]
			res, err := functionManager.SearchFunctionDefintions(fname)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}
