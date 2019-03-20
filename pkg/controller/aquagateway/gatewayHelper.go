package aquagateway

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type GatewayParameters struct {
	AquaGatewayDeploymentName string
	AquaGatewayServiceName    string
	AquaDbInternal            bool
	AquaGatewayImage          operatorv1alpha1.AquaImage
	AquaDbExteralData         operatorv1alpha1.AquaDatabaseInformation
	AquaServiceAccount        string
	AquaDbSecretName          string
	AquaDbSecretKey           string
	AquaDbName                string
	AquaGatewayServiceType    string
}

type AquaGatewayHelper struct {
	Parameters GatewayParameters
}

func newAquaGatewayHelper(cr *operatorv1alpha1.AquaGateway) *AquaGatewayHelper {
	params := GatewayParameters{
		AquaGatewayDeploymentName: fmt.Sprintf("%s-gateway", cr.Name),
		AquaGatewayServiceName:    fmt.Sprintf("%s-gateway-svc", cr.Name),
		AquaDbInternal:            true,
		AquaGatewayImage:          *cr.Spec.GatewayService.ImageData,
		AquaDbName:                cr.Spec.DbDeploymentName,
		AquaGatewayServiceType:    "ClusterIP",
	}

	if len(params.AquaDbName) == 0 {
		// Default AquaDatabase Name
		params.AquaDbName = "aqua"
	}

	if !cr.Spec.Requirements {
		params.AquaServiceAccount = cr.Spec.ServiceAccountName
		params.AquaDbSecretName = cr.Spec.DbSecretName
		params.AquaDbSecretKey = cr.Spec.DbSecretKey
	} else {
		params.AquaServiceAccount = fmt.Sprintf("%s-sa", cr.Name)
		params.AquaDbSecretName = fmt.Sprintf("%s-database-password", cr.Name)
		params.AquaDbSecretKey = "db-password"
	}

	if cr.Spec.ExternalDb != nil {
		params.AquaDbInternal = false
		params.AquaDbExteralData = *cr.Spec.ExternalDb
	} else {
		params.AquaDbExteralData = operatorv1alpha1.AquaDatabaseInformation{
			Name:          "scalock",
			Host:          fmt.Sprintf("%s-database-svc", params.AquaDbName),
			Port:          5432,
			Username:      "postgres",
			Password:      "password",
			AuditName:     "slk_audit",
			AuditHost:     fmt.Sprintf("%s-database-svc", params.AquaDbName),
			AuditPort:     5432,
			AuditUsername: "postgres",
			AuditPassword: "password",
		}
	}

	if len(cr.Spec.GatewayService.ServiceType) > 0 {
		params.AquaGatewayServiceType = cr.Spec.GatewayService.ServiceType
	}

	return &AquaGatewayHelper{
		Parameters: params,
	}
}

func (gw *AquaGatewayHelper) newDeployment(cr *operatorv1alpha1.AquaGateway) *appsv1.Deployment {
	labels := map[string]string{
		"app":                cr.Name + "-gateway",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
		"type":               "aqua-gateways",
	}
	annotations := map[string]string{
		"description": "Deploy the aqua gateway server",
	}
	env_vars := gw.getEnvVars(cr)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        gw.Parameters.AquaGatewayDeploymentName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: common.Int32Ptr(int32(cr.Spec.GatewayService.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-sa", cr.Name),
					Containers: []corev1.Container{
						{
							Name:            "aqua-gateway",
							Image:           fmt.Sprintf("%s/%s:%s", gw.Parameters.AquaGatewayImage.Registry, gw.Parameters.AquaGatewayImage.Repository, gw.Parameters.AquaGatewayImage.Tag),
							ImagePullPolicy: corev1.PullPolicy(gw.Parameters.AquaGatewayImage.PullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3622,
								},
							},
							Env: env_vars,
						},
					},
				},
			},
		},
	}

	if cr.Spec.GatewayService.Resources != nil {
		deployment.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.GatewayService.Resources
	}

	if cr.Spec.GatewayService.LivenessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.GatewayService.LivenessProbe
	}

	if cr.Spec.GatewayService.ReadinessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.GatewayService.ReadinessProbe
	}

	if cr.Spec.GatewayService.NodeSelector != nil {
		if len(cr.Spec.GatewayService.NodeSelector) > 0 {
			deployment.Spec.Template.Spec.NodeSelector = cr.Spec.GatewayService.NodeSelector
		}
	}

	if cr.Spec.GatewayService.Affinity != nil {
		deployment.Spec.Template.Spec.Affinity = cr.Spec.GatewayService.Affinity
	}

	if cr.Spec.GatewayService.Tolerations != nil {
		if len(cr.Spec.GatewayService.Tolerations) > 0 {
			deployment.Spec.Template.Spec.Tolerations = cr.Spec.GatewayService.Tolerations
		}
	}

	if cr.Spec.Openshift {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &cr.Spec.Openshift,
		}
	}

	return deployment
}

func (gw *AquaGatewayHelper) getEnvVars(cr *operatorv1alpha1.AquaGateway) []corev1.EnvVar {
	result := []corev1.EnvVar{
		{
			Name:  "SCALOCK_GATEWAY_PUBLIC_IP",
			Value: gw.Parameters.AquaGatewayServiceName,
		},
		{
			Name:  "SCALOCK_DBUSER",
			Value: gw.Parameters.AquaDbExteralData.Username,
		},
		{
			Name:  "SCALOCK_DBNAME",
			Value: gw.Parameters.AquaDbExteralData.Name,
		},
		{
			Name:  "SCALOCK_DBHOST",
			Value: gw.Parameters.AquaDbExteralData.Host,
		},
		{
			Name:  "SCALOCK_DBPORT",
			Value: fmt.Sprintf("%d", gw.Parameters.AquaDbExteralData.Port),
		},
		{
			Name:  "SCALOCK_AUDIT_DBUSER",
			Value: gw.Parameters.AquaDbExteralData.AuditUsername,
		},
		{
			Name:  "SCALOCK_AUDIT_DBNAME",
			Value: gw.Parameters.AquaDbExteralData.AuditName,
		},
		{
			Name:  "SCALOCK_AUDIT_DBHOST",
			Value: gw.Parameters.AquaDbExteralData.AuditHost,
		},
		{
			Name:  "SCALOCK_AUDIT_DBPORT",
			Value: fmt.Sprintf("%d", gw.Parameters.AquaDbExteralData.AuditPort),
		},
		{
			Name:  "SCALOCK_LOG_LEVEL",
			Value: "DEBUG",
		},
		{
			Name:  "HEALTH_MONITOR",
			Value: "0.0.0.0:8082",
		},
	}

	if !gw.Parameters.AquaDbInternal {
		scalock_password := corev1.EnvVar{
			Name:  "SCALOCK_DBPASSWORD",
			Value: gw.Parameters.AquaDbExteralData.Password,
		}
		scalock_audit_password := corev1.EnvVar{
			Name:  "SCALOCK_AUDIT_DBPASSWORD",
			Value: gw.Parameters.AquaDbExteralData.AuditPassword,
		}

		result = append(result, scalock_password, scalock_audit_password)
	} else {
		scalock_password := corev1.EnvVar{
			Name: "SCALOCK_DBPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: gw.Parameters.AquaDbSecretName,
					},
					Key: gw.Parameters.AquaDbSecretKey,
				},
			},
		}
		scalock_audit_password := corev1.EnvVar{
			Name: "SCALOCK_AUDIT_DBPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: gw.Parameters.AquaDbSecretName,
					},
					Key: gw.Parameters.AquaDbSecretKey,
				},
			},
		}

		result = append(result, scalock_password, scalock_audit_password)
	}

	return result
}

func (gw *AquaGatewayHelper) newService(cr *operatorv1alpha1.AquaGateway) *corev1.Service {
	labels := map[string]string{
		"app":                cr.Name + "-gateway",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy the aqua gateway server",
	}
	selectors := map[string]string{
		"type": "aqua-gateways",
	}
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        gw.Parameters.AquaGatewayServiceName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(gw.Parameters.AquaGatewayServiceType),
			Selector: selectors,
			Ports: []corev1.ServicePort{
				{
					Port:       3622,
					TargetPort: intstr.FromInt(3622),
				},
			},
		},
	}

	return service
}
