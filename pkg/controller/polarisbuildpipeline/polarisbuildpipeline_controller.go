package polarisbuildpipeline

import (
	"context"
	"fmt"
	"strings"

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
Description: 'Polaris operator - PolarisBuildPipeline'
Parameters:
  PipelineName:
    Type: String
    Description: >-
      Unique name of the pipeline
Resources:
  CodeBuildServiceRole:
    Type: AWS::IAM::Role
    Properties:
      Path: /
      AssumeRolePolicyDocument:
        Version: 2012-10-17
        Statement:
          - Effect: Allow
            Principal:
              Service: codebuild.amazonaws.com
            Action: sts:AssumeRole
      Policies:
        - PolicyName: root
          PolicyDocument:
            Version: 2012-10-17
            Statement:
              - Resource: "*"
                Effect: Allow
                Action:
                  - logs:CreateLogGroup
                  - logs:CreateLogStream
                  - logs:PutLogEvents
                  - ecr:GetAuthorizationToken
                  - s3:*
              - Resource: "arn:aws:ecr:*"
                Effect: Allow
                Action:
                  - ecr:GetDownloadUrlForLayer
                  - ecr:BatchGetImage
                  - ecr:BatchCheckLayerAvailability
                  - ecr:PutImage
                  - ecr:InitiateLayerUpload
                  - ecr:UploadLayerPart
                  - ecr:CompleteLayerUpload
  CodePipelineServiceRole:
    Type: AWS::IAM::Role
    Properties:
      Path: /
      AssumeRolePolicyDocument:
        Version: 2012-10-17
        Statement:
          - Effect: Allow
            Principal:
              Service: codepipeline.amazonaws.com
            Action: sts:AssumeRole
      Policies:
        - PolicyName: root
          PolicyDocument:
            Version: 2012-10-17
            Statement:
              - Resource: "*"
                Effect: Allow
                Action:
                  - ecs:DescribeServices
                  - ecs:DescribeTaskDefinition
                  - ecs:DescribeTasks
                  - ecs:ListTasks
                  - ecs:RegisterTaskDefinition
                  - ecs:UpdateService
                  - codebuild:StartBuild
                  - codebuild:BatchGetBuilds
                  - iam:PassRole
                  - codecommit:*
                  - s3:*
  ArtifactBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub ${PipelineName}-artifacts
      AccessControl: Private
{{range .Builds }}  
  CodeBuildProject{{.Name | CloudFormationName}}:
    Type: AWS::CodeBuild::Project
    Properties:
      Artifacts:
        Type: CODEPIPELINE
      Source:
        Type: CODEPIPELINE
        BuildSpec: {{.Buildspec}}
      Environment:
        ComputeType: BUILD_GENERAL1_LARGE
        Image: aws/codebuild/docker:17.09.0
        Type: LINUX_CONTAINER
        EnvironmentVariables:
          - Name: AWS_DEFAULT_REGION
            Value: !Ref AWS::Region
          - Name: REPOSITORY_URI
            Value: !Sub ${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/{{.ContainerRepository}}
      Name: !Sub ${PipelineName}-build-{{.Name }}
      ServiceRole: !Ref CodeBuildServiceRole
{{end}}
  Pipeline:
    Type: AWS::CodePipeline::Pipeline
    Properties:
      Name: !Sub ${PipelineName}-pipeline
      RoleArn: !GetAtt CodePipelineServiceRole.Arn
      ArtifactStore:
        Type: S3
        Location: !Ref ArtifactBucket
      Stages:
        - Name: Source
          Actions:
            - Name: App
              ActionTypeId:
                Category: Source
                Owner: AWS
                Version: 1
                Provider: CodeCommit
              Configuration:
                RepositoryName: {{.Source.CodeCommitRepo}}
                BranchName: {{.Source.Branch}}
              OutputArtifacts:
              - Name: App                
              RunOrder: 1
        - Name: Build
          Actions:
{{range .Builds }}
            - Name: Build-{{.Name}}
              ActionTypeId:
                Category: Build
                Owner: AWS
                Version: 1
                Provider: CodeBuild
              Configuration:
                ProjectName: !Ref CodeBuildProject{{.Name | CloudFormationName}}
              InputArtifacts:
                - Name: App
              RunOrder: 2
{{end}}
`

var log = logf.Log.WithName("controller_polarisbuildpipeline")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PolarisBuildPipeline Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePolarisBuildPipeline{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("polarisbuildpipeline-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PolarisBuildPipeline
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisBuildPipeline{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PolarisBuildPipeline
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &polarisv1alpha1.PolarisBuildPipeline{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolarisBuildPipeline{}

// ReconcilePolarisBuildPipeline reconciles a PolarisBuildPipeline object
type ReconcilePolarisBuildPipeline struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PolarisBuildPipeline object and makes changes based on the state read
// and what is in the PolarisBuildPipeline.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePolarisBuildPipeline) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PolarisBuildPipeline")

	// Fetch the PolarisBuildPipeline instance
	instance := &polarisv1alpha1.PolarisBuildPipeline{}
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

	templateBody, err := utils.RenderCloudformationTemplate(formationTemplate, instance.Spec)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)

	if instance.DeletionTimestamp != nil {
		utils.ProcessStackDeletion(sess, &instance.ObjectMeta, &instance.Stack)
	} else {

		// Deal with local parameters in the template
		//
		if instance.Status.PipelineName == "" {
			// Generate a new pipeline name
			//
			instance.Status.PipelineName = strings.ToLower(fmt.Sprintf("pbp-%s-%s-%s", request.Namespace, request.Name, utils.GetULID()))
		}

		// Parameters for the template
		//
		param := cloudformation.Parameter{
			ParameterKey:   aws.String("PipelineName"),
			ParameterValue: aws.String(instance.Status.PipelineName),
		}
		params := []*cloudformation.Parameter{&param}

		finalizers := []string{
			"polaris.cleanup.aws.stack",
			"polaris.cleanup.aws.stack.buckets",
		}

		utils.ProcessStackCreation(request, sess, "ci", &instance.ObjectMeta, &instance.Stack, params, templateBody, finalizers)
	}

	err = r.client.Update(context.TODO(), instance)
	reqLogger.Info("Updated status", "err", err)

	return reconcile.Result{}, nil
}
