package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// -------------------------------------------------------------------------------- NodeAssignmentGroup
// generation tags. The empty line after is IMPORTANT!
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeAssignmentGroup represents the configuration of a group of nodes that will be auto-labeled
type NodeAssignmentGroup struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the group and it's assignments
	Spec NodeAssignmentGroupSpec `json:"spec,omitempty"`

	// Status represents the current status of the group assignments.
	Status NodeAssignmentGroupStatus `json:"status,omitempty"`
}

// NodeAssignmentGroupSpec describes the group and it's assignments
type NodeAssignmentGroupSpec struct {
	// TargetLabels is optional. If not provided, the group will match all nodes in the cluster.
	// +optional
	TargetLabels labels.Set `json:"targetLabels,omitempty"`

	// Assignments is the array of assignments to be applied. This list should be ordered by the user
	// with the most important assignments first.
	// +optional
	DefaultAssignment *NodeAssignment `json:"defaultAssignment,omitempty"`

	// Assignments is the array of assignments to be applied. This list should be ordered by the user
	// with the most important assignments first.
	// +patchStrategy=merge
	Assignments []NodeAssignment `json:"assignments,omitempty"`
}

const (
	// NodeAssignmentDefaultTaintEffect defines the default taint effect to be
	// used when not specified in the resource
	NodeAssignmentTaintEffectDefault = corev1.TaintEffectNoSchedule

	// NodeAssignmentTaintEffectNotSpecified means that the user did not specify a taint effect
	NodeAssignmentTaintEffectNotSpecified = ""
)

// NodeAssignment describes the assignments possible for the group
// and the number of nodes for the assignment
type NodeAssignment struct {
	// Name is used when applying the assignment label to the nodes
	Name string `json:"name,omitempty"`

	// GroupMode determines whether labels, taints, labels and taints, or nothing
	// is applied to nodes that match the group
	Mode NodeAssignmentMode `json:"mode,omitempty"`

	// TaintEffect controls the effect of the taint. Possible values
	// come from the upstream type
	// +optional
	TaintEffect corev1.TaintEffect `json:"taintEffect,omitempty"`

	// NumDesired is the number of nodes that should be assigned to this group. Default: 0
	NumDesired int `json:"numDesired,omitempty"`
}

// NodeAssignmentMode defines the operation mode of the rule
type NodeAssignmentMode string

const (
	// NodeAssignmentModeDefault sets the default behavior to "LabelOnly"
	NodeAssignmentModeDefault NodeAssignmentMode = "LabelOnly"

	// NodeAssignmentModeLabelOnly tells the system to only apply labels to the node
	NodeAssignmentModeLabelOnly NodeAssignmentMode = "LabelOnly"

	// NodeAssignmentModeLabelAndTaint tells the system to apply both labels and taints for the rule
	NodeAssignmentModeLabelAndTaint NodeAssignmentMode = "LabelAndTaint"

	// NodeAssignmentModeUndefined means that the resource did not have this
	// property set and the default behavior will be used
	NodeAssignmentModeUndefined NodeAssignmentMode = ""
)

// NodeAssignmentGroupStatus represents the current status of the group.
type NodeAssignmentGroupStatus struct {
	// NumMatched represents the number of nodes that matched the targetlabels
	NumMatched int64 `json:"numMatched,omitempty"`

	// // NumSatisfied represents the total number of assignments that have been satisifed.
	// // If this is less than NumMatched then there weren't enough maching nodes to
	// // fufill all the assignments
	// NumSatisfied int64 `json:"numSatisfied"`

	// // State reports the overall health of the group
	// State NodeAssignmentGroupState `json:"state"`

	// // AssignmentStates reports the satisfaction for each assignment
	// AssignmentStates NodeAssignmentGroupState `json:"state"`
}

// NodeAssignmentGroupState reports the overall health of the group
type NodeAssignmentGroupState string

const (
	// NodeAssignmentGroupStateSatisfied means that the group
	// has matched and operated on enough nodes to satisfy all assignments
	NodeAssignmentGroupStateSatisfied NodeAssignmentGroupState = "Satisfied"

	// NodeAssignmentGroupStateNotSatisfied means that the group
	// did not find enough nodes to satisfy all assignments
	NodeAssignmentGroupStateNotSatisfied NodeAssignmentGroupState = "NotSatisfied"

	// NodeAssignmentGroupStateError means that the controller
	// was unable to process the group properly
	NodeAssignmentGroupStateError NodeAssignmentGroupState = "Error"
)

// AssignmentStates reports the satisfaction for each assignment
type AssignmentStates struct {
	// Name is the name of the assignment
	Name string `json:"name,omitempty"`

	// NumSatisfied represents the number of nodes that were assigned
	NumAssigned int64 `json:"numAssigned,omitempty"`
}

// generation tags. The empty line after is IMPORTANT!
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeAssignmentGroupList is a list of NodeAssignmentGroups
type NodeAssignmentGroupList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of NodeAssignmentGroups
	Items []NodeAssignmentGroup `json:"items"`
}

// -------------------------------------------------------------------------------- PodAssignmentRule
// generation tags. The empty line after is IMPORTANT!
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodAssignmentRule describes pods to match and attributes to apply to them
type PodAssignmentRule struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the rule
	Spec PodAssignmentRuleSpec `json:"spec"`
}

// PodAssignmentRuleSpec defines the behavior of the PodAssignmentRule
type PodAssignmentRuleSpec struct {

	// TargetLabels defines which pods this rule will be applied to. Optional.
	// When not given, the rule will match all pods.
	// +optional
	TargetLabels labels.Set `json:"targetLabels,omitempty"`

	// Scheduling defines the scheduling objects to be applied to the pod
	Scheduling PodAssignmentRuleScheduling `json:"scheduling"`
}

// PodAssignmentRuleScheduling defines the scheduling objects to be applied to the pod
type PodAssignmentRuleScheduling struct {
	// MergeStrategy defines the behavior of the rule when pods already have existing
	// scheduling details defined
	MergeStrategy PodAssignmentRuleSchedulingMergeStrategy `json:"mergeStrategy"`

	// NodeSelector is a simple key-value matching for nodes
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity is the upstream pod affinity resource
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations is a list of upstream pod toleration resources
	// +optional
	// +patchStrategy=merge
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// PodAssignmentRuleSchedulingMergeStrategy defines the behavior of the rule when pods already have existing
// scheduling details defined
type PodAssignmentRuleSchedulingMergeStrategy string

const (
	// PodAssignmentRuleSchedulingMergeStrategyDefault is the default behavior to be used
	// when on is provided
	PodAssignmentRuleSchedulingMergeStrategyDefault PodAssignmentRuleSchedulingMergeStrategy = "OverwriteAll"

	// PodAssignmentRuleSchedulingMergeStrategyOverwriteAll tells the system to overwrite
	// any scheduling details in the pod with the details in the rule
	PodAssignmentRuleSchedulingMergeStrategyOverwriteAll PodAssignmentRuleSchedulingMergeStrategy = "OverwriteAll"

	// PodAssignmentRuleSchedulingMergeStrategyUndefined means that the document did not
	// include a strategy and the default will be used.
	PodAssignmentRuleSchedulingMergeStrategyUndefined PodAssignmentRuleSchedulingMergeStrategy = ""
)

// generation tags. The empty line after is IMPORTANT!
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodAssignmentRuleList is a list of PodAssignmentRules
type PodAssignmentRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`

	Items []PodAssignmentRule `json:"items"`
}

// -------------------------------------------------------------------------------- ClusterPodAssignmentRule
// generation tags. The empty line after is IMPORTANT!
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPodAssignmentRule defines PodAssignmentRules that are applied cluster-wide
type ClusterPodAssignmentRule struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the rule
	Spec PodAssignmentRuleSpec `json:"spec"`
}

// generation tags. The empty line after is IMPORTANT!
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPodAssignmentRuleList is a list of ClusterPodAssignmentRules
type ClusterPodAssignmentRuleList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterPodAssignmentRule `json:"items"`
}
