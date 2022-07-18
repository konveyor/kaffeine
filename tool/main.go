package main

import (
	"kaffine-mod/cmd/config"
	"kaffine-mod/cmd/install"
	"kaffine-mod/cmd/list"
	"kaffine-mod/cmd/remove"
	"kaffine-mod/cmd/search"
	"kaffine-mod/cmd/update"
	"kaffine-mod/cmd/version"
	"kaffine-mod/kaffine"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	kaffine.InitializeGlobals()

	var rootCmd = &cobra.Command{
		Use:   "kaffine",
		Short: "Kaffine is a KRM Function Manager",
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
		log.Fatalf("kaffine encountered an error.\n%v\n", rootErr)
	}

	deleteErr := kaffine.DestroyGlobals()
	if deleteErr != nil {
		log.Fatalf("%v", deleteErr)
	}

	// fmt.Println("main:", kaffine.Directory)
	// x, _ := fm.CatMan.Search("Logger@v1.0.1")
	// for _, y := range x {
	// 	data, _ := yaml.Marshal(y)
	// 	fmt.Println(string(data))
	// }

	// err = kaffine.DestroyGlobals()
	// if err != nil {
	// 	panic(err)
	// }
}
