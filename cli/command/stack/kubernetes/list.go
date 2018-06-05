package kubernetes

import (
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStacks lists the kubernetes stacks
func GetStacks(kubeCli *KubeCli, opts options.List) ([]*formatter.Stack, []error, error) {
	clientFactory := &stackClientFactoryImpl{kubeCli: *kubeCli}
	nLister := &userVisibleNamespaceLister{dockerCli: kubeCli}
	kubeConfig := kubeCli.ConfigFile().Kubernetes
	return getStacks(clientFactory, nLister, opts, kubeConfig)
}

func getStacks(clientFactory stackClientFactory, nLister namespaceLister, opts options.List, kubeConfig *configfile.KubernetesConfig) ([]*formatter.Stack, []error, error) {
	if opts.AllNamespaces || len(opts.Namespaces) == 0 {
		if isAllNamespacesDisabled(kubeConfig) {
			opts.AllNamespaces = true
		}
		return getStacksWithAllNamespaces(clientFactory, nLister, opts)
	}
	return getStacksWithNamespaces(clientFactory, removeDuplicates(opts.Namespaces))
}

func isAllNamespacesDisabled(kubeCliConfig *configfile.KubernetesConfig) bool {
	return kubeCliConfig == nil || kubeCliConfig != nil && kubeCliConfig.AllNamespaces != "disabled"
}

func convertStacks(lister stackLister) ([]*formatter.Stack, []error, error) {
	stacks, errs, err := lister.List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	var formattedStacks []*formatter.Stack
	for _, stack := range stacks {
		formattedStacks = append(formattedStacks, &formatter.Stack{
			Name:         stack.name,
			Services:     len(stack.getServices()),
			Orchestrator: "Kubernetes",
			Namespace:    stack.namespace,
		})
	}
	return formattedStacks, errs, nil
}

func getStacksWithAllNamespaces(clientFactory stackClientFactory, nLister namespaceLister, opts options.List) ([]*formatter.Stack, []error, error) {
	lister, err := clientFactory.Stacks("", opts.AllNamespaces)
	if err != nil {
		return nil, nil, err
	}
	stacks, errs, err := convertStacks(lister)
	if !apierrs.IsForbidden(err) {
		return stacks, errs, err
	}
	namespaces, err2 := nLister.List()
	if err2 != nil {
		return nil, nil, errors.Wrap(err2, "failed to query user visible namespaces")
	}
	if namespaces == nil {
		// UCP API not present, fall back to Kubernetes error
		return nil, nil, err
	}
	opts.AllNamespaces = false
	return getStacksWithNamespaces(clientFactory, namespaces)
}

func getStacksWithNamespaces(clientFactory stackClientFactory, namespaces []string) ([]*formatter.Stack, []error, error) {
	var listers []stackLister
	for _, namespace := range namespaces {
		lister, err := clientFactory.Stacks(namespace, false)
		if err != nil {
			return nil, nil, err
		}
		listers = append(listers, lister)
	}
	return convertStacks(&compositeStackLister{listers: listers})
}

func removeDuplicates(namespaces []string) []string {
	found := make(map[string]bool)
	results := namespaces[:0]
	for _, n := range namespaces {
		if !found[n] {
			results = append(results, n)
			found[n] = true
		}
	}
	return results
}
