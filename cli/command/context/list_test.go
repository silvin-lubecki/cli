package context

import (
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/context/docker"
	"gotest.tools/assert"
	"gotest.tools/env"
	"gotest.tools/golden"
)

func createTestContextWithKubeAndSwarm(t *testing.T, cli command.Cli, name string, orchestrator string) {
	revert := env.Patch(t, "KUBECONFIG", "./testdata/test-kubeconfig")
	defer revert()

	err := runCreate(cli, &createOptions{
		name:                     name,
		defaultStackOrchestrator: orchestrator,
		description:              "description of " + name,
		kubernetes:               map[string]string{keyFromCurrent: "true"},
		docker:                   map[string]string{keyHost: "https://someswarmserver"},
	})
	assert.NilError(t, err)
}

func TestList(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKubeAndSwarm(t, cli, "current", "all")
	createTestContextWithKubeAndSwarm(t, cli, "other", "all")
	cli.SetCurrentContext("current")
	cli.OutBuffer().Reset()
	assert.NilError(t, runList(cli, &listOptions{
		format: formatter.TableFormatKey,
	}))
	golden.Assert(t, cli.OutBuffer().String(), "list.golden")
}

func TestListUnset(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKubeAndSwarm(t, cli, "current", "unset")
	createTestContextWithKubeAndSwarm(t, cli, "other", "all")
	cli.SetCurrentContext("current")
	cli.OutBuffer().Reset()
	assert.NilError(t, runList(cli, &listOptions{
		format: formatter.TableFormatKey,
	}))
	golden.Assert(t, cli.OutBuffer().String(), "list.unset.golden")
}

func TestListNoContext(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	defer env.Patch(t, "KUBECONFIG", "./testdata/test-kubeconfig")()
	cli.SetDockerEndpoint(docker.Endpoint{
		EndpointMeta: docker.EndpointMeta{
			Host: "https://someswarmserver",
		},
	})
	cli.OutBuffer().Reset()
	assert.NilError(t, runList(cli, &listOptions{
		format: formatter.TableFormatKey,
	}))
	golden.Assert(t, cli.OutBuffer().String(), "list.no-context.golden")
}

func TestListQuiet(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKubeAndSwarm(t, cli, "current", "all")
	createTestContextWithKubeAndSwarm(t, cli, "other", "all")
	cli.SetCurrentContext("current")
	cli.OutBuffer().Reset()
	assert.NilError(t, runList(cli, &listOptions{
		format: formatter.TableFormatKey,
		quiet:  true,
	}))
	golden.Assert(t, cli.OutBuffer().String(), "quiet-list.golden")
}

func TestInspect(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKubeAndSwarm(t, cli, "current", "all")
	cli.OutBuffer().Reset()
	assert.NilError(t, runInspect(cli, inspectOptions{
		refs: []string{"current"},
	}))
	golden.Assert(t, cli.OutBuffer().String(), "inspect.golden")
}
