package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AquaGatewaySpec defines the desired state of AquaGateway
type AquaGatewaySpec struct {
	Requirements       bool                `json:"requirements,required"`
	ServiceAccountName string              `json:"serviceAccount,omitempty"`
	DbSecretName       string              `json:"dbSecretName,omitempty"`
	DbSecretKey        string              `json:"dbSecretKey,omitempty"`
	RegistryData       *AquaDockerRegistry `json:"registry,omitempty"`

	GatewayService *AquaService `json:"deploy,required"`

	ExternalDb       *AquaDatabaseInformation `json:"externalDb,omitempty"`
	DbDeploymentName string                   `json:"aquaDb,omitempty"`
	DbSsl            bool                     `json:"ssl,omitempty"`
	DbAuditSsl       bool                     `json:"auditSsl,omitempty"`
	Openshift        bool                     `json:"openshift,omitempty"`
}

// AquaGatewayStatus defines the observed state of AquaGateway
type AquaGatewayStatus struct {
	Nodes []string            `json:"nodes"`
	State AquaDeploymentState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaGateway is the Schema for the aquagateways API
// +k8s:openapi-gen=true
type AquaGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AquaGatewaySpec   `json:"spec,omitempty"`
	Status AquaGatewayStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaGatewayList contains a list of AquaGateway
type AquaGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AquaGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AquaGateway{}, &AquaGatewayList{})
}
