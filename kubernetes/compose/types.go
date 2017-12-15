package compose

import (
	composetypes "github.com/docker/cli/cli/compose/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImpersonationConfig contains the data required to impersonate a user.
type ImpersonationConfig struct {
	// UserName is the username to impersonate on each request.
	UserName string
	// Groups are the groups to impersonate on each request.
	Groups []string
	// Extra is a free-form field which can be used to link some authentication information
	// to authorization information.  This field allows you to impersonate it.
	Extra map[string][]string
}

// Stack is the internal representation of a compose stack
// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Stack struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec   StackSpec    `json:"spec,omitempty"`
	Status *StackStatus `json:"status,omitempty"`
}

// StackStatus is the current status of a stack
type StackStatus struct {
	Phase   StackPhase
	Message string
}

// Config contains the stack description
type Config composetypes.Config

// StackSpec is the Spec field of a Stack
type StackSpec struct {
	ComposeFile string              `json:"composeFile,omitempty"`
	Stack       *Config             `json:"stack,omitempty"`
	Owner       ImpersonationConfig `json:"owner,omitempty"`
}

// StackPhase is the current status phase.
type StackPhase string

// These are valid conditions of a stack.
const (
	// StackAvailable means the stack is available.
	StackAvailable StackPhase = "Available"
	// StackProgressing means the deployment is progressing.
	StackProgressing StackPhase = "Progressing"
	// StackFailure is added in a stack when one of its members fails to be created
	// or deleted.
	StackFailure StackPhase = "Failure"
)

// StackList is a list of stacks
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StackList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Stack
}

// Owner is the user who created the stack
type Owner struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Owner ImpersonationConfig
}

// OwnerList is a list of owners
type OwnerList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Owner
}
