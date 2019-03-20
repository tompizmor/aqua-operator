package aquadatabase

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AquaDatabaseParameters struct {
	AquaDbDeploymentName    string
	AquaDbServiceName       string
	AquaDbPvcName           string
	AquaDbInternal          bool
	AquaDbImage             operatorv1alpha1.AquaImage
	AquaServiceAccount      string
	AquaDbSecretName        string
	AquaDbSecretKey         string
	AquaDatabaseServiceType string
}

type AquaDatabaseHelper struct {
	Parameters AquaDatabaseParameters
}

func newAquaDatabaseHelper(cr *operatorv1alpha1.AquaDatabase) *AquaDatabaseHelper {
	params := AquaDatabaseParameters{
		AquaDbDeploymentName:    fmt.Sprintf("%s-database", cr.Name),
		AquaDbServiceName:       fmt.Sprintf("%s-database-svc", cr.Name),
		AquaDbPvcName:           fmt.Sprintf("%s-database-pvc", cr.Name),
		AquaDbInternal:          false,
		AquaDbImage:             *cr.Spec.DbService.ImageData,
		AquaDatabaseServiceType: "ClusterIP",
	}

	if &params.AquaDbImage != nil {
		params.AquaDbInternal = true
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

	if len(cr.Spec.DbService.ServiceType) > 0 {
		params.AquaDatabaseServiceType = cr.Spec.DbService.ServiceType
	}

	return &AquaDatabaseHelper{
		Parameters: params,
	}
}

func (db *AquaDatabaseHelper) newDeployment(cr *operatorv1alpha1.AquaDatabase) *appsv1.Deployment {
	labels := map[string]string{
		"app":                cr.Name + "-database",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy the aqua database server",
	}
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        db.Parameters.AquaDbDeploymentName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: common.Int32Ptr(int32(cr.Spec.DbService.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: db.Parameters.AquaServiceAccount,
					Containers: []corev1.Container{
						{
							Name:            "aqua-db",
							Image:           fmt.Sprintf("%s/%s:%s", db.Parameters.AquaDbImage.Registry, db.Parameters.AquaDbImage.Repository, db.Parameters.AquaDbImage.Tag),
							ImagePullPolicy: corev1.PullPolicy(db.Parameters.AquaDbImage.PullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "postgres-database",
									MountPath: "/var/lib/postgresql/data",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 5432,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/db-files",
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: db.Parameters.AquaDbSecretName,
											},
											Key: db.Parameters.AquaDbSecretKey,
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "postgres-database",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: db.Parameters.AquaDbPvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	if cr.Spec.DbService.Resources != nil {
		deployment.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.DbService.Resources
	}

	if cr.Spec.DbService.LivenessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.DbService.LivenessProbe
	}

	if cr.Spec.DbService.ReadinessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.DbService.ReadinessProbe
	}

	if cr.Spec.DbService.NodeSelector != nil {
		if len(cr.Spec.DbService.NodeSelector) > 0 {
			deployment.Spec.Template.Spec.NodeSelector = cr.Spec.DbService.NodeSelector
		}
	}

	if cr.Spec.DbService.Affinity != nil {
		deployment.Spec.Template.Spec.Affinity = cr.Spec.DbService.Affinity
	}

	if cr.Spec.DbService.Tolerations != nil {
		if len(cr.Spec.DbService.Tolerations) > 0 {
			deployment.Spec.Template.Spec.Tolerations = cr.Spec.DbService.Tolerations
		}
	}

	if cr.Spec.Openshift {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: &cr.Spec.Openshift,
		}
	}

	return deployment
}

func (db *AquaDatabaseHelper) newService(cr *operatorv1alpha1.AquaDatabase) *corev1.Service {
	labels := map[string]string{
		"app":                cr.Name + "-database",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy the aqua database service",
	}
	selectors := map[string]string{
		"app": cr.Name + "-database",
	}
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        db.Parameters.AquaDbServiceName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(db.Parameters.AquaDatabaseServiceType),
			Selector: selectors,
			Ports: []corev1.ServicePort{
				{
					Port: 5432,
				},
			},
		},
	}

	return service
}

func (db *AquaDatabaseHelper) newPersistentVolumeClaim(cr *operatorv1alpha1.AquaDatabase) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app":                cr.Name + "-database",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Persistent Volume Claim for aqua database server",
	}
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        db.Parameters.AquaDbPvcName,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse("30Gi"),
				},
			},
		},
	}

	return pvc
}
