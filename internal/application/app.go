package application

import (
	"sync-backend/pkg/console"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var cmds = map[string]console.Command{
	"app:serve": NewServeCommand(),
}

func GetSubCommands(opt fx.Option) []*cobra.Command {
	subCommands := make([]*cobra.Command, 0)
	for name, cmd := range cmds {
		subCommands = append(subCommands, console.WrapSubCommand(name, cmd, opt))
	}
	return subCommands
}

var rootCmd = &cobra.Command{
	Use:              "sync-backend",
	Short:            "Backend service for Sync",
	Long:             "Sync backend is a backend service for the social media platform - Sync. It is a backend service that provides APIs for the mobile application.",
	TraverseChildren: true,
}

// App root of the application
type App struct {
	*cobra.Command
}

// NewApp creates new root command
func NewApp() App {
	cmd := App{
		Command: rootCmd,
	}
	cmd.AddCommand(GetSubCommands(CommonModules)...)
	return cmd
}

var RootApp = NewApp()
