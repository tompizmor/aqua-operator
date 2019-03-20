package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type AquaDockerRegistry struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AquaDatabaseInformation struct {
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int64  `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	AuditName     string `json:"auditName"`
	AuditHost     string `json:"auditHost"`
	AuditPort     int64  `json:"auditPort"`
	AuditUsername string `json:"auditUsername"`
	AuditPassword string `json:"auditPassword"`
}

type AquaImage struct {
	Repository string `json:"repository"`
	Registry   string `json:"registry"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
}

// AquaService Struct for deployment spec
type AquaService struct {
	// Number of instances to deploy for a specific aqua deployment.
	Replicas       int64                        `json:"replicas"`
	ServiceType    string                       `json:"service,omitempty"`
	ImageData      *AquaImage                   `json:"image,omitempty"`
	Resources      *corev1.ResourceRequirements `json:"resources,omitempty"`
	LivenessProbe  *corev1.Probe                `json:"livenessProbe,omitempty"`
	ReadinessProbe *corev1.Probe                `json:"readinessProbe,omitempty"`
	NodeSelector   map[string]string            `json:"nodeSelector,omitempty"`
	Affinity       *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations    []corev1.Toleration          `json:"tolerations,omitempty"`
}

type AquaRbacSettings struct {
	Enable     bool   `json:"enable"`
	RoleRef    string `json:"roleref"`
	Openshift  bool   `json:"openshift"`
	Privileged bool   `json:"privileged"`
}

type AquaGatewayInformation struct {
	Host string `json:"host"`
	Port int64  `json:"port"`
}

type AquaLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

type AquaScannerCliScale struct {
	Deploy           *AquaService `json:"deploy"`
	Name             string       `json:"name"`
	Max              int64        `json:"max"`
	Min              int64        `json:"min"`
	ImagesPerScanner int64        `json:"imagesPerScanner"`
}

type AquaDeploymentState string

const (
	AquaDeploymentStatePending     AquaDeploymentState = "Pending"
	AquaDeploymentStateWaitingDB   AquaDeploymentState = "Waiting For DataBase"
	AquaDeploymentStateWaitingAqua AquaDeploymentState = "Waiting For Aqua Server and Gateway"
	AquaDeploymentStateRunning     AquaDeploymentState = "Running"
)
