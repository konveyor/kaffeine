package list

import (
	"fmt"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists the current installed catalog of functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			b, err := functionManager.GenerateInstalledCatalog()
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println(string(b))
			return nil
		},
	}

	return cmd
}
