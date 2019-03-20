package common

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RbacParameters struct {
	Name                   string
	Namespace              string
	AquaServiceAccountName string
	Requirement            bool
	Openshift              bool
	Rbac                   operatorv1alpha1.AquaRbacSettings
	Privileged             bool
}

type AquaRbacHelper struct {
	Parameters RbacParameters
}

func NewAquaRbacHelper(requirments bool, serviceaccount string, rbac operatorv1alpha1.AquaRbacSettings, name string, namespace string, openshift bool) *AquaRbacHelper {
	params := RbacParameters{
		Name:                   name,
		Namespace:              namespace,
		AquaServiceAccountName: fmt.Sprintf("%s-sa", name),
		Requirement:            requirments,
		Rbac:                   rbac,
		Openshift:              openshift,
		Privileged:             true,
	}

	if !requirments {
		if len(serviceaccount) > 0 {
			params.AquaServiceAccountName = serviceaccount
		}
	}

	if &params.Rbac != nil {
		params.Privileged = params.Rbac.Privileged
	}

	return &AquaRbacHelper{
		Parameters: params,
	}
}

/*	----------------------------------------------------------------------------------------------------------------
							Aqua RBAC
	----------------------------------------------------------------------------------------------------------------
*/

func (rb *AquaRbacHelper) NewPodSecurityPolicy() *v1beta1.PodSecurityPolicy {
	labels := map[string]string{
		"app":                rb.Parameters.Name + "-rbac",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": rb.Parameters.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Pod Security Policy",
	}
	psp := &v1beta1.PodSecurityPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "PodSecurityPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-psp", rb.Parameters.Name),
			Namespace:   rb.Parameters.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1beta1.PodSecurityPolicySpec{
			Privileged: rb.Parameters.Privileged,
			AllowedCapabilities: []corev1.Capability{
				"*",
			},
			FSGroup: v1beta1.FSGroupStrategyOptions{
				Rule: v1beta1.FSGroupStrategyRunAsAny,
			},
			RunAsUser: v1beta1.RunAsUserStrategyOptions{
				Rule: v1beta1.RunAsUserStrategyRunAsAny,
			},
			SELinux: v1beta1.SELinuxStrategyOptions{
				Rule: v1beta1.SELinuxStrategyRunAsAny,
			},
			SupplementalGroups: v1beta1.SupplementalGroupsStrategyOptions{
				Rule: v1beta1.SupplementalGroupsStrategyRunAsAny,
			},
			Volumes: []v1beta1.FSType{
				v1beta1.All,
			},
		},
	}

	return psp
}

func (rb *AquaRbacHelper) NewClusterRole() *rbacv1.ClusterRole {
	labels := map[string]string{
		"app":                rb.Parameters.Name + "-rbac",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": rb.Parameters.Name,
	}
	annotations := map[string]string{
		"description":              "Deploy Aqua Cluster Role",
		"openshift.io/description": "A user who can search and scan images from an OpenShift integrated registry.",
	}
	crole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-cluster-role", rb.Parameters.Name),
			Namespace:   rb.Parameters.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"extensions",
				},
				Resources: []string{
					"podsecuritypolicies",
				},
				Verbs: []string{
					"use",
				},
			},
		},
	}

	if rb.Parameters.Openshift {
		role := rbacv1.PolicyRule{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"imagestreams",
				"imagestreams/layers",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		}
		crole.Rules = append(crole.Rules, role)
	}

	return crole
}

func (rb *AquaRbacHelper) NewClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app":                rb.Parameters.Name + "-rbac",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": rb.Parameters.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Cluster Role Binding",
	}
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-role-binding", rb.Parameters.Name),
			Namespace:   rb.Parameters.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      rb.Parameters.AquaServiceAccountName,
				Namespace: rb.Parameters.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     fmt.Sprintf("%s-cluster-role", rb.Parameters.Name),
		},
	}

	if &rb.Parameters.Rbac != nil {
		if len(rb.Parameters.Rbac.RoleRef) > 0 {
			crb.RoleRef.Name = rb.Parameters.Rbac.RoleRef
		}
	}

	return crb
}
