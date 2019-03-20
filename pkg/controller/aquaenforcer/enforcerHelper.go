package aquaenforcer

import (
	"fmt"
	"strconv"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnforcerParameters :
type EnforcerParameters struct {
	AquaEnforcerDaemonSetName      string
	AquaEnforcerTokenSecretName    string
	AquaEnforcerServiceAccountName string
	AquaEnforcerImage              operatorv1alpha1.AquaImage
	Privileged                     bool
	RuncInterception               string
}

// AquaEnforcerHelper :
type AquaEnforcerHelper struct {
	Parameters EnforcerParameters
}

func newAquaEnforcerHelper(cr *operatorv1alpha1.AquaEnforcer) *AquaEnforcerHelper {
	params := EnforcerParameters{
		AquaEnforcerDaemonSetName:      fmt.Sprintf("%s-ds", cr.Name),
		AquaEnforcerTokenSecretName:    fmt.Sprintf("%s-token", cr.Name),
		AquaEnforcerImage:              *cr.Spec.EnforcerService.ImageData,
		AquaEnforcerServiceAccountName: fmt.Sprintf("%s-sa", cr.Name),
		Privileged:                     true,
		RuncInterception:               "0",
	}

	if !cr.Spec.Requirements {
		if len(cr.Spec.ServiceAccountName) > 0 {
			params.AquaEnforcerServiceAccountName = cr.Spec.ServiceAccountName
		}
	}

	if cr.Spec.Rbac != nil {
		params.Privileged = cr.Spec.Rbac.Privileged
	}

	if cr.Spec.RuncInterception {
		params.RuncInterception = "1"
	}

	return &AquaEnforcerHelper{
		Parameters: params,
	}
}

// CreateTokenSecret : Create Enforcer Token Secret For The Enforcer connection to the aqua csp environment
func (enf *AquaEnforcerHelper) CreateTokenSecret(cr *operatorv1alpha1.AquaEnforcer) *corev1.Secret {
	labels := map[string]string{
		"app":                cr.Name + "-requirments",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Secret for aqua database password",
	}
	token := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        enf.Parameters.AquaEnforcerTokenSecretName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(cr.Spec.Token),
		},
	}

	return token
}

// CreateDaemonSet :
func (enf *AquaEnforcerHelper) CreateDaemonSet(cr *operatorv1alpha1.AquaEnforcer) *v1beta1.DaemonSet {
	labels := map[string]string{
		"app":                cr.Name + "-requirments",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Secret for aqua database password",
	}
	ds := &v1beta1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        enf.Parameters.AquaEnforcerDaemonSetName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Name:   enf.Parameters.AquaEnforcerDaemonSetName,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: enf.Parameters.AquaEnforcerServiceAccountName,
					HostPID:            true,
					Containers: []corev1.Container{
						{
							Name:            "aqua-enforcer",
							Image:           fmt.Sprintf("%s/%s:%s", enf.Parameters.AquaEnforcerImage.Registry, enf.Parameters.AquaEnforcerImage.Repository, enf.Parameters.AquaEnforcerImage.Tag),
							ImagePullPolicy: corev1.PullPolicy(enf.Parameters.AquaEnforcerImage.PullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "var-run",
									MountPath: "/var/run",
								},
								{
									Name:      "dev",
									MountPath: "/dev",
								},
								{
									Name:      "sys",
									MountPath: "/host/sys",
									ReadOnly:  true,
								},
								{
									Name:      "proc",
									MountPath: "/host/proc",
									ReadOnly:  true,
								},
								{
									Name:      "etc",
									MountPath: "/host/etc",
									ReadOnly:  true,
								},
								{
									Name:      "aquasec",
									MountPath: "/host/opt/aquasec",
									ReadOnly:  true,
								},
								{
									Name:      "aquasec-tmp",
									MountPath: "/opt/aquasec/tmp",
								},
								{
									Name:      "aquasec-audit",
									MountPath: "/opt/aquasec/audit",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "AQUA_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: enf.Parameters.AquaEnforcerTokenSecretName,
											},
											Key: "token",
										},
									},
								},
								{
									Name:  "AQUA_SERVER",
									Value: fmt.Sprintf("%s:%d", cr.Spec.Gateway.Host, cr.Spec.Gateway.Port),
								},
								{
									Name:  "AQUA_INSTALL_PATH",
									Value: "/var/lib/aquasec",
								},
								{
									Name:  "AQUA_LOGICAL_NAME",
									Value: fmt.Sprintf("%s-operator", cr.Name),
								},
								{
									Name:  "AQUA_NETWORK_CONTROL",
									Value: "1",
								},
								{
									Name:  "RESTART_CONTAINERS",
									Value: "no",
								},
								{
									Name:  "SENDING_HOST_IMAGES_DISABLED",
									Value: strconv.FormatBool(cr.Spec.SendingHostImages),
								},
								{
									Name:  "AQUA_RUNC_INTERCEPTION",
									Value: enf.Parameters.RuncInterception,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "var-run",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run",
								},
							},
						},
						{
							Name: "dev",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev",
								},
							},
						},
						{
							Name: "sys",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys",
								},
							},
						},
						{
							Name: "proc",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/proc",
								},
							},
						},
						{
							Name: "etc",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc",
								},
							},
						},
						{
							Name: "aquasec",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/aquasec",
								},
							},
						},
						{
							Name: "aquasec-tmp",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/aquasec/tmp",
								},
							},
						},
						{
							Name: "aquasec-audit",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/aquasec/audit",
								},
							},
						},
					},
				},
			},
		},
	}

	if cr.Spec.EnforcerService.Resources != nil {
		ds.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.EnforcerService.Resources
	}

	if cr.Spec.EnforcerService.LivenessProbe != nil {
		ds.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.EnforcerService.LivenessProbe
	}

	if cr.Spec.EnforcerService.ReadinessProbe != nil {
		ds.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.EnforcerService.ReadinessProbe
	}

	if cr.Spec.EnforcerService.NodeSelector != nil {
		if len(cr.Spec.EnforcerService.NodeSelector) > 0 {
			ds.Spec.Template.Spec.NodeSelector = cr.Spec.EnforcerService.NodeSelector
		}
	}

	if enf.Parameters.Privileged {
		ds.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &enf.Parameters.Privileged,
		}
	} else {
		ds.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &enf.Parameters.Privileged,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"SYS_ADMIN",
					"NET_ADMIN",
					"NET_RAW",
					"SYS_PTRACE",
					"KILL",
					"MKNOD",
					"SETGID",
					"SETUID",
					"SYS_MODULE",
					"AUDIT_CONTROL",
					"SYSLOG",
					"SYS_CHROOT",
				},
			},
		}
	}

	return ds
}
