package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -------------------------------------------------------------------------------- WorkshopAttendee
// generation tags. The empty line after is IMPORTANT!
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkshopAttendee represents an attendee that needs resource provisioned
type WorkshopAttendee struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the config
	Spec WorkshopAttendeeSpec `json:"spec,omitempty"`

	// Status defines the current state
	Status WorkshopAttendeeStatus `json:"status,omitempty"`
}

// WorkshopAttendeeSpec describes extra attributes of the attendee
type WorkshopAttendeeSpec struct {
	// Email is the address that the credentials should be sent to
	Email string `json:"email,omitempty"`
}

type WorkshopAttendeeState string

const (
	WorkshopAttendeeStateReady    WorkshopAttendeeState = "Ready"
	WorkshopAttendeeStateCreating WorkshopAttendeeState = "Creating"
	WorkshopAttendeeStateDeleting WorkshopAttendeeState = "Deleting"
)

// WorkshopAttendeeStatus describes the current state of provisioning
type WorkshopAttendeeStatus struct {
	// State is the current overall state of provisioning
	State WorkshopAttendeeState `json:"state,omitempty"`

	// Notified is the last email notified on completion
	Notified string `json:"notified,omitempty"`

	// Children defines the time each child resource was last made
	Children map[string]metav1.Time `json:"children,omitempty"`

	// Kubeconfig is the multiline yaml string that represents a valid kubectl config file for a completed attendee
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

// generation tags. The empty line after is IMPORTANT!
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkshopAttendeeList is a list of PodAssignmentRules
type WorkshopAttendeeList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`

	Items []WorkshopAttendee `json:"items"`
}
