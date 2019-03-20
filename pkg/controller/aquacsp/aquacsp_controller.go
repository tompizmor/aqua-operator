package aquacsp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	syserrors "errors"

	operatorv1alpha1 "github.com/niso120b/aqua-operator/pkg/apis/operator/v1alpha1"
	"github.com/niso120b/aqua-operator/pkg/controller/common"
	appsv1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
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

var log = logf.Log.WithName("controller_aquacsp")

const (
	MinScanners         int64 = 1
	MaxScanners         int64 = 5
	MaxImagesPerScanner int64 = 250
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AquaCsp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAquaCsp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("aquacsp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AquaCsp
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AquaCsp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AquaCsp
	// Requirments
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	// AquaCsp Components

	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AquaDatabase{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AquaGateway{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AquaServer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	// RBAC

	err = c.Watch(&source.Kind{Type: &v1beta1.PodSecurityPolicy{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRole{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.AquaCsp{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAquaCsp{}

// ReconcileAquaCsp reconciles a AquaCsp object
type ReconcileAquaCsp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AquaCsp object and makes changes based on the state read
// and what is in the AquaCsp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAquaCsp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AquaCsp")

	// Fetch the AquaCsp instance
	instance := &operatorv1alpha1.AquaCsp{}
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
		reqLogger.Info("Start Setup Requirment For Aqua CSP")
		_, err = r.CreateImagePullSecret(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		_, err = r.CreateAquaServiceAccount(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if len(instance.Spec.Password) > 0 {
			reqLogger.Info("Start Setup Secret For Database Password")
			_, err = r.CreateDbPasswordSecret(instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	dbstatus := true
	if instance.Spec.DbService != nil {
		reqLogger.Info("CSP Deployment: Start Setup Internal Aqua Database (Not For Production Usage)")
		_, err = r.InstallAquaDatabase(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		dbstatus, _ = r.WaitForDatabase(instance)
	}

	if dbstatus {
		if instance.Spec.GatewayService == nil {
			reqLogger.Error(syserrors.New("Missing Aqua Gateway Deployment Data!, Please fix and redeploy template!"), "Aqua CSP Deployment Missing Gateway Deployment Data!")
		}

		if instance.Spec.ServerService == nil {
			reqLogger.Error(syserrors.New("Missing Aqua Server Deployment Data!, Please fix and redeploy template!"), "Aqua CSP Deployment Missing Server Deployment Data!")
		}

		_, err = r.InstallAquaGateway(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		_, err = r.InstallAquaServer(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		gwstatus, _ := r.WaitForGateway(instance)
		srstatus, _ := r.WaitForServer(instance)

		if !gwstatus || !srstatus {
			reqLogger.Info("CSP Deployment: Waiting internal for aqua to start")
			if !reflect.DeepEqual(operatorv1alpha1.AquaDeploymentStateWaitingAqua, instance.Status.State) {
				instance.Status.State = operatorv1alpha1.AquaDeploymentStateWaitingAqua
				_ = r.client.Update(context.TODO(), instance)
			}
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
		}
	} else {
		reqLogger.Info("CSP Deployment: Waiting internal for database to start")
		if !reflect.DeepEqual(operatorv1alpha1.AquaDeploymentStateWaitingDB, instance.Status.State) {
			instance.Status.State = operatorv1alpha1.AquaDeploymentStateWaitingDB
			_ = r.client.Update(context.TODO(), instance)
		}
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
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

	if instance.Spec.Scanner != nil {
		if instance.Spec.Scanner.Deploy != nil {
			_, err = r.InstallAquaScanner(instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		_, _ = r.ScaleScannerCLI(instance)
	}

	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
}

/*	----------------------------------------------------------------------------------------------------------------
							Aqua CSP
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaCsp) InstallAquaDatabase(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - AquaDatabase Phase", "Install Aqua Database")
	reqLogger.Info("Start installing AquaDatabase")

	// Define a new AquaDatabase object
	cspHelper := newAquaCspHelper(cr)
	aquadb := cspHelper.newAquaDatabase(cr)

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, aquadb, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this AquaDatabase already exists
	found := &operatorv1alpha1.AquaDatabase{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: aquadb.Name, Namespace: aquadb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Database", "AquaDatabase.Namespace", aquadb.Namespace, "AquaDatabase.Name", aquadb.Name)
		err = r.client.Create(context.TODO(), aquadb)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if found != nil {
		size := aquadb.Spec.DbService.Replicas
		if found.Spec.DbService.Replicas != size {
			found.Spec.DbService.Replicas = size
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Aqua CSP: Failed to update aqua database replicas.", "AquaDatabase.Namespace", found.Namespace, "AquaDatabase.Name", found.Name)
				return reconcile.Result{}, err
			}
			// Spec updated - return and requeue
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// AquaDatabase already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Database Exists", "AquaDatabase.Namespace", found.Namespace, "AquaDatabase.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) InstallAquaGateway(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - AquaGateway Phase", "Install Aqua Database")
	reqLogger.Info("Start installing AquaGateway")

	// Define a new AquaGateway object
	cspHelper := newAquaCspHelper(cr)
	aquagw := cspHelper.newAquaGateway(cr)

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, aquagw, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this AquaGateway already exists
	found := &operatorv1alpha1.AquaGateway{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: aquagw.Name, Namespace: aquagw.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Gateway", "AquaGateway.Namespace", aquagw.Namespace, "AquaGateway.Name", aquagw.Name)
		err = r.client.Create(context.TODO(), aquagw)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if found != nil {
		size := aquagw.Spec.GatewayService.Replicas
		if found.Spec.GatewayService.Replicas != size {
			found.Spec.GatewayService.Replicas = size
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Aqua CSP: Failed to update aqua gateway replicas.", "AquaServer.Namespace", found.Namespace, "AquaServer.Name", found.Name)
				return reconcile.Result{}, err
			}
			// Spec updated - return and requeue
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// AquaGateway already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Gateway Exists", "AquaGateway.Namespace", found.Namespace, "AquaGateway.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) InstallAquaServer(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - AquaServer Phase", "Install Aqua Database")
	reqLogger.Info("Start installing AquaServer")

	// Define a new AquaServer object
	cspHelper := newAquaCspHelper(cr)
	aquasr := cspHelper.newAquaServer(cr)

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, aquasr, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this AquaServer already exists
	found := &operatorv1alpha1.AquaServer{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: aquasr.Name, Namespace: aquasr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua AquaServer", "AquaServer.Namespace", aquasr.Namespace, "AquaServer.Name", aquasr.Name)
		err = r.client.Create(context.TODO(), aquasr)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if found != nil {
		size := aquasr.Spec.ServerService.Replicas
		if found.Spec.ServerService.Replicas != size {
			found.Spec.ServerService.Replicas = size
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Aqua CSP: Failed to update aqua server replicas.", "AquaServer.Namespace", found.Namespace, "AquaServer.Name", found.Name)
				return reconcile.Result{}, err
			}
			// Spec updated - return and requeue
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// AquaServer already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Server Exists", "AquaServer.Namespace", found.Namespace, "AquaServer.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

// Aqua Scanner Optional

func (r *ReconcileAquaCsp) InstallAquaScanner(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - AquaScanner Phase", "Install Aqua Scanner")
	reqLogger.Info("Start installing AquaScanner")

	// Define a new AquaScanner object
	cspHelper := newAquaCspHelper(cr)
	scanner := cspHelper.newAquaScanner(cr)

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, scanner, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this AquaScanner already exists
	found := &operatorv1alpha1.AquaScanner{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: scanner.Name, Namespace: scanner.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Scanner", "AquaScanner.Namespace", scanner.Namespace, "AquaScanner.Name", scanner.Name)
		err = r.client.Create(context.TODO(), scanner)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if found != nil {
		size := scanner.Spec.ScannerService.Replicas
		if found.Spec.ScannerService.Replicas != size {
			found.Spec.ScannerService.Replicas = size
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Aqua CSP: Failed to update aqua scanner replicas.", "AquaScanner.Namespace", found.Namespace, "AquaScanner.Name", found.Name)
				return reconcile.Result{}, err
			}
			// Spec updated - return and requeue
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// AquaScanner already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Scanner Exists", "AquaScanner.Namespace", found.Namespace, "AquaScanner.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

/*	----------------------------------------------------------------------------------------------------------------
							Requirments
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaCsp) CreateImagePullSecret(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("Csp Requirments Phase", "Create Image Pull Secret")
	reqLogger.Info("Start creating aqua images pull secret")

	// Define a new secret object
	requirementsHelper := common.NewAquaRequirementsHelper(cr.Spec.RegistryData, cr.Name, cr.Spec.Password)
	secret := requirementsHelper.NewImagePullSecret(cr.Name, cr.Namespace)

	// Set AquaCspKind instance as the owner and controller
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

func (r *ReconcileAquaCsp) CreateDbPasswordSecret(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("Csp Requirments Phase", "Create Db Password Secret")
	reqLogger.Info("Start creating aqua db password secret")

	// Define a new secret object
	requirementsHelper := common.NewAquaRequirementsHelper(cr.Spec.RegistryData, cr.Name, cr.Spec.Password)
	secret := requirementsHelper.NewDbPasswordSecret(cr.Name, cr.Namespace)

	// Set AquaCspKind instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this secret already exists
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a New Aqua Db Password Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua Db Password Secret Already Exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) CreateAquaServiceAccount(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("Csp Requirments Phase", "Create Aqua Service Account")
	reqLogger.Info("Start creating aqua service account")

	// Define a new service account object
	requirementsHelper := common.NewAquaRequirementsHelper(cr.Spec.RegistryData, cr.Name, cr.Spec.Password)
	sa := requirementsHelper.NewServiceAccount(cr.Name, cr.Namespace)

	// Set AquaCspKind instance as the owner and controller
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
							Check Functions - Internal Only
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaCsp) WaitForDatabase(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	reqLogger := log.WithValues("Csp Wait For Database Phase", "Wait For Database")
	reqLogger.Info("Start waiting to aqua database")

	ready, err := r.GetPostgresReady(cr)
	if err != nil {
		return false, err
	}

	if !ready {
		return false, nil
	}

	return true, nil
}

func (r *ReconcileAquaCsp) GetPostgresReady(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	resource := appsv1.Deployment{}

	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Name + "-database",
	}

	err := r.client.Get(context.TODO(), selector, &resource)
	if err != nil {
		return false, err
	}

	return resource.Status.ReadyReplicas == 1, nil
}

func (r *ReconcileAquaCsp) WaitForGateway(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	reqLogger := log.WithValues("Csp Wait For Aqua Gateway Phase", "Wait For Aqua Gateway")
	reqLogger.Info("Start waiting to aqua gateway")

	ready, err := r.GetGatewayReady(cr)
	if err != nil {
		return false, err
	}

	if !ready {
		return false, nil
	}

	return true, nil
}

func (r *ReconcileAquaCsp) GetGatewayReady(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	resource := appsv1.Deployment{}

	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Name + "-gateway",
	}

	err := r.client.Get(context.TODO(), selector, &resource)
	if err != nil {
		return false, err
	}

	return resource.Status.ReadyReplicas == 1, nil
}

func (r *ReconcileAquaCsp) WaitForServer(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	reqLogger := log.WithValues("Csp Wait For Aqua Server Phase", "Wait For Aqua Server")
	reqLogger.Info("Start waiting to aqua server")

	ready, err := r.GetServerReady(cr)
	if err != nil {
		return false, err
	}

	if !ready {
		return false, nil
	}

	return true, nil
}

func (r *ReconcileAquaCsp) GetServerReady(cr *operatorv1alpha1.AquaCsp) (bool, error) {
	resource := appsv1.Deployment{}

	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Name + "-server",
	}

	err := r.client.Get(context.TODO(), selector, &resource)
	if err != nil {
		return false, err
	}

	return resource.Status.ReadyReplicas == 1, nil
}

/*	----------------------------------------------------------------------------------------------------------------
							RBAC
	----------------------------------------------------------------------------------------------------------------
*/

func (r *ReconcileAquaCsp) CreatePodSecurityPolicy(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - RBAC Phase", "Create PodSecurityPolicy")
	reqLogger.Info("Start creating PodSecurityPolicy")

	// Define a new PodSecurityPolicy object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	psp := rbacHelper.NewPodSecurityPolicy()

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, psp, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PodSecurityPolicy already exists
	found := &v1beta1.PodSecurityPolicy{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: psp.Name, Namespace: psp.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Aqua CSP: Creating a New PodSecurityPolicy", "PodSecurityPolicy.Namespace", psp.Namespace, "PodSecurityPolicy.Name", psp.Name)
		err = r.client.Create(context.TODO(), psp)
		if err != nil {
			return reconcile.Result{Requeue: true}, nil // TODO: reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// PodSecurityPolicy already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua PodSecurityPolicy Already Exists", "PodSecurityPolicy.Namespace", found.Namespace, "PodSecurityPolicy.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) CreateClusterRole(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - RBAC Phase", "Create ClusterRole")
	reqLogger.Info("Start creating ClusterRole")

	// Define a new ClusterRole object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	crole := rbacHelper.NewClusterRole()

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, crole, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRole already exists
	found := &rbacv1.ClusterRole{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crole.Name, Namespace: crole.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Aqua CSP: Creating a New ClusterRole", "ClusterRole.Namespace", crole.Namespace, "ClusterRole.Name", crole.Name)
		err = r.client.Create(context.TODO(), crole)
		if err != nil {
			return reconcile.Result{Requeue: true}, nil // TODO: reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// ClusterRole already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua ClusterRole Exists", "ClusterRole.Namespace", found.Namespace, "ClusterRole.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) CreateClusterRoleBinding(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	reqLogger := log.WithValues("CSP - RBAC Phase", "Create ClusterRoleBinding")
	reqLogger.Info("Start creating ClusterRole")

	// Define a new ClusterRoleBinding object
	rbacHelper := common.NewAquaRbacHelper(cr.Spec.Requirements, cr.Spec.ServiceAccountName, *cr.Spec.Rbac, cr.Name, cr.Namespace, cr.Spec.Rbac.Openshift)
	crb := rbacHelper.NewClusterRoleBinding()

	// Set AquaCsp instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, crb, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRoleBinding already exists
	found := &rbacv1.ClusterRoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name, Namespace: crb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Aqua CSP: Creating a New ClusterRoleBinding", "ClusterRoleBinding.Namespace", crb.Namespace, "ClusterRoleBinding.Name", crb.Name)
		err = r.client.Create(context.TODO(), crb)
		if err != nil {
			return reconcile.Result{Requeue: true}, nil // TODO: reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// ClusterRoleBinding already exists - don't requeue
	reqLogger.Info("Skip reconcile: Aqua ClusterRoleBinding Exists", "ClusterRoleBinding.Namespace", found.Namespace, "ClusterRole.Name", found.Name)
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileAquaCsp) ScaleScannerCLI(cr *operatorv1alpha1.AquaCsp) (reconcile.Result, error) {
	max := MaxScanners
	min := MinScanners
	name := cr.Name
	scannersPerImages := MaxImagesPerScanner

	if cr.Spec.Scanner.Max != 0 {
		max = cr.Spec.Scanner.Max
	}

	if cr.Spec.Scanner.Min != 0 {
		min = cr.Spec.Scanner.Min
	}

	if cr.Spec.Scanner.ImagesPerScanner != 0 {
		scannersPerImages = cr.Spec.Scanner.ImagesPerScanner
	}

	if len(cr.Spec.Scanner.Name) > 0 {
		name = cr.Spec.Scanner.Name
	}

	reqLogger := log.WithValues("CSP - Scale", "Scale Aqua Scanner CLI")
	reqLogger.Info("Start get scanner cli data")

	result, err := common.GetPendingScanQueue("administrator", cr.Spec.AdminPassword, fmt.Sprintf("%s-server-svc", cr.Name))
	if err != nil {
		reqLogger.Info("Waiting for aqua server to be up...")
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, err
	}

	reqLogger.Info("Count of pending scan queue", "Pending Scan Queue", result.Count)

	if result.Count == 0 {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
	}

	nodes := &corev1.NodeList{}
	count := int64(0)
	err = r.client.List(context.TODO(), &client.ListOptions{}, nodes)
	if err != nil {
		return reconcile.Result{}, err
	}

	for index := 0; index < len(nodes.Items); index++ {
		if val, ok := nodes.Items[index].Labels["kubernetes.io/role"]; ok {
			if val == "node" {
				count++
			}
		}
	}

	reqLogger.Info("Aqua CSP Scanner Scale:", "Kubernetes Nodes Count:", count)

	scanners := result.Count / scannersPerImages
	extraScanners := result.Count % scannersPerImages

	if scanners == 0 {
		scanners = min
	} else {
		if extraScanners > 0 {
			scanners++
		}

		if (max * count) < scanners {
			scanners = (max * count)
		}
	}

	found := &operatorv1alpha1.AquaScanner{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if found != nil && err == nil {
		if found.Spec.ScannerService.Replicas == scanners {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
		}

		if result.Count > 0 {
			found.Spec.ScannerService.Replicas = scanners
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Aqua CSP Scanner Scale: Failed to update Aqua Scanner.", "AquaScanner.Namespace", found.Namespace, "AquaScanner.Name", found.Name)
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(0)}, nil
}
