package aquaserver

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServerParameters struct {
	AquaServerDeploymentName string
	AquaServerServiceName    string
	AquaServerSecretsName    string
	AquaDbInternal           bool
	AquaServerImage          operatorv1alpha1.AquaImage
	AquaDbExteralData        operatorv1alpha1.AquaDatabaseInformation
	AquaServiceAccount       string
	AquaDbSecretName         string
	AquaDbSecretKey          string
	AquaDbName               string
	AquaServerServiceType    string
	AquaLicenseTokenKey      string
	AquaAdminPasswordKey     string
	Dockerless               string
}

type AquaServerHelper struct {
	Parameters ServerParameters
}

func newAquaServerHelper(cr *operatorv1alpha1.AquaServer) *AquaServerHelper {
	params := ServerParameters{
		AquaServerDeploymentName: fmt.Sprintf("%s-server", cr.Name),
		AquaServerServiceName:    fmt.Sprintf("%s-server-svc", cr.Name),
		AquaServerSecretsName:    fmt.Sprintf("%s-server-secrets", cr.Name),
		AquaDbInternal:           true,
		AquaServerImage:          *cr.Spec.ServerService.ImageData,
		AquaDbName:               cr.Spec.DbDeploymentName,
		AquaServerServiceType:    "ClusterIP",
		AquaLicenseTokenKey:      "license-token",
		AquaAdminPasswordKey:     "admin-password",
		Dockerless:               "0",
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

	if !cr.Spec.Dockerless {
		params.Dockerless = "1"
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

	if len(cr.Spec.ServerService.ServiceType) > 0 {
		params.AquaServerServiceType = cr.Spec.ServerService.ServiceType
	}

	return &AquaServerHelper{
		Parameters: params,
	}
}

func (sr *AquaServerHelper) newDeployment(cr *operatorv1alpha1.AquaServer) *appsv1.Deployment {
	labels := map[string]string{
		"app":                cr.Name + "-server",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
		"type":               "aqua-servers",
	}
	annotations := map[string]string{
		"description": "Deploy the aqua console server",
	}
	envvars := sr.getEnvVars(cr)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        sr.Parameters.AquaServerDeploymentName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: common.Int32Ptr(int32(cr.Spec.ServerService.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: sr.Parameters.AquaServiceAccount,
					Containers: []corev1.Container{
						{
							Name:            "aqua-server",
							Image:           fmt.Sprintf("%s/%s:%s", sr.Parameters.AquaServerImage.Registry, sr.Parameters.AquaServerImage.Repository, sr.Parameters.AquaServerImage.Tag),
							ImagePullPolicy: corev1.PullPolicy(sr.Parameters.AquaServerImage.PullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3622,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "docker-socket-mount",
									MountPath: "/var/run/docker.sock",
								},
							},
							Env: envvars,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "docker-socket-mount",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/docker.sock",
								},
							},
						},
					},
				},
			},
		},
	}

	if cr.Spec.ServerService.Resources != nil {
		deployment.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.ServerService.Resources
	}

	if cr.Spec.ServerService.LivenessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.ServerService.LivenessProbe
	}

	if cr.Spec.ServerService.ReadinessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.ServerService.ReadinessProbe
	}

	if cr.Spec.ServerService.NodeSelector != nil {
		if len(cr.Spec.ServerService.NodeSelector) > 0 {
			deployment.Spec.Template.Spec.NodeSelector = cr.Spec.ServerService.NodeSelector
		}
	}

	if cr.Spec.ServerService.Affinity != nil {
		deployment.Spec.Template.Spec.Affinity = cr.Spec.ServerService.Affinity
	}

	if cr.Spec.ServerService.Tolerations != nil {
		if len(cr.Spec.ServerService.Tolerations) > 0 {
			deployment.Spec.Template.Spec.Tolerations = cr.Spec.ServerService.Tolerations
		}
	}

	if cr.Spec.Openshift {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &cr.Spec.Openshift,
		}
	}

	return deployment
}

func (sr *AquaServerHelper) getEnvVars(cr *operatorv1alpha1.AquaServer) []corev1.EnvVar {
	result := []corev1.EnvVar{
		{
			Name:  "SCALOCK_DBUSER",
			Value: sr.Parameters.AquaDbExteralData.Username,
		},
		{
			Name:  "SCALOCK_DBNAME",
			Value: sr.Parameters.AquaDbExteralData.Name,
		},
		{
			Name:  "SCALOCK_DBHOST",
			Value: sr.Parameters.AquaDbExteralData.Host,
		},
		{
			Name:  "SCALOCK_DBPORT",
			Value: fmt.Sprintf("%d", sr.Parameters.AquaDbExteralData.Port),
		},
		{
			Name:  "SCALOCK_AUDIT_DBUSER",
			Value: sr.Parameters.AquaDbExteralData.AuditUsername,
		},
		{
			Name:  "SCALOCK_AUDIT_DBNAME",
			Value: sr.Parameters.AquaDbExteralData.AuditName,
		},
		{
			Name:  "SCALOCK_AUDIT_DBHOST",
			Value: sr.Parameters.AquaDbExteralData.AuditHost,
		},
		{
			Name:  "SCALOCK_AUDIT_DBPORT",
			Value: fmt.Sprintf("%d", sr.Parameters.AquaDbExteralData.AuditPort),
		},
		{
			Name:  "AQUA_DOCKERLESS_SCANNING",
			Value: sr.Parameters.Dockerless,
		},
		{
			Name:  "AQUA_PPROF_ENABLED",
			Value: sr.Parameters.Dockerless,
		},
		{
			Name:  "DISABLE_IP_BAN",
			Value: sr.Parameters.Dockerless,
		},
	}

	if !sr.Parameters.AquaDbInternal {
		scalockpassword := corev1.EnvVar{
			Name:  "SCALOCK_DBPASSWORD",
			Value: sr.Parameters.AquaDbExteralData.Password,
		}
		scalockauditpassword := corev1.EnvVar{
			Name:  "SCALOCK_AUDIT_DBPASSWORD",
			Value: sr.Parameters.AquaDbExteralData.AuditPassword,
		}

		result = append(result, scalockpassword, scalockauditpassword)
	} else {
		scalockpassword := corev1.EnvVar{
			Name: "SCALOCK_DBPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sr.Parameters.AquaDbSecretName,
					},
					Key: sr.Parameters.AquaDbSecretKey,
				},
			},
		}
		scalockauditpassword := corev1.EnvVar{
			Name: "SCALOCK_AUDIT_DBPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sr.Parameters.AquaDbSecretName,
					},
					Key: sr.Parameters.AquaDbSecretKey,
				},
			},
		}

		result = append(result, scalockpassword, scalockauditpassword)
	}

	if len(cr.Spec.LicenseToken) > 0 {
		result = append(result, corev1.EnvVar{
			Name: "LICENSE_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sr.Parameters.AquaServerSecretsName,
					},
					Key: sr.Parameters.AquaLicenseTokenKey,
				},
			},
		})
	}

	if len(cr.Spec.AdminPassword) > 0 {
		result = append(result, corev1.EnvVar{
			Name: "ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sr.Parameters.AquaServerSecretsName,
					},
					Key: sr.Parameters.AquaAdminPasswordKey,
				},
			},
		})
	}

	if len(cr.Spec.AquaSslCertPath) > 0 {
		result = append(result, corev1.EnvVar{
			Name:  "AQUA_SSL_CERT_PATH",
			Value: cr.Spec.AquaSslCertPath,
		})
	}

	if cr.Spec.DbSsl {
		result = append(result, corev1.EnvVar{
			Name:  "SCALOCK_DBSSL",
			Value: "require",
		})
	}

	if cr.Spec.DbAuditSsl {
		result = append(result, corev1.EnvVar{
			Name:  "SCALOCK_AUDIT_DBSSL",
			Value: "require",
		})
	}

	if cr.Spec.ClusterMode {
		result = append(result, corev1.EnvVar{
			Name:  "CLUSTER_MODE",
			Value: "enable",
		})
	}

	return result
}

func (sr *AquaServerHelper) newService(cr *operatorv1alpha1.AquaServer) *corev1.Service {
	labels := map[string]string{
		"app":                cr.Name + "-server",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy the aqua server service",
	}
	selectors := map[string]string{
		"type": "aqua-servers",
	}
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        sr.Parameters.AquaServerServiceName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(sr.Parameters.AquaServerServiceType),
			Selector: selectors,
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	return service
}

func (sr *AquaServerHelper) newServerSecrets(cr *operatorv1alpha1.AquaServer) *corev1.Secret {
	labels := map[string]string{
		"app":                cr.Name + "-server",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy the aqua server secrets",
	}
	websecrets := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        sr.Parameters.AquaServerSecretsName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}

	if len(cr.Spec.AdminPassword) > 0 {
		websecrets.Data[sr.Parameters.AquaAdminPasswordKey] = []byte(cr.Spec.AdminPassword)
	}

	if len(cr.Spec.LicenseToken) > 0 {
		websecrets.Data[sr.Parameters.AquaLicenseTokenKey] = []byte(cr.Spec.LicenseToken)
	}

	return websecrets
}
