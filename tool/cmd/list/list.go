package list

import (
	"fmt"

	"kaffine-mod/kaffine"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists the current installed catalog of functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := kaffine.Fm.GenerateInstalledCatalog()
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}

	return cmd
}
