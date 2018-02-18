package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -------------------------------------------------------------------------------- PodLabelConfig
// generation tags. The empty line after is IMPORTANT!
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodLabelConfig represents a set of labels to be applied to pods in a namespace
type PodLabelConfig struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the config
	Spec PodLabelConfigSpec `json:"spec,omitempty"`
}

// PodLabelConfigSpec describes the labels to apply to all pods in a namespace
type PodLabelConfigSpec struct {
	// Labels is a map of the labels to be applied to pods in the namespace
	Labels map[string]string `json:"labels,omitempty"`
}

// generation tags. The empty line after is IMPORTANT!
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodLabelConfigList is a list of PodAssignmentRules
type PodLabelConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`

	Items []PodLabelConfig `json:"items"`
}
