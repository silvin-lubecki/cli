package swarm

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

// GetStacks lists the swarm stacks.
func GetStacks(dockerCli command.Cli) ([]*formatter.Stack, []error, error) {
	services, err := dockerCli.Client().ServiceList(
		context.Background(),
		types.ServiceListOptions{Filters: getAllStacksFilter()})
	if err != nil {
		return nil, nil, err
	}
	m := make(map[string]*formatter.Stack)
	var errs []error
	for _, service := range services {
		labels := service.Spec.Labels
		name, ok := labels[convert.LabelNamespace]
		if !ok {
			errs = append(errs, errors.Errorf("cannot get label %s for service %s", convert.LabelNamespace, service.ID))
			continue
		}
		ztack, ok := m[name]
		if !ok {
			m[name] = &formatter.Stack{
				Name:         name,
				Services:     1,
				Orchestrator: "Swarm",
			}
		} else {
			ztack.Services++
		}
	}
	var stacks []*formatter.Stack
	for _, stack := range m {
		stacks = append(stacks, stack)
	}
	return stacks, errs, nil
}
