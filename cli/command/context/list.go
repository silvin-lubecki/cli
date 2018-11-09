package context

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/context/docker"
	kubecontext "github.com/docker/cli/cli/context/kubernetes"
	"github.com/docker/cli/kubernetes"
	"github.com/spf13/cobra"
)

type listOptions struct {
	format string
	quiet  bool
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List contexts",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.format, "format", formatter.TableFormatKey, "Pretty-print contexts using a Go template")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Only show context names")
	return cmd
}

func runList(dockerCli command.Cli, opts *listOptions) error {
	curContext := dockerCli.CurrentContext()
	contextMap, err := dockerCli.ContextStore().ListContexts()
	if err != nil {
		return err
	}
	var contexts []*formatter.ClientContext
	for _, rawMeta := range contextMap {
		meta, err := command.GetDockerContext(rawMeta)
		if err != nil {
			return err
		}
		dockerEndpoint, err := docker.EndpointFromContext(rawMeta)
		if err != nil {
			return err
		}
		kubernetesEndpoint := kubecontext.EndpointFromContext(rawMeta)
		kubEndpointText := ""
		if kubernetesEndpoint != nil {
			kubEndpointText = fmt.Sprintf("%s (%s)", kubernetesEndpoint.Host, kubernetesEndpoint.DefaultNamespace)
		}
		desc := formatter.ClientContext{
			Name:               rawMeta.Name,
			Current:            rawMeta.Name == curContext,
			Description:        meta.Description,
			StackOrchestrator:  hideUnset(meta.StackOrchestrator),
			DockerEndpoint:     dockerEndpoint.Host,
			KubernetesEndpoint: kubEndpointText,
		}
		contexts = append(contexts, &desc)
	}
	if dockerCli.CurrentContext() == "" && !opts.quiet {
		orchestrator, _ := dockerCli.StackOrchestrator("")
		kubEndpointText := ""
		kubeconfig := kubernetes.NewKubernetesConfig("")
		if cfg, err := kubeconfig.ClientConfig(); err == nil {
			ns, _, _ := kubeconfig.Namespace()
			if ns == "" {
				ns = "default"
			}
			kubEndpointText = fmt.Sprintf("%s (%s)", cfg.Host, ns)
		}
		// prepend a "virtual context"
		desc := &formatter.ClientContext{
			Current:            true,
			Description:        "Current DOCKER_HOST based configuration",
			StackOrchestrator:  hideUnset(orchestrator),
			DockerEndpoint:     dockerCli.DockerEndpoint().Host,
			KubernetesEndpoint: kubEndpointText,
		}
		contexts = append([]*formatter.ClientContext{desc}, contexts...)
	}
	return format(dockerCli, opts, contexts)
}

func format(dockerCli command.Cli, opts *listOptions, contexts []*formatter.ClientContext) error {
	contextCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewClientContextFormat(opts.format, opts.quiet),
	}
	return formatter.ClientContextWrite(contextCtx, contexts)
}

func hideUnset(src command.Orchestrator) string {
	res := string(src)
	if res == "unset" {
		return ""
	}
	return res
}
