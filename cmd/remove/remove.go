package remove

import (
	"fmt"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
)

func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Searches the managed catalogs for a function with the specified name, and installs it",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			fname := args[len(args)-1]
			krmFunc, err := functionManager.RemoveFunctionDefinition(fname)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println("Successfully removed KRM Function '" + krmFunc.GroupName() + "'")
			return nil
		},
	}

	return cmd
}
