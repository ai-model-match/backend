package main

import (
	"os"

	"github.com/ai-model-match/backend/cmd/cli/commands"
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

/*
This is the entrypoint for the CLI where it is defined the list of
available commands a developer can execute.
Please check in the ´commands´ folder all the available commands.

To execute a command from the main directory of the project
you can run ´go run ./cmd/cli/cli.go <command-name>´
E.g. ´go run ./cmd/cli/cli.go default-command´
*/
func main() {
	// Set default Timezone
	os.Setenv("TZ", "UTC")
	// ENV Variables
	envs := mm_env.ReadEnvs()
	// Set Logger
	logger := zap.Must(zap.NewProduction())
	if envs.AppMode != "release" {
		logger = zap.Must(zap.NewDevelopment())
	}
	zap.ReplaceGlobals(logger)
	// Start CLI
	app := cli.NewApp()
	app.Name = "Backend"
	app.Usage = "CLI"

	// Define list of commands available in the CLI
	app.Commands = []cli.Command{
		{
			Name:   "default-command",
			Action: commands.DefaultCommand,
			Usage:  "Call a defaul command for the purpose of see how it works",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "user-id",
					Usage: "The ID of the user",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		zap.L().Error("Something went wrong during execution", zap.String("service", "cli"), zap.Error(err))
	}
}
