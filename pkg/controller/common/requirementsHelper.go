package common

import (
	"encoding/json"
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequirementsParameters struct {
	AquaPullImageSecretName string
	AquaDbSecretName        string
	AquaServiceAccountName  string
	AquaRegistry            operatorv1alpha1.AquaDockerRegistry
	AquaDbPassword          string
}

type AquaRequirementsHelper struct {
	Parameters RequirementsParameters
}

func NewAquaRequirementsHelper(dr *operatorv1alpha1.AquaDockerRegistry, name string, dbpassword string) *AquaRequirementsHelper {
	params := RequirementsParameters{
		AquaPullImageSecretName: fmt.Sprintf("%s-registry-secret", name),
		AquaDbSecretName:        fmt.Sprintf("%s-database-password", name),
		AquaServiceAccountName:  fmt.Sprintf("%s-sa", name),
		AquaRegistry:            *dr,
		AquaDbPassword:          dbpassword,
	}

	return &AquaRequirementsHelper{
		Parameters: params,
	}
}

func (rq *AquaRequirementsHelper) NewServiceAccount(name string, namespace string) *corev1.ServiceAccount {
	labels := map[string]string{
		"app":                name + "-requirments",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": name,
	}
	annotations := map[string]string{
		"description": "Service account for pulling aqua images",
	}
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        rq.Parameters.AquaServiceAccountName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			corev1.LocalObjectReference{
				Name: rq.Parameters.AquaPullImageSecretName,
			},
		},
	}

	return sa
}

func (rq *AquaRequirementsHelper) NewImagePullSecret(name string, namespace string) *corev1.Secret {
	labels := map[string]string{
		"app":                name + "-requirments",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": name,
	}
	annotations := map[string]string{
		"description": "Secret for pulling aqua images",
	}
	auth := map[string]interface{}{
		"auths": map[string]interface{}{
			rq.Parameters.AquaRegistry.Url: map[string]interface{}{
				"username": rq.Parameters.AquaRegistry.Username,
				"password": rq.Parameters.AquaRegistry.Password,
				"email":    rq.Parameters.AquaRegistry.Email,
			},
		},
	}

	authBytes, _ := json.Marshal(auth)

	imageps := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        rq.Parameters.AquaPullImageSecretName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: authBytes,
		},
	}

	return imageps
}

func (rq *AquaRequirementsHelper) NewDbPasswordSecret(name string, namespace string) *corev1.Secret {
	labels := map[string]string{
		"app":                name + "-requirments",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": name,
	}
	annotations := map[string]string{
		"description": "Secret for aqua database password",
	}
	imageps := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        rq.Parameters.AquaDbSecretName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"db-password": []byte(rq.Parameters.AquaDbPassword),
		},
	}

	return imageps
}
