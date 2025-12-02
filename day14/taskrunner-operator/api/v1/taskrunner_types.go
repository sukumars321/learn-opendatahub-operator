/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskRunnerSpec defines the desired state of TaskRunner
type TaskRunnerSpec struct {
	// Command to execute
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Command []string `json:"command"`

	// Container image to run
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Number of parallel executions
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Parallelism int32 `json:"parallelism,omitempty"`

	// Completion deadline in seconds
	// +kubebuilder:validation:Minimum=1
	DeadlineSeconds *int64 `json:"deadlineSeconds,omitempty"`

	// Management state following ODH pattern
	// +kubebuilder:default=Managed
	// +kubebuilder:validation:Enum=Managed;Unmanaged;Removed
	ManagementState string `json:"managementState,omitempty"`
}

// TaskRunnerStatus defines the observed state of TaskRunner
type TaskRunnerStatus struct {
	// Current phase of the task
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
	Phase string `json:"phase,omitempty"`

	// Number of active jobs
	Active int32 `json:"active,omitempty"`

	// Number of successful completions
	Succeeded int32 `json:"succeeded,omitempty"`

	// Number of failed attempts
	Failed int32 `json:"failed,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TaskRunner is the Schema for the taskrunners API
type TaskRunner struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of TaskRunner
	// +required
	Spec TaskRunnerSpec `json:"spec"`

	// status defines the observed state of TaskRunner
	// +optional
	Status TaskRunnerStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// TaskRunnerList contains a list of TaskRunner
type TaskRunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskRunner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskRunner{}, &TaskRunnerList{})
}
