package kubernetes

import (
	"fmt"
	"testing"

	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/kubernetes/compose/v1beta2"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Fake implementations
type fakeStackLister struct {
	stacks []stack
	errs   []error
	err    error
}

func (f *fakeStackLister) List(opts metav1.ListOptions) ([]stack, []error, error) {
	return f.stacks, f.errs, f.err
}

type fakeNamespaceLister struct {
	namespaces []string
	err        error
}

func (f *fakeNamespaceLister) List() ([]string, error) {
	return f.namespaces, f.err
}

type fakeStackClientFactory struct {
	clients map[string]*fakeStackLister
}

func (f *fakeStackClientFactory) Stacks(namespace string, allNamespaces bool) (stackLister, error) {
	if c, ok := f.clients[fmt.Sprintf("%s-%v", namespace, allNamespaces)]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("failed to find stack client factory for namespace '%s'", namespace)
}

func makeStack(name, namespace string, services []string) stack {
	spec := &v1beta2.StackSpec{
		Services: make([]v1beta2.ServiceConfig, len(services)),
	}
	for i, s := range services {
		spec.Services[i].Name = s
	}
	return stack{
		name:      name,
		namespace: namespace,
		spec:      spec,
	}
}

func TestGetStacksReturnsStacksAndErrorsWithAllNamespaces(t *testing.T) {
	stackLister := &fakeStackLister{
		stacks: []stack{
			makeStack("foo", "space", []string{"s1", "s2"}),
			makeStack("bar", "space", []string{"s3"}),
		},
		errs: []error{
			errors.New("invalid stack bar"),
			errors.New("could not parse stack baz"),
		},
	}
	// List all namespaces with a user who has the right to do it
	clientFactory := &fakeStackClientFactory{
		clients: map[string]*fakeStackLister{
			"-true": stackLister,
		},
	}
	opts := options.List{AllNamespaces: true}
	stacks, errs, err := getStacks(clientFactory, nil, opts, nil)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(stacks, 2))
	assert.Assert(t, is.Len(errs, 2))
}

func TestGetStacksMergesStacksAndErrorsFromNamespaces(t *testing.T) {
	validStackLister := &fakeStackLister{
		stacks: []stack{
			makeStack("foo", "valid", []string{"s1", "s2"}),
			makeStack("bar", "valid", []string{"s3"}),
		},
	}
	validWithErrorStackLister := &fakeStackLister{
		stacks: []stack{
			makeStack("foo", "witherrors", []string{"s1", "s2"}),
		},
		errs: []error{
			errors.New("invalid stack bar"),
			errors.New("could not parse stack baz"),
		},
	}
	invalidStackLister := &fakeStackLister{
		err: errors.New("invalid stack lister"),
	}
	clientFactory := &fakeStackClientFactory{
		clients: map[string]*fakeStackLister{
			"valid-false":      validStackLister,
			"witherrors-false": validWithErrorStackLister,
			"invalid-false":    invalidStackLister,
		},
	}
	opts := options.List{
		Namespaces: []string{"valid", "witherrors", "invalid"},
	}
	stacks, errs, err := getStacks(clientFactory, nil, opts, nil)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(stacks, 3))
	assert.Assert(t, is.Len(errs, 3))
}
