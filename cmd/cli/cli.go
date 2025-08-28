package main

import (
	"os"
	"time"

	"github.com/ai-model-match/backend/cmd/cli/commands"
	"github.com/ai-model-match/backend/internal/app/auth"
	"github.com/ai-model-match/backend/internal/app/flow"
	"github.com/ai-model-match/backend/internal/app/flowStatistics"
	"github.com/ai-model-match/backend/internal/app/flowStep"
	"github.com/ai-model-match/backend/internal/app/flowStepStatistics"
	"github.com/ai-model-match/backend/internal/app/healthCheck"
	"github.com/ai-model-match/backend/internal/app/useCase"
	"github.com/ai-model-match/backend/internal/app/useCaseStep"
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

/*
This is the entrypoint for the CLI where it is defined the list of
available commands a developer can execute.
Please check in the ´commands´ folder all the available commands.

To execute a command from the main directory of the project
you can run ´go run ./cmd/cli/cli.go <command-name>´
E.g. ´go run ./cmd/cli/cli.go event-replay´
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
	// DB Connection
	dbConnection := mm_db.NewDatabaseConnection(
		envs.DbHost,
		envs.DbUsername,
		envs.DbPassword,
		envs.DbName,
		envs.DbPort,
		envs.DbSslMode,
		envs.DbLogSlowQueryThreshold,
		envs.AppMode,
	)
	// PUB-SUB agent
	pubSubAgent := mm_pubsub.NewPubSubAgent()

	// Scheduler (only declaration, not start)
	scheduler := mm_scheduler.NewScheduler()

	// Init modules
	r := gin.Default()
	v1Api := r.Group("cli")
	healthCheck.Init(envs, dbConnection, v1Api)
	auth.Init(envs, dbConnection, scheduler, v1Api)
	useCase.Init(envs, dbConnection, pubSubAgent, v1Api)
	useCaseStep.Init(envs, dbConnection, pubSubAgent, v1Api)
	flow.Init(envs, dbConnection, pubSubAgent, v1Api)
	flowStep.Init(envs, dbConnection, pubSubAgent, v1Api)
	flowStatistics.Init(envs, dbConnection, pubSubAgent, v1Api)
	flowStepStatistics.Init(envs, dbConnection, pubSubAgent, v1Api)

	// Create CLI app
	app := cli.NewApp()
	app.Name = "Backend"
	app.Usage = "CLI"

	// Define list of commands available in the CLI
	app.Commands = []cli.Command{
		{
			Name: "event-replay",
			Action: func(c *cli.Context) error {
				return commands.EventReplayCommand(c, pubSubAgent, dbConnection)
			},
			Usage: "Replay historical events optionally filtered by topic and start date",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "start-from",
					Usage:    "Optional ISO 8601 date to start replay from",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "topic-name",
					Usage:    "Optional topic name to filter events",
					Required: false,
				},
			},
		},
	}
	// Start the CLI
	err := app.Run(os.Args)
	if err != nil {
		zap.L().Error("Something went wrong during execution", zap.String("service", "cli"), zap.Error(err))
	}
	// Ensure there is enough time before shutting down the CLI
	// to allow all goroutines to be executed
	zap.L().Info("Shutdown CLI in 3 seconds...", zap.String("service", "webapp"))
	time.Sleep(3 * time.Second)
}
