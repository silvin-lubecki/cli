package context

import (
	"bytes"
	"fmt"
	"github.com/docker/cli/cli/command"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestExportImportWithFile(t *testing.T) {
	contextDir, err := ioutil.TempDir("", t.Name()+"context")
	assert.NilError(t, err)
	defer os.RemoveAll(contextDir)
	contextFile := filepath.Join(contextDir, "exported")
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKube(t, cli)
	cli.ErrBuffer().Reset()
	assert.NilError(t, runExport(cli, &exportOptions{
		contextName: "test",
		dest:        contextFile,
	}))
	assert.Equal(t, cli.ErrBuffer().String(), fmt.Sprintf("Written file %q\n", contextFile))
	assert.NilError(t, runImport(cli, "test2", contextFile))
	context1, err := cli.ContextStore().GetContextMetadata("test")
	assert.NilError(t, err)
	context2, err := cli.ContextStore().GetContextMetadata("test2")
	assert.NilError(t, err)
	assert.DeepEqual(t, context1.Endpoints, context2.Endpoints)
	assert.DeepEqual(t, context1.Metadata, context2.Metadata)
	assert.Equal(t, "test", context1.Name)
	assert.Equal(t, "test2", context2.Name)
}

func TestExportImportPipe(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKube(t, cli)
	cli.ErrBuffer().Reset()
	cli.OutBuffer().Reset()
	assert.NilError(t, runExport(cli, &exportOptions{
		contextName: "test",
		dest:        "-",
	}))
	assert.Equal(t, cli.ErrBuffer().String(), "")
	cli.SetIn(command.NewInStream(ioutil.NopCloser(bytes.NewBuffer(cli.OutBuffer().Bytes()))))
	assert.NilError(t, runImport(cli, "test2", "-"))
	context1, err := cli.ContextStore().GetContextMetadata("test")
	assert.NilError(t, err)
	context2, err := cli.ContextStore().GetContextMetadata("test2")
	assert.NilError(t, err)
	assert.DeepEqual(t, context1.Endpoints, context2.Endpoints)
	assert.DeepEqual(t, context1.Metadata, context2.Metadata)
	assert.Equal(t, "test", context1.Name)
	assert.Equal(t, "test2", context2.Name)
}

func TestExportKubeconfig(t *testing.T) {
	contextDir, err := ioutil.TempDir("", t.Name()+"context")
	assert.NilError(t, err)
	defer os.RemoveAll(contextDir)
	contextFile := filepath.Join(contextDir, "exported")
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContextWithKube(t, cli)
	cli.ErrBuffer().Reset()
	assert.NilError(t, runExport(cli, &exportOptions{
		contextName: "test",
		dest:        contextFile,
		kubeconfig:  true,
	}))
	assert.Equal(t, cli.ErrBuffer().String(), fmt.Sprintf("Written file %q\n", contextFile))
	assert.NilError(t, runCreate(cli, &createOptions{
		name: "test2",
		kubernetes: map[string]string{
			keyKubeconfig: contextFile,
		},
		docker: map[string]string{},
	}))
	validateTestKubeEndpoint(t, cli.ContextStore(), "test2")
}
