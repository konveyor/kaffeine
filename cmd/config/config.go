package config

import (
	"fmt"

	"github.com/konveyor/kaffeine/kaffeine"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// TODO: support both local and global config
// global config determined by KAFFEINE_GLOBAL_CONFIG env variable.
// If unset, defaults to ~/.kaffeine/config
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Edit the configuration of kaffeine.",
	}

	addCatalog := &cobra.Command{
		Use:   "add-catalog [catalog uri]",
		Short: "Adds catalog to list of managed catalogs in kaffeine",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			uri := args[len(args)-1]
			err := functionManager.CatMan.AddCatalogFromUri(uri)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Printf("Successfully added catalog '%s'", uri)
			return nil
		},
	}

	remCatalog := &cobra.Command{
		Use:   "remove-catalog [catalog uri]",
		Short: "Removes catalog to list of managed catalogs in kaffeine",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			uri := args[len(args)-1]
			_, err := functionManager.CatMan.RemoveCatalog(uri)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Printf("Successfully removed catalog '%s'", uri)
			return nil
		},
	}

	listConfig := &cobra.Command{
		Use:   "list",
		Short: "Lists current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			functionManager := kaffeine.NewFunctionManager("")

			functionManager.UpdateConfig()
			data, err := yaml.Marshal(functionManager.Cfg)
			if err != nil {
				return err
			}

			err = functionManager.Save()
			if err != nil {
				return err
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.AddCommand(addCatalog)
	cmd.AddCommand(remCatalog)
	cmd.AddCommand(listConfig)

	return cmd
}
