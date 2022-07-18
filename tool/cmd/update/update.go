package update

import (
	"fmt"
	"kaffine-mod/kaffine"
	"os"

	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates all functions to their latest versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, errs := kaffine.Fm.CatMan.UpdateAllCatalogs()
			for _, err := range errs {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
			}

			_, errs = kaffine.Fm.UpdateAllFunctionDefinitions()
			for _, err := range errs {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					// return err
				}
			}

			fmt.Println("Successfully updated catalogs and functions")

			return nil
		},
	}

	return cmd
}
