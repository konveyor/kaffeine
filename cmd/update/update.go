package update

import (
	"fmt"
	"os"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates all functions to their latest versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			_, errs := functionManager.CatMan.UpdateAllCatalogs()
			for _, err := range errs {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
			}

			_, errs = functionManager.UpdateAllFunctionDefinitions()
			for _, err := range errs {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					// return err
				}
			}

			err := functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println("Successfully updated catalogs and functions")
			return nil
		},
	}

	return cmd
}
