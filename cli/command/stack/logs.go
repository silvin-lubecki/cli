package stack

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/kubernetes"
	"github.com/spf13/cobra"
)

var noColor bool

func newLogsCommand(dockerCli command.Cli) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "logs STACK",
		Short:   "Stream logs on an existing stack",
		Args:    cobra.MinimumNArgs(1),
		Annotations: map[string]string {"kubernetes": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dockerCli.ClientInfo().HasKubernetes {
				kli, err := kubernetes.WrapCli(dockerCli, cmd)
				if err != nil {
					return err
				}
				return kubernetes.RunLogs(kli, args, !noColor)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable output coloring")
	return cmd
}
