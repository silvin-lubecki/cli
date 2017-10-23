package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/commands"
	"github.com/docker/cli/cli/debug"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestClientDebugEnabled(t *testing.T) {
	defer debug.Disable()

	cmd := commands.NewDockerCommand(&command.DockerCli{}, commands.AddCommands)
	cmd.Flags().Set("debug", "true")

	err := cmd.PersistentPreRunE(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "1", os.Getenv("DEBUG"))
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestExitStatusForInvalidSubcommandWithHelpFlag(t *testing.T) {
	discard := ioutil.Discard
	cmd := commands.NewDockerCommand(command.NewDockerCli(os.Stdin, discard, discard), commands.AddCommands)
	cmd.SetArgs([]string{"help", "invalid"})
	err := cmd.Execute()
	assert.EqualError(t, err, "unknown help topic: invalid")
}
