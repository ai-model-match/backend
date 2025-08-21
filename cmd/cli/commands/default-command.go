package commands

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

/*
DefaultCommand is just an example of a command you can build and integtrate withint the CLI.
*/
func DefaultCommand(c *cli.Context) error {
	if !c.IsSet("user-id") || c.String("user-id") == "" {
		return errors.New("user-id cannot be empty")
	}
	userID := c.String("user-id")
	zap.L().Info(fmt.Sprintf("Dry run command for user with ID: %s", userID), zap.String("service", "cli-default-command"))
	return nil
}
