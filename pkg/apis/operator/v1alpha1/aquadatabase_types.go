package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AquaDatabaseSpec defines the desired state of AquaDatabase
type AquaDatabaseSpec struct {
	Requirements       bool                `json:"requirements,required"`
	ServiceAccountName string              `json:"serviceAccount,omitempty"`
	DbSecretName       string              `json:"dbSecretName,omitempty"`
	DbSecretKey        string              `json:"dbSecretKey,omitempty"`
	RegistryData       *AquaDockerRegistry `json:"registry,omitempty"`

	DbService *AquaService `json:"deploy,required"`
	Openshift bool         `json:"openshift,omitempty"`
}

// AquaDatabaseStatus defines the observed state of AquaDatabase
type AquaDatabaseStatus struct {
	Nodes []string            `json:"nodes"`
	State AquaDeploymentState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaDatabase is the Schema for the aquadatabases API
// +k8s:openapi-gen=true
type AquaDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AquaDatabaseSpec   `json:"spec,omitempty"`
	Status AquaDatabaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaDatabaseList contains a list of AquaDatabase
type AquaDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AquaDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AquaDatabase{}, &AquaDatabaseList{})
}
