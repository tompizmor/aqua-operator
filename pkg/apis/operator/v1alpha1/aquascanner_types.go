package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AquaScannerSpec defines the desired state of AquaScanner
type AquaScannerSpec struct {
	Requirements       bool                `json:"requirements,required"`
	ServiceAccountName string              `json:"serviceAccount,omitempty"`
	RegistryData       *AquaDockerRegistry `json:"registry,omitempty"`

	ScannerService *AquaService `json:"deploy,required"`
	Login          *AquaLogin   `json:"login,required"`
	Openshift      bool         `json:"openshift,omitempty"`
}

// AquaScannerStatus defines the observed state of AquaScanner
type AquaScannerStatus struct {
	Nodes []string            `json:"nodes"`
	State AquaDeploymentState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaScanner is the Schema for the aquascanners API
// +k8s:openapi-gen=true
type AquaScanner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AquaScannerSpec   `json:"spec,omitempty"`
	Status AquaScannerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaScannerList contains a list of AquaScanner
type AquaScannerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AquaScanner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AquaScanner{}, &AquaScannerList{})
}
