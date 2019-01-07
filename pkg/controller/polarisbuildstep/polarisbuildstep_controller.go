package polarisbuildstep

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	polarisv1alpha1 "github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"
	"github.com/synthesis-labs/polaris-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_polarisbuildstep")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PolarisBuildStep Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePolarisBuildStep{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("polarisbuildstep-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PolarisBuildStep
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisBuildStep{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolarisBuildStep{}

// ReconcilePolarisBuildStep reconciles a PolarisBuildStep object
type ReconcilePolarisBuildStep struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PolarisBuildStep object and makes changes based on the state read
// and what is in the PolarisBuildStep.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePolarisBuildStep) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PolarisBuildStep")

	// Fetch the PolarisBuildStep instance
	instance := &polarisv1alpha1.PolarisBuildStep{}
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

	// Try and find the referenced build pipeline (based on pipeline name in the same namespace)
	//
	pipeline := polarisv1alpha1.PolarisBuildPipeline{}

	// Run the query
	//
	r.client.Get(
		context.TODO(),
		client.ObjectKey{Name: instance.Spec.Pipeline, Namespace: instance.Namespace},
		&pipeline)

	if pipeline.Name == "" {
		return reconcile.Result{}, fmt.Errorf("Unable to find pipeline matching Name: %s Namespace: %s", instance.Spec.Pipeline, instance.Namespace)
	}
	reqLogger.Info("Found pipeline", "Pipeline:", pipeline.Name, "Namespace:", pipeline.Namespace)

	// If we are being deleted...
	//
	if instance.DeletionTimestamp != nil {

		// Remove our build steps from the pipeline
		//
		for _, build := range instance.Spec.Builds {
			// Making sure we haven't already been added
			//
			var foundAt = -1
			for pos, existingBuild := range pipeline.Spec.Builds {
				if existingBuild.Name == build.Name {
					foundAt = pos
				}
			}
			if foundAt != -1 {
				reqLogger.Info("Removing existing build step")
				pipeline.Spec.Builds = append(pipeline.Spec.Builds[:foundAt], pipeline.Spec.Builds[foundAt+1:]...)
			}
		}

		// Remove finalizer from ourself
		//
		utils.RemoveFinalizer(&instance.ObjectMeta, "polaris.cleanup.buildstep")

		// Remove finalizer from the buildpipeline
		//
		utils.RemoveFinalizer(&pipeline.ObjectMeta, fmt.Sprintf("polaris.cleanup.buildstep.%s", instance.Name))

		// Update pipeline first (it might be already updated so we may need to retry if we get an error)
		//
		err = r.client.Update(context.TODO(), &pipeline)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Now update the instance and let it be deleted
		//
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: false}, nil
	}

	// Make sure that pipeline is our owner
	//
	if err := controllerutil.SetControllerReference(&pipeline, instance, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Merge our builds into the pipeline
	//
	for _, build := range instance.Spec.Builds {
		// Making sure we haven't already been added
		//
		var foundAt = -1
		for pos, existingBuild := range pipeline.Spec.Builds {
			if existingBuild.Name == build.Name {
				foundAt = pos
			}
		}

		// If already existing, then just copy any changed values
		//
		if foundAt != -1 {
			reqLogger.Info("Removing existing build step")
			pipeline.Spec.Builds = append(pipeline.Spec.Builds[:foundAt], pipeline.Spec.Builds[foundAt+1:]...)
		}

		// Otherwise add it
		//
		reqLogger.Info("Adding new build step")
		pipeline.Spec.Builds = append(pipeline.Spec.Builds, build)
	}

	// Add finalizer to myself
	//
	utils.AddFinalizers(&instance.ObjectMeta, "polaris.cleanup.buildstep")

	// Add finalizer to the buildpipeline
	//
	utils.AddFinalizers(&pipeline.ObjectMeta, fmt.Sprintf("polaris.cleanup.buildstep.%s", instance.Name))

	// Update the pipeline first (it might already be updated and we would need to retry)
	//
	err = r.client.Update(context.TODO(), &pipeline)
	if err != nil {
		return reconcile.Result{}, err
	}

	instance.Status.Status = "TBD - fetch from CodeBuild"

	// Update
	//
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Thats it!

	// Continue to requeue - so we can update the build status from codebuild
	//
	return reconcile.Result{}, nil
}
