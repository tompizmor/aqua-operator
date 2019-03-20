package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AquaServerSpec defines the desired state of AquaServer
type AquaServerSpec struct {
	Requirements       bool                `json:"requirements,required"`
	Add                bool                `json:"add,required"`
	ServiceAccountName string              `json:"serviceAccount,omitempty"`
	DbSecretName       string              `json:"dbSecretName,omitempty"`
	DbSecretKey        string              `json:"dbSecretKey,omitempty"`
	Password           string              `json:"password,omitempty"`
	RegistryData       *AquaDockerRegistry `json:"registry,omitempty"`

	ServerService *AquaService `json:"deploy,required"`

	DbDeploymentName string                   `json:"aquaDb,omitempty"`
	ExternalDb       *AquaDatabaseInformation `json:"externalDb,omitempty"`

	LicenseToken    string `json:"licenseToken,omitempty"`
	AdminPassword   string `json:"adminPassword,omitempty"`
	AquaSslCertPath string `json:"sslCertPath,omitempty"`
	ClusterMode     bool   `json:"clusterMode,omitempty"`
	DbSsl           bool   `json:"ssl,omitempty"`
	DbAuditSsl      bool   `json:"auditSsl,omitempty"`
	Dockerless      bool   `json:"dockerless,omitempty"`
	Openshift       bool   `json:"openshift,omitempty"`
}

// AquaServerStatus defines the observed state of AquaServer
type AquaServerStatus struct {
	Nodes []string            `json:"nodes"`
	State AquaDeploymentState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaServer is the Schema for the aquaservers API
// +k8s:openapi-gen=true
type AquaServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AquaServerSpec   `json:"spec,omitempty"`
	Status AquaServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AquaServerList contains a list of AquaServer
type AquaServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AquaServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AquaServer{}, &AquaServerList{})
}
