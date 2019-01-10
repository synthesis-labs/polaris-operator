package utils

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("utils")

// HasFinalizer checks whether a particular finalizer is set
//
func HasFinalizer(instance *metav1.ObjectMeta, finalizer string) bool {
	for _, containedFinalizer := range instance.Finalizers {
		if containedFinalizer == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizers adds a particular finalizer to an instance
//
func AddFinalizers(instance *metav1.ObjectMeta, finalizers ...string) {
	for _, finalizer := range finalizers {
		if !HasFinalizer(instance, finalizer) {
			instance.SetFinalizers(append(instance.Finalizers, finalizer))
		}
	}
}

// RemoveFinalizer removes a particular finalizer from an instance
//
func RemoveFinalizer(instance *metav1.ObjectMeta, finalizer string) {
	newFinalizers := []string{}
	for _, containedFinalizer := range instance.Finalizers {
		if containedFinalizer != finalizer {
			newFinalizers = append(newFinalizers, containedFinalizer)
		}
	}
	instance.SetFinalizers(newFinalizers)
}

// ProcessStackDeletion processes all the finalizers and deletes the
//
func ProcessStackDeletion(sess *session.Session, instance *metav1.ObjectMeta, stack *v1alpha1.PolarisCloudformationStatus) {

	reqLogger := log.WithValues()

	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	// Describe the stack to tear down any S3 buckets or whatevers
	//
	resp, err := cloudformationsvc.DescribeStackResources((&cloudformation.DescribeStackResourcesInput{}).
		SetStackName(stack.StackName))
	if err != nil {
		if stack.StackError != err.Error() {
			stack.StackError = err.Error()
		}
		fmt.Printf("Error from DescribeStackResources -> %s", err.Error())

	} else {
		stack.StackResponse = resp.String()

		for _, resource := range resp.StackResources {
			reqLogger.Info(" - Checking whether we should delete this resource", "ResourceType", *resource.ResourceType)

			if resource.PhysicalResourceId == nil {
				continue
			}

			if *resource.ResourceType == "AWS::S3::Bucket" &&
				HasFinalizer(instance, "polaris.cleanup.aws.stack.buckets") {

				// Delete that bucket!
				//
				reqLogger.Info("Deleting Bucket", "PhysicalId", *resource.PhysicalResourceId)

				// First we must delete all the objects
				//
				s3svc := s3.New(sess)

				// Iteratively delete each object in the bucket
				//
				iter := s3manager.NewDeleteListIterator(s3svc, (&s3.ListObjectsInput{}).SetBucket(*resource.PhysicalResourceId))
				err := s3manager.NewBatchDeleteWithClient(s3svc).Delete(aws.BackgroundContext(), iter)

				// Then delete the bucket
				//
				resp, err := s3svc.DeleteBucket((&s3.DeleteBucketInput{}).SetBucket(*resource.PhysicalResourceId))
				reqLogger.Info("DeleteBucket returned", "Resp", resp, "Err", err)

				RemoveFinalizer(instance, "polaris.cleanup.aws.stack.buckets")

			} else if *resource.ResourceType == "AWS::ECR::Repository" &&
				HasFinalizer(instance, "polaris.cleanup.aws.stack.containerregistry") {

				// Force delete the ECR
				//
				ecrsvc := ecr.New(sess)

				err, resp := ecrsvc.DeleteRepository((&ecr.DeleteRepositoryInput{}).
					SetRepositoryName(*resource.PhysicalResourceId).
					SetForce(true))

				reqLogger.Info("DeleteRepository returned", "Resp", resp, "Err", err)

				RemoveFinalizer(instance, "polaris.cleanup.aws.stack.containerregistry")
			}
		}

		// Delete the stack and remove the finalizer
		//
		if HasFinalizer(instance, "polaris.cleanup.aws.stack") {
			reqLogger.Info("Deleting stack!")
			resp, err := cloudformationsvc.DeleteStack((&cloudformation.DeleteStackInput{}).
				SetStackName(stack.StackName))

			reqLogger.Info("DeleteStack returned", "Resp", resp, "Err", err)

			if err != nil {
				stack.StackError = err.Error()
			}
			if resp != nil {
				stack.StackResponse = resp.String()
			}

			RemoveFinalizer(instance, "polaris.cleanup.aws.stack")
		}
	}
}

// ProcessStackCreation creates and sets up the stack
//
func ProcessStackCreation(
	request reconcile.Request,
	sess *session.Session,
	friendlyType string,
	instance *metav1.ObjectMeta,
	stack *v1alpha1.PolarisCloudformationStatus,
	parameters []*cloudformation.Parameter,
	templateBody string,
	finalizers []string) {
	reqLogger := log.WithValues()

	// Create cloudformation service client
	//
	cloudformationsvc := cloudformation.New(sess)

	if stack.StackCreationAttempted {
		reqLogger.Info("Stack creation already attempted")

		// Maybe we need to do an UpdateStack? Check if the details have changed i guess?
		//
		capabilityIam := "CAPABILITY_IAM"
		resp, err := cloudformationsvc.UpdateStack((&cloudformation.UpdateStackInput{}).
			SetStackName(stack.StackName).
			SetTemplateBody(templateBody).
			SetParameters(parameters).
			SetCapabilities([]*string{&capabilityIam}),
		)
		reqLogger.Info("UpdateStack returned", "Resp", resp, "Err", err)
		if err != nil {

			if strings.Contains(err.Error(), "ValidationError: No updates are to be performed.") {
				// Ignore the error if it already exists (thats totally ok)
				//
				stack.StackError = ""
			} else {
				stack.StackError = err.Error()
			}
		}
		if resp != nil {
			stack.StackResponse = resp.String()
		}
	} else {

		// Create a new stack
		//
		reqLogger.Info("Creating stack!")

		// Stack name example:    "polaris-buildpipeline-default-example-01CYBVGAEQSBVPDWAMDSSPJQA4"
		// Pipeline name example: "pbp-default-example-01CYBVGAEQSBVPDWAMDSSPJQA4"

		stackName := strings.ToLower(fmt.Sprintf("polaris-stack-%s-%s-%s-%s", friendlyType, request.Namespace, request.Name, GetULID()))

		// Remember the stackname
		//
		stack.StackName = stackName
		stack.StackCreationAttempted = true

		capabilityIam := "CAPABILITY_IAM"
		resp, err := cloudformationsvc.CreateStack((&cloudformation.CreateStackInput{}).
			SetStackName(stackName).
			SetTemplateBody(templateBody).
			SetParameters(parameters).
			SetCapabilities([]*string{&capabilityIam}),
		)

		reqLogger.Info("CreateStack returned", "Resp", resp, "Err", err)
		if err != nil {
			stack.StackError = err.Error()
		}
		if resp != nil {
			stack.StackResponse = resp.String()
		}

		// Add the finalizers so that we can delete the stack in future
		//
		AddFinalizers(instance, finalizers...)
	}
}
