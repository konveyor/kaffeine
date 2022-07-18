package remove

import (
	"fmt"

	"kaffine-mod/kaffine"

	"github.com/spf13/cobra"
)

func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Searches the managed catalogs for a function with the specified name, and installs it",
		RunE: func(cmd *cobra.Command, args []string) error {
			fname := args[len(args)-1]
			krmFunc, err := kaffine.Fm.RemoveFunctionDefinition(fname)
			if err != nil {
				return err
			}

			fmt.Println("Successfully removed KRM Function '" + krmFunc.GroupName() + "'")

			return nil
		},
	}

	return cmd
}
