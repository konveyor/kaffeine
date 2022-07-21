package main

import (
	"log"

	"github.com/konveyor/kaffeine/cmd/config"
	"github.com/konveyor/kaffeine/cmd/install"
	"github.com/konveyor/kaffeine/cmd/list"
	"github.com/konveyor/kaffeine/cmd/remove"
	"github.com/konveyor/kaffeine/cmd/search"
	"github.com/konveyor/kaffeine/cmd/update"
	"github.com/konveyor/kaffeine/cmd/version"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kaffeine",
		Short: "kaffeine is a KRM Function Manager",
	}

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())
	rootCmd.AddCommand(list.NewListCommand())
	rootCmd.AddCommand(search.NewSearchCommand())
	rootCmd.AddCommand(install.NewInstallCommand())
	rootCmd.AddCommand(remove.NewRemoveCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffeine encountered an error.\n%v\n", rootErr)
	}
}
