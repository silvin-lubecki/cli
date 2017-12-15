package v1beta2

import (
	"encoding/json"

	"github.com/docker/cli/kubernetes/compose"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// StackList is a list of stacks
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Stack `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Stack is v1beta2's representation of a Stack
// +k8s:openapi-gen=true
// +resource:path=stacks,strategy=StackStrategy
// +subresource:request=Owner,path=owner,rest=OwnerStackREST
// +subresource:request=ComposeFile,path=composefile,rest=ComposeFileStackREST
type Stack struct {
	StackImpl
}

// StackImpl contains the stack's actual fields
type StackImpl struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec    `json:"spec,omitempty"`
	Status *StackStatus `json:"status,omitempty"`
}

// StackSpec defines the desired state of Stack
type StackSpec struct {
	Stack *compose.Config `json:"stack,omitempty"`
}

// StackPhase is the deployment phase of a stack
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

// StackStatus defines the observed state of Stack
type StackStatus struct {
	// Current condition of the stack.
	// +optional
	Phase StackPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=StackPhase"`
	// A human readable message indicating details about the stack.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// Clone clones a Stack
func (s *Stack) Clone() (*Stack, error) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	return s.DeepCopy(), nil
}

/* Do not remove me! This explicit implementation of json.Marshaler overrides
 * the default behavior of ToUnstructured(), which would otherwise convert
 * all field names to lowercase, which makes patching fail in case of update
 * conflict
 *
 */

// MarshalJSON implements the json.Marshaler interface
func (s *Stack) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.StackImpl)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *Stack) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.StackImpl)
}
