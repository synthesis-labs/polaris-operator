package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisBuildPipelineSourceSpec defines the "source" part of the spec
type PolarisBuildPipelineSourceSpec struct {
	CodeCommitRepo string `json:"codecommitrepo"`
	Branch         string `json:"branch"`
}

// PolarisBuildPipelineBuildSpec defines a particular "build" part of the spec
type PolarisBuildPipelineBuildSpec struct {
	Name                string `json:"name"`
	Buildspec           string `json:"buildspec"`
	ContainerRepository string `json:"containerrepository"`
	Tag                 string `json:"tag"`
}

// PolarisBuildPipelineSpec defines the desired state of PolarisBuildPipeline
type PolarisBuildPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Source PolarisBuildPipelineSourceSpec  `json:"source"`
	Builds []PolarisBuildPipelineBuildSpec `json:"builds"`
}

// PolarisBuildPipelineStatus defines the observed state of PolarisBuildPipeline
type PolarisBuildPipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	StackCreationAttempted bool   `json:"stackCreationAttempted"`
	StackResponse          string `json:"stackResponse"`
	StackError             string `json:"stackError"`
	StackName              string `json:"stackName"`
	PipelineName           string `json:"pipelineName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisBuildPipeline is the Schema for the polarisbuildpipelines API
// +k8s:openapi-gen=true
type PolarisBuildPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisBuildPipelineSpec   `json:"spec,omitempty"`
	Status PolarisBuildPipelineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisBuildPipelineList contains a list of PolarisBuildPipeline
type PolarisBuildPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisBuildPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisBuildPipeline{}, &PolarisBuildPipelineList{})
}
