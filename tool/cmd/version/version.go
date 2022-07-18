package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Kaffine",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Kaffine version 0.0.0")

			return nil
		},
	}

	return cmd
}
