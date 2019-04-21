package aquacsp

import (
	"fmt"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CspParameters struct {
	AquaServiceAccount string
	AquaDbSecretName   string
	AquaDbSecretKey    string
	AquaOpenshift      bool
}

type AquaCspHelper struct {
	Parameters CspParameters
}

func newAquaCspHelper(cr *operatorv1alpha1.AquaCsp) *AquaCspHelper {
	params := CspParameters{}
	params.AquaOpenshift = false

	if !cr.Spec.Requirements {
		params.AquaServiceAccount = cr.Spec.ServiceAccountName
		params.AquaDbSecretName = cr.Spec.DbSecretName
		params.AquaDbSecretKey = cr.Spec.DbSecretKey
	} else {
		params.AquaServiceAccount = fmt.Sprintf("%s-sa", cr.Name)
		params.AquaDbSecretName = fmt.Sprintf("%s-database-password", cr.Name)
		params.AquaDbSecretKey = "db-password"
	}

	if cr.Spec.Rbac != nil {
		params.AquaOpenshift = cr.Spec.Rbac.Openshift
	}

	return &AquaCspHelper{
		Parameters: params,
	}
}

func (csp *AquaCspHelper) newAquaDatabase(cr *operatorv1alpha1.AquaCsp) *operatorv1alpha1.AquaDatabase {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Database (not for production environments)",
	}
	aquadb := &operatorv1alpha1.AquaDatabase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaDatabase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: operatorv1alpha1.AquaDatabaseSpec{
			Requirements:       false,
			ServiceAccountName: csp.Parameters.AquaServiceAccount,
			DbSecretName:       csp.Parameters.AquaDbSecretName,
			DbSecretKey:        csp.Parameters.AquaDbSecretKey,
			RegistryData:       cr.Spec.RegistryData,
			DbService:          cr.Spec.DbService,
			Openshift:          csp.Parameters.AquaOpenshift,
		},
	}

	return aquadb
}

func (csp *AquaCspHelper) newAquaGateway(cr *operatorv1alpha1.AquaCsp) *operatorv1alpha1.AquaGateway {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Gateway",
	}
	aquadb := &operatorv1alpha1.AquaGateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaGateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: operatorv1alpha1.AquaGatewaySpec{
			Requirements:       false,
			ServiceAccountName: csp.Parameters.AquaServiceAccount,
			DbSecretName:       csp.Parameters.AquaDbSecretName,
			DbSecretKey:        csp.Parameters.AquaDbSecretKey,
			RegistryData:       cr.Spec.RegistryData,
			GatewayService:     cr.Spec.GatewayService,
			DbDeploymentName:   cr.Name,
			ExternalDb:         cr.Spec.ExternalDb,
			DbSsl:              cr.Spec.DbSsl,
			DbAuditSsl:         cr.Spec.DbAuditSsl,
			Openshift:          csp.Parameters.AquaOpenshift,
		},
	}

	return aquadb
}

func (csp *AquaCspHelper) newAquaServer(cr *operatorv1alpha1.AquaCsp) *operatorv1alpha1.AquaServer {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Server",
	}
	aquadb := &operatorv1alpha1.AquaServer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaServer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: operatorv1alpha1.AquaServerSpec{
			Requirements:       false,
			ServiceAccountName: csp.Parameters.AquaServiceAccount,
			DbSecretName:       csp.Parameters.AquaDbSecretName,
			DbSecretKey:        csp.Parameters.AquaDbSecretKey,
			RegistryData:       cr.Spec.RegistryData,
			ServerService:      cr.Spec.ServerService,
			ExternalDb:         cr.Spec.ExternalDb,
			DbDeploymentName:   cr.Name,
			LicenseToken:       cr.Spec.LicenseToken,
			AdminPassword:      cr.Spec.AdminPassword,
			AquaSslCertPath:    cr.Spec.AquaSslCertPath,
			ClusterMode:        cr.Spec.ClusterMode,
			DbSsl:              cr.Spec.DbSsl,
			DbAuditSsl:         cr.Spec.DbAuditSsl,
			Dockerless:         cr.Spec.Dockerless,
			Openshift:          csp.Parameters.AquaOpenshift,
		},
	}

	return aquadb
}

func (csp *AquaCspHelper) newAquaScanner(cr *operatorv1alpha1.AquaCsp) *operatorv1alpha1.AquaScanner {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Scanner",
	}
	scanner := &operatorv1alpha1.AquaScanner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaScanner",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: operatorv1alpha1.AquaScannerSpec{
			Requirements:       false,
			ServiceAccountName: csp.Parameters.AquaServiceAccount,
			RegistryData:       cr.Spec.RegistryData,
			ScannerService:     cr.Spec.Scanner.Deploy,
			Login: &operatorv1alpha1.AquaLogin{
				Username: "administrator",
				Password: cr.Spec.AdminPassword,
				Host:     fmt.Sprintf("http://%s-server-svc:8080", cr.Name),
			},
			Openshift: csp.Parameters.AquaOpenshift,
		},
	}

	return scanner
}

/*	----------------------------------------------------------------------------------------------------------------
							Aqua CSP RBAC
	----------------------------------------------------------------------------------------------------------------
*/

func (csp *AquaCspHelper) newPodSecurityPolicy(cr *operatorv1alpha1.AquaCsp) *v1beta1.PodSecurityPolicy {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Pod Security Policy",
	}
	psp := &v1beta1.PodSecurityPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaServer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-psp", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1beta1.PodSecurityPolicySpec{
			Privileged: true,
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

func (csp *AquaCspHelper) newClusterRole(cr *operatorv1alpha1.AquaCsp) *rbacv1.ClusterRole {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Cluster Role",
	}
	crole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaServer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-cluster-role", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"extensions",
				},
				ResourceNames: []string{
					fmt.Sprintf("%s-psp", cr.Name),
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

	return crole
}

func (csp *AquaCspHelper) newClusterRoleBinding(cr *operatorv1alpha1.AquaCsp) *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app":                cr.Name + "-csp",
		"deployedby":         "aqua-operator",
		"aquasecoperator_cr": cr.Name,
	}
	annotations := map[string]string{
		"description": "Deploy Aqua Cluster Role Binding",
	}
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.aquasec.com/v1alpha1",
			Kind:       "AquaServer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-cluster-role", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      csp.Parameters.AquaServiceAccount,
				Namespace: cr.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     fmt.Sprintf("%s-cluster-role", cr.Name),
		},
	}

	if cr.Spec.Rbac != nil {
		if len(cr.Spec.Rbac.RoleRef) > 0 {
			crb.RoleRef.Name = cr.Spec.Rbac.RoleRef
		}
	}

	return crb
}
