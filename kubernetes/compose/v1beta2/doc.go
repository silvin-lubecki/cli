// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// Package v1beta2 is the second version of the stack, containing a structured spec
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:defaulter-gen=TypeMeta
// +groupName=compose.docker.com
package v1beta2 // import "github.com/docker/cli/kubernetes/compose/v1beta2"
