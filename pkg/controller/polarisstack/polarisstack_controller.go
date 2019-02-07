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
		Region: aws.String(utils.Region())},
	)

	// If instructed to delete
	//
	if instance.DeletionTimestamp != nil {
		err = deleteStack(sess, instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}

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

	stackName := strings.ToLower(fmt.Sprintf("polaris-stack-%s-%s-%s", nickname, request.Namespace, request.Name))

	// We already have a stack name, check what cloudformation has...
	//
	stack, err := getStack(sess, stackName)
	if err != nil {
		// Cloudformation does not have it so create it (re-using the same name)
		//
		instance.Status.Name = stackName
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
		err = updateStack(sess, stackName, params, templateBody)
		if err != nil {
			// ROLLBACK_COMPLETE is a different case
			//
			if strings.Contains(err.Error(), "ROLLBACK_COMPLETE state and can not be updated") {
				reqLogger.Info("Deleteing rolled back stack!")
				err = deleteStack(sess, instance)
				if err != nil {
					reqLogger.Error(err, "Unable to delete stack after it was rolled back")
					return reconcile.Result{}, err
				}
			} else if !(strings.Contains(err.Error(), "ValidationError: No updates are to be performed") ||
				strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated")) { // Deals with CREATE_ and UPDATE_ cases
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

	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

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

	if err != nil {
		return err
	}

	log.Info("Created stack", "StackName", stackName)
	return nil
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

	// TODO: Detect drift here too maybe?
	//

	return resp.Stacks[0], nil
}

// Higher order function which encapsulates the logic of:
// if correctResourceType() and hasFinalizer() then
//    deleteResource()
//    removeFinalizer()
//
func deleteResourceAndRemoveFinalizer(
	instance *polarisv1alpha1.PolarisStack,
	resource *cloudformation.StackResource,
	finalizer string,
	resourceType string,
	deleteFunc func() error) error {
	if *resource.ResourceType == resourceType && utils.HasFinalizer(&instance.ObjectMeta, finalizer) {
		log.Info("Deleting specific resource", "ResourceType", resourceType, "Finalizer", finalizer)
		err := deleteFunc()
		if err != nil {
			return err
		}
		utils.RemoveFinalizer(&instance.ObjectMeta, finalizer)
	}
	return nil
}

func deleteStack(sess *session.Session, instance *polarisv1alpha1.PolarisStack) error {

	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	// Describe the stack to tear down any S3 buckets or whatevs
	//
	resp, err := cloudformationsvc.DescribeStackResources((&cloudformation.DescribeStackResourcesInput{}).
		SetStackName(instance.Status.Name))

	// Ignore the fact that the stack might already be deleted
	//
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		// Do nothing - stack already deleted, but need to carry on to remove all the finalizers
		//
	} else if err != nil {
		log.Info("Error while trying to delete stack: ", "error", err)
		return err
	}

	// Very specific delete apis for each naughty resource that cloudformation refuses to delete
	//
	for _, resource := range resp.StackResources {
		if resource.PhysicalResourceId == nil {
			continue
		}

		// Handle S3 Buckets (delete all objects, then the bucket itself)
		//
		err = deleteResourceAndRemoveFinalizer(instance, resource, "polaris.cleanup.aws.stack.buckets", "AWS::S3::Bucket", func() error {
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
			return nil
		})

		// Handle ecrs - they may contain images so cloudformation doesn't remove them
		//
		err = deleteResourceAndRemoveFinalizer(instance, resource, "polaris.cleanup.aws.stack.containerregistry", "AWS::ECR::Repository", func() error {
			// Force delete the ECR
			//
			ecrsvc := ecr.New(sess)

			_, err := ecrsvc.DeleteRepository((&ecr.DeleteRepositoryInput{}).
				SetRepositoryName(*resource.PhysicalResourceId).
				SetForce(true))
			if err != nil {
				return err
			}

			return nil
		})
	}

	// Delete the stack and remove the finalizer
	//
	if utils.HasFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack") {
		utils.RemoveFinalizer(&instance.ObjectMeta, "polaris.cleanup.aws.stack")
	}

	log.Info("Deleting stack", "StackName", instance.Status.Name)
	_, err = cloudformationsvc.DeleteStack((&cloudformation.DeleteStackInput{}).
		SetStackName(instance.Status.Name))

	if err != nil && strings.Contains(err.Error(), "does not exist") {
		log.Info("Stack does not exist")
	} else if err != nil {
		return err
	}

	return nil
}
