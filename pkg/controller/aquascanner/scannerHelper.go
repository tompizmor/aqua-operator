package aquascanner

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScannerParameters struct {
	AquaScannerDeploymentName     string
	AquaScannerServiceAccountName string
	AquaScannerImage              operatorv1alpha1.AquaImage
}

type AquaScannerHelper struct {
	Parameters ScannerParameters
}

func newAquaScannerHelper(cr *operatorv1alpha1.AquaScanner) *AquaScannerHelper {
	params := ScannerParameters{
		AquaScannerDeploymentName:     fmt.Sprintf("%s-scanner", cr.Name),
		AquaScannerServiceAccountName: fmt.Sprintf("%s-sa", cr.Name),
		AquaScannerImage:              *cr.Spec.ScannerService.ImageData,
	}

	if !cr.Spec.Requirements {
		if len(cr.Spec.ServiceAccountName) > 0 {
			params.AquaScannerServiceAccountName = cr.Spec.ServiceAccountName
		}
	}

	return &AquaScannerHelper{
		Parameters: params,
	}
}

func (as *AquaScannerHelper) newDeployment(cr *operatorv1alpha1.AquaScanner) *appsv1.Deployment {
	labels := map[string]string{
		"app":                cr.Name + "-scanner",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	env_vars := as.getEnvVars(cr)
	annotations := map[string]string{
		"description": "Deploy the aqua scanner",
	}
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        as.Parameters.AquaScannerDeploymentName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: common.Int32Ptr(int32(cr.Spec.ScannerService.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Name:   as.Parameters.AquaScannerDeploymentName,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: as.Parameters.AquaScannerServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "aqua-scanner",
							Image:           fmt.Sprintf("%s/%s:%s", as.Parameters.AquaScannerImage.Registry, as.Parameters.AquaScannerImage.Repository, as.Parameters.AquaScannerImage.Tag),
							ImagePullPolicy: corev1.PullPolicy(as.Parameters.AquaScannerImage.PullPolicy),
							Args: []string{
								"daemon",
								"--user",
								cr.Spec.Login.Username,
								"--password",
								cr.Spec.Login.Password,
								"--host",
								cr.Spec.Login.Host,
							},
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							Env: env_vars,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "docker-socket-mount",
									MountPath: "/var/run/docker.sock",
								},
							},
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

	if cr.Spec.ScannerService.Resources != nil {
		deployment.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.ScannerService.Resources
	}

	if cr.Spec.ScannerService.LivenessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.ScannerService.LivenessProbe
	}

	if cr.Spec.ScannerService.ReadinessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.ScannerService.ReadinessProbe
	}

	if cr.Spec.ScannerService.NodeSelector != nil {
		if len(cr.Spec.ScannerService.NodeSelector) > 0 {
			deployment.Spec.Template.Spec.NodeSelector = cr.Spec.ScannerService.NodeSelector
		}
	}

	if cr.Spec.ScannerService.Affinity != nil {
		deployment.Spec.Template.Spec.Affinity = cr.Spec.ScannerService.Affinity
	}

	if cr.Spec.ScannerService.Tolerations != nil {
		if len(cr.Spec.ScannerService.Tolerations) > 0 {
			deployment.Spec.Template.Spec.Tolerations = cr.Spec.ScannerService.Tolerations
		}
	}

	if cr.Spec.Openshift {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &cr.Spec.Openshift,
		}
	}

	return deployment
}

func (as *AquaScannerHelper) getEnvVars(cr *operatorv1alpha1.AquaScanner) []corev1.EnvVar {

	result := []corev1.EnvVar{
		{
			Name:  "SCANNER_PASSWORD",
			Value: cr.Spec.Login.Password,
		},
	}

	/*if len(cr.Spec.Login.PasswordSecretName) != 0 && len(cr.Spec.Login.PasswordSecretKey) != 0 {
		result = []corev1.EnvVar{
			{
				Name: "SCANNER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Spec.Login.PasswordSecretName,
						},
						Key: cr.Spec.Login.PasswordSecretKey,
					},
				},
			},
		}
	}*/

	return result
}
