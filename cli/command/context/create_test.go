package context

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/kubernetes"
	"github.com/docker/cli/cli/context/store"
	"github.com/docker/cli/internal/test"
	"gotest.tools/assert"
	"gotest.tools/env"
)

func makeFakeCli(t *testing.T, opts ...func(*test.FakeCli)) (*test.FakeCli, func()) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NilError(t, err)
	storeConfig := store.NewConfig(
		func() interface{} { return &command.DockerContext{} },
		store.EndpointTypeGetter(docker.DockerEndpoint, func() interface{} { return &docker.EndpointMeta{} }),
		store.EndpointTypeGetter(kubernetes.KubernetesEndpoint, func() interface{} { return &kubernetes.EndpointMeta{} }),
	)
	store := store.New(dir, storeConfig)
	cleanup := func() {
		os.RemoveAll(dir)
	}
	result := test.NewFakeCli(nil, opts...)
	for _, o := range opts {
		o(result)
	}
	result.SetContextStore(store)
	return result, cleanup
}

func withCliConfig(configFile *configfile.ConfigFile) func(*test.FakeCli) {
	return func(m *test.FakeCli) {
		m.SetConfigFile(configFile)
	}
}

func withPipeInOut(close <-chan struct{}) func(*test.FakeCli) {
	return func(m *test.FakeCli) {
		pipeReader, pipeWriter := io.Pipe()
		inStream := command.NewInStream(pipeReader)
		outStream := command.NewOutStream(pipeWriter)
		m.SetIn(inStream)
		m.SetOut(outStream)
		go func() {
			<-close
			pipeWriter.Close()
		}()
	}
}

func TestCreateNoName(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	err := runCreate(cli, &createOptions{})
	assert.ErrorContains(t, err, `context name cannot be empty`)
}

func TestCreateExitingContext(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	assert.NilError(t, cli.ContextStore().CreateOrUpdateContext(store.ContextMetadata{Name: "test"}))

	err := runCreate(cli, &createOptions{
		name: "test",
	})
	assert.ErrorContains(t, err, `context "test" already exists`)
}

func TestCreateInvalidOrchestrator(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()

	err := runCreate(cli, &createOptions{
		name:                     "test",
		defaultStackOrchestrator: "invalid",
	})
	assert.ErrorContains(t, err, `specified orchestrator "invalid" is invalid, please use either kubernetes, swarm or all`)
}

func TestCreateOrchestratorSwarm(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()

	err := runCreate(cli, &createOptions{
		name:                     "test",
		defaultStackOrchestrator: "swarm",
		docker:                   map[string]string{},
	})
	assert.NilError(t, err)
}

func TestCreateOrchestratorEmpty(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()

	err := runCreate(cli, &createOptions{
		name:   "test",
		docker: map[string]string{},
	})
	assert.NilError(t, err)
}

func TestCreateOrchestratorKubernetesNoEndpoint(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()

	err := runCreate(cli, &createOptions{
		name:                     "test",
		defaultStackOrchestrator: "kubernetes",
		docker:                   map[string]string{},
	})
	assert.ErrorContains(t, err, `cannot specify orchestrator "kubernetes" without configuring a Kubernetes endpoint`)
}

func TestCreateOrchestratorAllNoKubernetesEndpoint(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()

	err := runCreate(cli, &createOptions{
		name:                     "test",
		defaultStackOrchestrator: "all",
		docker:                   map[string]string{},
	})
	assert.ErrorContains(t, err, `cannot specify orchestrator "all" without configuring a Kubernetes endpoint`)
}

func validateTestKubeEndpoint(t *testing.T, s store.Store, name string) {
	ctxMetadata, err := s.GetContextMetadata(name)
	assert.NilError(t, err)
	kubeMeta := ctxMetadata.Endpoints[kubernetes.KubernetesEndpoint].(kubernetes.EndpointMeta)
	kubeEP, err := kubeMeta.WithTLSData(s, name)
	assert.NilError(t, err)
	assert.Equal(t, "https://someserver", kubeEP.Host)
	assert.Equal(t, "the-ca", string(kubeEP.TLSData.CA))
	assert.Equal(t, "the-cert", string(kubeEP.TLSData.Cert))
	assert.Equal(t, "the-key", string(kubeEP.TLSData.Key))
}

func createTestContextWithKube(t *testing.T, cli command.Cli) {
	revert := env.Patch(t, "KUBECONFIG", "./testdata/test-kubeconfig")
	defer revert()

	err := runCreate(cli, &createOptions{
		name:                     "test",
		defaultStackOrchestrator: "all",
		kubernetes: map[string]string{
			keyFromCurrent: "true",
		},
		docker: map[string]string{},
	})
	assert.NilError(t, err)
}

func TestCreateOrchestratorAllKubernetesEndpointFromCurrent(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKube(t, cli)
	validateTestKubeEndpoint(t, cli.ContextStore(), "test")
}

func TestCreateInvalidDockerHost(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	err := runCreate(cli, &createOptions{
		name: "test",
		docker: map[string]string{
			keyHost: "some///invalid/host",
		},
	})
	assert.ErrorContains(t, err, "unable to parse docker host")
}
