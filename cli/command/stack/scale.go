package stack

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/kubernetes"
	"github.com/spf13/cobra"
)

func newScaleCommand(dockerCli command.Cli) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "scale STACK service=replicas...",
		Short:   "Scale services on an existing stack",
		Args:    cobra.MinimumNArgs(2),
		Annotations: map[string]string {"kubernetes": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dockerCli.ClientInfo().HasKubernetes() {
				kli, err := kubernetes.WrapCli(dockerCli, cmd)
				if err != nil {
					return err
				}
				return kubernetes.RunScale(kli, args)
			}
			return nil
		},
	}

	return cmd
}
