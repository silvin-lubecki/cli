package kubernetes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/cli/cli/command"
	"github.com/pkg/errors"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Interfaces
type stackLister interface {
	List(opts metav1.ListOptions) ([]stack, []error, error)
}

type namespaceLister interface {
	List() ([]string, error)
}

type stackClientFactory interface {
	Stacks(namespace string, allNamespaces bool) (stackLister, error)
}

// Implementations
type stackClientFactoryImpl struct {
	kubeCli KubeCli
}

func (f *stackClientFactoryImpl) Stacks(namespace string, allNamespaces bool) (stackLister, error) {
	if namespace != "" {
		f.kubeCli.kubeNamespace = namespace
	}
	composeClient, err := f.kubeCli.composeClient()
	if err != nil {
		return nil, err
	}
	return composeClient.Stacks(allNamespaces)
}

type compositeStackLister struct {
	listers []stackLister
}

func (c *compositeStackLister) List(opts metav1.ListOptions) ([]stack, []error, error) {
	var (
		stacks  []stack
		allErrs []error
	)
	for _, l := range c.listers {
		ss, errs, err := l.List(opts)
		if err != nil {
			allErrs = append(allErrs, err)
		} else {
			stacks = append(stacks, ss...)
			allErrs = append(allErrs, errs...)
		}
	}
	return stacks, allErrs, nil
}

type userVisibleNamespaceLister struct {
	dockerCli command.Cli
}

func (l *userVisibleNamespaceLister) List() ([]string, error) {
	host := l.dockerCli.Client().DaemonHost()
	endpoint, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	endpoint.Scheme = "https"
	endpoint.Path = "/kubernetesNamespaces"
	resp, err := l.dockerCli.Client().HTTPClient().Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "received %d status and unable to read response", resp.StatusCode)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		nms := &core_v1.NamespaceList{}
		if err := json.Unmarshal(body, nms); err != nil {
			return nil, errors.Wrapf(err, "unmarshal failed: %s", string(body))
		}
		namespaces := make([]string, len(nms.Items))
		for i, namespace := range nms.Items {
			namespaces[i] = namespace.Name
		}
		return namespaces, nil
	case http.StatusNotFound:
		// UCP API not present
		return nil, nil
	default:
		return nil, fmt.Errorf("received %d status while retrieving namespaces: %s", resp.StatusCode, string(body))
	}
}
