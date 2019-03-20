package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AquaEnforcerSpec defines the desired state of AquaEnforcer
type AquaEnforcerSpec struct {
	Requirements       bool                `json:"requirements,required"`
	ServiceAccountName string              `json:"serviceAccount,omitempty"`
	RegistryData       *AquaDockerRegistry `json:"registry,omitempty"`

	Token           string            `json:"token,required"`
	Rbac            *AquaRbacSettings `json:"rbac,required"`
	EnforcerService *AquaService      `json:"deploy,required"`

	Gateway           *AquaGatewayInformation `json:"gateway,omitempty"`
	SendingHostImages bool                    `json:"sendingHostImages,omitempty"`
	RuncInterception  bool                    `json:"runcInterception,omitempty"`
}

// AquaEnforcerStatus defines the observed state of AquaEnforcer
type AquaEnforcerStatus struct {
	State AquaDeploymentState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaEnforcer is the Schema for the aquaenforcers API
// +k8s:openapi-gen=true
type AquaEnforcer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AquaEnforcerSpec   `json:"spec,omitempty"`
	Status AquaEnforcerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaEnforcerList contains a list of AquaEnforcer
type AquaEnforcerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AquaEnforcer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AquaEnforcer{}, &AquaEnforcerList{})
}
