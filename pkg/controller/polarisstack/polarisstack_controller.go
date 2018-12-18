package polarisstack

import (
	"context"
	goerrors "errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	polarisv1alpha1 "github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"
	"github.com/synthesis-labs/polaris-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_polarisstack")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PolarisStack Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePolarisStack{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("polarisstack-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PolarisStack
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisStack{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolarisStack{}

// ReconcilePolarisStack reconciles a PolarisStack object
type ReconcilePolarisStack struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PolarisStack object and makes changes based on the state read
// and what is in the PolarisStack.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePolarisStack) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PolarisStack")

	// Fetch the PolarisStack instance
	instance := &polarisv1alpha1.PolarisStack{}
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

	// If instructed to delete
	//
	if instance.DeletionTimestamp != nil {
		deleteStack(sess, instance)

		err = r.client.Update(context.TODO(), instance)
		reqLogger.Info("Updated status after deletion", "err", err)

		return reconcile.Result{Requeue: false}, nil
	}

	// Ok we are definitely either creating or updating the stack
	//
	nickname := instance.Spec.Nickname
	templateBody := instance.Spec.Template
	params := []*cloudformation.Parameter{}
	for parameterKey, parameterValue := range instance.Spec.Parameters {
		param := cloudformation.Parameter{
			ParameterKey:   aws.String(parameterKey),
			ParameterValue: aws.String(parameterValue),
		}
		params = append(params, &param)
	}
	finalizers := instance.Spec.Finalizers

	// Have we got a stack name?
	//
	if instance.Status.Name != "" {
		// We already have a stack name, check what cloudformation has...
		//
		stack, err := getStack(sess, instance.Status.Name)
		if err != nil {
			// Cloudformation does not have it so create it (re-using the same name)
			//
			err = createStack(sess, instance.Status.Name, params, templateBody, finalizers)
			if err != nil {
				reqLogger.Error(err, "Unable to create stack")
				return reconcile.Result{}, err
			}

			// Add the finalizers so that we can delete the stack in future
			//
			utils.AddFinalizers(&instance.ObjectMeta, finalizers...)

		} else {
			// Cloudformation has got it - just update it at CF and then pull status back down
			//
			err = updateStack(sess, instance.Status.Name, params, templateBody)
			if err != nil {
				if !(strings.Contains(err.Error(), "ValidationError: No updates are to be performed.") ||
					strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated")) {
					reqLogger.Error(err, "Unable to update stack")
					return reconcile.Result{}, err
				}
			}

			// Update the state etc
			//
			instance.Status.Status = *stack.StackStatus
			instance.Status.Name = *stack.StackName
			instance.Status.ID = *stack.StackId
			instance.Status.Outputs = map[string]string{}
			for _, output := range stack.Outputs {
				instance.Status.Outputs[*output.OutputKey] = *output.OutputValue
			}
			if stack.StackStatusReason != nil {
				instance.Status.StatusReason = *stack.StackStatusReason
			}
		}
	} else {
		// First time we are creating it
		//
		instance.Status.Name = strings.ToLower(fmt.Sprintf("polaris-stack-%s-%s-%s-%s", nickname, request.Namespace, request.Name, utils.GetULID()))
		err = createStack(sess, instance.Status.Name, params, templateBody, finalizers)
		if err != nil {
			reqLogger.Error(err, "Unable to create stack")
			return reconcile.Result{}, err
		}

		// Add the finalizers so that we can delete the stack in future
		//
		utils.AddFinalizers(&instance.ObjectMeta, finalizers...)
	}

	err = r.client.Update(context.TODO(), instance)
	return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
}

func updateStack(
	sess *session.Session,
	stackName string,
	parameters []*cloudformation.Parameter,
	templateBody string) error {
	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	// Maybe we need to do an UpdateStack? Check if the details have changed i guess?
	//
	capabilityIam := "CAPABILITY_IAM"
	_, err := cloudformationsvc.UpdateStack((&cloudformation.UpdateStackInput{}).
		SetStackName(stackName).
		SetTemplateBody(templateBody).
		SetParameters(parameters).
		SetCapabilities([]*string{&capabilityIam}),
	)
	return err
}

func createStack(
	sess *session.Session,
	stackName string,
	parameters []*cloudformation.Parameter,
	templateBody string,
	finalizers []string) error {
	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	capabilityIam := "CAPABILITY_IAM"
	_, err := cloudformationsvc.CreateStack((&cloudformation.CreateStackInput{}).
		SetStackName(stackName).
		SetTemplateBody(templateBody).
		SetParameters(parameters).
		SetCapabilities([]*string{&capabilityIam}),
	)

	return err
}

func getStack(sess *session.Session, stackName string) (*cloudformation.Stack, error) {
	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	resp, err := cloudformationsvc.DescribeStacks((&cloudformation.DescribeStacksInput{}).
		SetStackName(stackName))
	if err != nil {
		return nil, err
	} else if len(resp.Stacks) < 1 {
		return nil, goerrors.New("Error from DescribeStacks -> Unable to find stack")
	}

	// TODO: Detect drift here too
	//

	return resp.Stacks[0], nil
}

func deleteStack(sess *session.Session, instance *polarisv1alpha1.PolarisStack) error {

	reqLogger := log.WithValues()

	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	// Describe the stack to tear down any S3 buckets or whatevs
	//
	resp, err := cloudformationsvc.DescribeStackResources((&cloudformation.DescribeStackResourcesInput{}).
		SetStackName(instance.Status.Name))
	if err != nil {
		return err
	}

	for _, resource := range resp.StackResources {
		reqLogger.Info(" - Checking whether we should delete this resource", "ResourceType", *resource.ResourceType)

		if resource.PhysicalResourceId == nil {
			continue
		}

		if *resource.ResourceType == "AWS::S3::Bucket" &&
			utils.HasFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack.buckets") {

			// First we must delete all the objects
			//
			s3svc := s3.New(sess)

			// Iteratively delete each object in the bucket
			//
			iter := s3manager.NewDeleteListIterator(s3svc, (&s3.ListObjectsInput{}).SetBucket(*resource.PhysicalResourceId))
			err := s3manager.NewBatchDeleteWithClient(s3svc).Delete(aws.BackgroundContext(), iter)
			if err != nil {
				return err
			}

			// Then delete the bucket
			//
			_, err = s3svc.DeleteBucket((&s3.DeleteBucketInput{}).SetBucket(*resource.PhysicalResourceId))
			if err != nil {
				return err
			}

			utils.RemoveFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack.buckets")

		} else if *resource.ResourceType == "AWS::ECR::Repository" &&
			utils.HasFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack.containerregistry") {

			// Force delete the ECR
			//
			ecrsvc := ecr.New(sess)

			_, err := ecrsvc.DeleteRepository((&ecr.DeleteRepositoryInput{}).
				SetRepositoryName(*resource.PhysicalResourceId).
				SetForce(true))
			if err != nil {
				return err
			}

			utils.RemoveFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack.containerregistry")
		}
	}

	// Delete the stack and remove the finalizer
	//
	if utils.HasFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack") {
		reqLogger.Info("Deleting stack!")
		_, err := cloudformationsvc.DeleteStack((&cloudformation.DeleteStackInput{}).
			SetStackName(instance.Status.Name))
		if err != nil {
			return err
		}

		utils.RemoveFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack")
	}

	return nil
}
