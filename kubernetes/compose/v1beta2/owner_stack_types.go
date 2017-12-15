package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/docker/cli/kubernetes/compose"
)

// Owner describes the user who created the stack
// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +subresource-request
type Owner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Owner             compose.ImpersonationConfig `json:"owner,omitempty"`
}

// OwnerList is a list of owners
type OwnerList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Owner
}
