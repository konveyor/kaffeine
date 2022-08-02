package install

import (
	"fmt"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
)

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Searches the managed catalogs for a function with the specified name, and installs it",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			fname := args[len(args)-1]
			fn, err := functionManager.AddFunctionDefinition(fname)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println("Successfully added KRM Function '" + fn.GroupName() + "'\n")
			return nil
		},
	}

	return cmd
}
