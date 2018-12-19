package polarissourcerepository

import (
	"context"

	polarisv1alpha1 "github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var formationTemplate = `AWSTemplateFormatVersion: 2010-09-09
Description: 'Polaris operator - PolarisSourceRepository'
Parameters:
    RepositoryName:
        Type: String
        Description: >-
            What should the ECR Repository be named
Resources:
    SourceRepository:
        Type: AWS::CodeCommit::Repository
        Properties: 
            RepositoryDescription: !Ref RepositoryName
            RepositoryName: !Ref RepositoryName
Outputs:
    RepositoryName:
        Value: !Ref SourceRepository
        Description: Name of the repository
    RepositoryARN:
        Value: !GetAtt SourceRepository.Arn
        Description: ARN of the Repository
`

var log = logf.Log.WithName("controller_polarissourcerepository")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PolarisSourceRepository Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePolarisSourceRepository{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("polarissourcerepository-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PolarisSourceRepository
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisSourceRepository{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PolarisSourceRepository
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisStack{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &polarisv1alpha1.PolarisSourceRepository{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolarisSourceRepository{}

// ReconcilePolarisSourceRepository reconciles a PolarisSourceRepository object
type ReconcilePolarisSourceRepository struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PolarisSourceRepository object and makes changes based on the state read
// and what is in the PolarisSourceRepository.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePolarisSourceRepository) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PolarisSourceRepository")

	// Fetch the PolarisSourceRepository instance
	instance := &polarisv1alpha1.PolarisSourceRepository{}
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

	// Create the definition of the stack for this instance
	//
	stack := getStackForInstance(instance)

	// Set instance as the owner and controller
	//
	if err := controllerutil.SetControllerReference(instance, stack, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the stack already exists
	//
	found := &polarisv1alpha1.PolarisStack{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: stack.Name, Namespace: stack.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new stack", "Stack.Namespace", stack.Namespace, "Stack.Name", stack.Name)
		err = r.client.Create(context.TODO(), stack)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Stack created successfully - don't requeue
		//
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Stack already exists - don't requeue
	reqLogger.Info("Skip reconcile: Stack already exists", "Stack.Namespace", found.Namespace, "Stack.Name", found.Name)
	return reconcile.Result{}, nil
}

func getStackForInstance(instance *polarisv1alpha1.PolarisSourceRepository) *polarisv1alpha1.PolarisStack {
	labels := map[string]string{
		"app": instance.Name,
	}
	return &polarisv1alpha1.PolarisStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-sourcerepository-stack",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: polarisv1alpha1.PolarisStackSpec{
			Nickname: "sourcerepository",
			Finalizers: []string{
				"polaris.cleanup.aws.stack",
			},
			Parameters: map[string]string{
				"RepositoryName": instance.Spec.Name,
			},
			Template: formationTemplate,
		},
	}
}
