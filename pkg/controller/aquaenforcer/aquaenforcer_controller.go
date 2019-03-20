package aquaenforcer

import (
	"context"
	"reflect"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_aquaenforcer")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AquaEnforcer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAquaEnforcer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("aquaenforcer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AquaEnforcer
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AquaEnforcer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AquaEnforcer
	// Requirments
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	// AquaEnforcer Components

	err = c.Watch(&source.Kind{Type: &v1beta1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	// RBAC

	err = c.Watch(&source.Kind{Type: &v1beta1.PodSecurityPolicy{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRole{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaEnforcer{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAquaEnforcer{}

// ReconcileAquaEnforcer reconciles a AquaEnforcer object
type ReconcileAquaEnforcer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AquaEnforcer object and makes changes based on the state read
// and what is in the AquaEnforcer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAquaEnforcer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AquaEnforcer")

	// Fetch the AquaEnforcer instance
	instance := &operatorv1alpha1.AquaEnforcer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Spec.Requirements {
		reqLogger.Info("Start Setup Requirment For Aqua Enforcer")
		_, err = r.CreateImagePullSecret(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		_, err = r.CreateAquaServiceAccount(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if instance.Spec.EnforcerService != nil {
		if instance.Spec.EnforcerService.ImageData != nil {
			_, err = r.InstallEnforcerToken(instance)
			if err != nil {
				return reconcile.Result{}, err
			}

			_, err = r.InstallEnforcerDaemonSet(instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	if instance.Spec.Rbac != nil {
		if instance.Spec.Rbac.Enable {
			if !(len(instance.Spec.Rbac.RoleRef) > 0) {
				_, err = r.CreatePodSecurityPolicy(instance)
				if err != nil {
					return reconcile.Result{}, err
				}

				_, err = r.CreateClusterRole(instance)
				if err != nil {
					return reconcile.Result{}, err
				}
			}

			_, err = r.CreateClusterRoleBinding(instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	if !reflect.DeepEqual(operatorv1alpha1.AquaDeploymentStateRunning, instance.Status.State) {
		instance.Status.State = operatorv1alpha1.AquaDeploymentStateRunning
		_ = r.client.Update(context.TODO(), instance)
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaEnforcer) InstallEnforcerDaemonSet(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("Aqua Enforcer DaemonSet Phase", "Install Aqua Enforcer DaemonSet")
	reqLogger.Info("Start installing enforcer")

	// Define a new DaemonSet object
	enforcerHelper := newAquaEnforcerHelper(cr)
	ds := enforcerHelper.CreateDaemonSet(cr)

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, ds, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DaemonSet already exists
	found := &v1beta1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: ds.Name, Namespace: ds.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		err = r.client.Create(context.TODO(), ds)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// DaemonSet already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Enforcer DaemonSet Already Exists", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaEnforcer) InstallEnforcerToken(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("Aqua Enforcer Phase", "Create Aqua Enforcer Token Secret")
	reqLogger.Info("Start creating enforcer token secret")

	// Define a new DaemonSet object
	enforcerHelper := newAquaEnforcerHelper(cr)
	token := enforcerHelper.CreateTokenSecret(cr)

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, token, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DaemonSet already exists
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: token.Name, Namespace: token.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "Secret.Namespace", token.Namespace, "Secret.Name", token.Name)
		err = r.client.Create(context.TODO(), token)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Enforcer Token Secret Already Exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

/*	----------------------------------------------------------------------------------------------------------------
							Requirments
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaEnforcer) CreateImagePullSecret(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("Enforcer Requirments Phase", "Create Image Pull Secret")
	reqLogger.Info("Start creating aqua images pull secret")

	// Define a new secret object
	requirementsHelper := common.NewAquaRequirementsHelper(cr.Spec.RegistryData, cr.Name, "")
	secret := requirementsHelper.NewImagePullSecret(cr.Name, cr.Namespace)

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this secret already exists
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Image Pull Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Image Pull Secret Already Exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaEnforcer) CreateAquaServiceAccount(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("Enforcer Requirments Phase", "Create Aqua Service Account")
	reqLogger.Info("Start creating aqua service account")

	// Define a new service account object
	requirementsHelper := common.NewAquaRequirementsHelper(cr.Spec.RegistryData, cr.Name, "")
	sa := requirementsHelper.NewServiceAccount(cr.Name, cr.Namespace)

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, sa, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this service account already exists
	found := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Service Account", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
		err = r.client.Create(context.TODO(), sa)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Service account already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Service Account Already Exists", "ServiceAccount.Namespace", found.Namespace, "ServiceAccount.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

/*	----------------------------------------------------------------------------------------------------------------
							RBAC
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaEnforcer) CreatePodSecurityPolicy(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("AquaEnforcer - RBAC Phase", "Create PodSecurityPolicy")
	reqLogger.Info("Start creating PodSecurityPolicy")

	// Define a new PodSecurityPolicy object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	psp := rbacHelper.NewPodSecurityPolicy()

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, psp, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PodSecurityPolicy already exists
	found := &v1beta1.PodSecurityPolicy{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: psp.Name, Namespace: psp.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "PodSecurityPolicy.Namespace", psp.Namespace, "PodSecurityPolicy.Name", psp.Name)
		err = r.client.Create(context.TODO(), psp)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// PodSecurityPolicy already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua PodSecurityPolicy Exists", "PodSecurityPolicy.Namespace", found.Namespace, "PodSecurityPolicy.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaEnforcer) CreateClusterRole(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("AquaEnforcer - RBAC Phase", "Create ClusterRole")
	reqLogger.Info("Start creating ClusterRole")

	// Define a new ClusterRole object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	crole := rbacHelper.NewClusterRole()

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, crole, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRole already exists
	found := &rbacv1.ClusterRole{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crole.Name, Namespace: crole.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "ClusterRole.Namespace", crole.Namespace, "ClusterRole.Name", crole.Name)
		err = r.client.Create(context.TODO(), crole)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// ClusterRole already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua ClusterRole Exists", "ClusterRole.Namespace", found.Namespace, "ClusterRole.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaEnforcer) CreateClusterRoleBinding(cr *operatorv1alpha1.AquaEnforcer) (reconcile.Result, error) {
	reqLogger := log.WithValues("AquaEnforcer - RBAC Phase", "Create ClusterRoleBinding")
	reqLogger.Info("Start creating ClusterRole")

	// Define a new ClusterRoleBinding object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	crb := rbacHelper.NewClusterRoleBinding()

	// Set AquaEnforcer instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, crb, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRoleBinding already exists
	found := &rbacv1.ClusterRoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name, Namespace: crb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "ClusterRoleBinding.Namespace", crb.Namespace, "ClusterRoleBinding.Name", crb.Name)
		err = r.client.Create(context.TODO(), crb)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// ClusterRoleBinding already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua ClusterRoleBinding Exists", "ClusterRoleBinding.Namespace", found.Namespace, "ClusterRole.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}
