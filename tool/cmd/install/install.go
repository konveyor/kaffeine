package install

import (
	"fmt"

	"kaffine-mod/kaffine"

	"github.com/spf13/cobra"
)

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Searches the managed catalogs for a function with the specified name, and installs it",
		RunE: func(cmd *cobra.Command, args []string) error {
			fname := args[len(args)-1]
			fn, err := kaffine.Fm.AddFunctionDefinition(fname)
			if err != nil {
				return err
			}

			fmt.Println("Successfully added KRM Function '" + fn.GroupName() + "'")

			return nil
		},
	}

	return cmd
}
