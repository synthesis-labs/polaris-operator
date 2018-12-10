package polarissourcerepository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	polarisv1alpha1 "github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"
	"github.com/synthesis-labs/polaris-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
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
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)

	// Create cloudformation service client
	svc := cloudformation.New(sess)

	if instance.DeletionTimestamp != nil {
		// Delete the stack and remove the finalizer
		//
		reqLogger.Info("Deleting stack!")
		resp, err := svc.DeleteStack((&cloudformation.DeleteStackInput{}).
			SetStackName(instance.Status.StackName))
		reqLogger.Info("DeleteStack returned", "Resp", resp, "Err", err)

		if err != nil {
			instance.Status.StackError = err.Error()
		}
		if resp != nil {
			instance.Status.StackResponse = resp.String()
		}

		instance.SetFinalizers([]string{})
	} else {

		if instance.Status.StackCreationAttempted {
			reqLogger.Info("Stack creation already attempted")
			return reconcile.Result{}, nil
		}

		reqLogger.Info("Creating stack!")

		stackName := fmt.Sprintf("polaris-sourcerepository-%s-%s-%s", request.Namespace, request.Name, utils.GetULID())

		// Remember the stackname
		//
		instance.Status.StackName = stackName
		instance.Status.StackCreationAttempted = true

		param := cloudformation.Parameter{
			ParameterKey:   aws.String("RepositoryName"),
			ParameterValue: aws.String(instance.Spec.Name),
		}
		params := []*cloudformation.Parameter{&param}
		resp, err := svc.CreateStack((&cloudformation.CreateStackInput{}).
			SetStackName(stackName).
			SetTemplateBody(formationTemplate).
			SetParameters(params),
		)
		reqLogger.Info("CreateStack returned", "Resp", resp, "Err", err)
		if err != nil {
			instance.Status.StackError = err.Error()
		}
		if resp != nil {
			instance.Status.StackResponse = resp.String()
		}

		// Add a finalizer so that we can delete the stack in future
		//
		instance.SetFinalizers([]string{"polarisbuildoperator.must.delete.aws.cloudformation"})
	}

	err = r.client.Update(context.TODO(), instance)
	reqLogger.Info("Updated status", "err", err)

	return reconcile.Result{}, nil
}
