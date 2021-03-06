package polarisbuildpipeline

import (
	"context"

	polarisv1alpha1 "github.com/synthesis-labs/polaris-operator/pkg/apis/polaris/v1alpha1"
	"github.com/synthesis-labs/polaris-operator/pkg/utils"
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
{{if .Builds}}
{{range .Builds }}  
  CodeBuildProject{{.Name | CloudFormationName}}:
    Type: AWS::CodeBuild::Project
    Properties:
      Artifacts:
        Type: CODEPIPELINE
      Source:
        Type: CODEPIPELINE
        BuildSpec: |
          version: 0.2
          phases:
            pre_build:
              commands:
                - $(aws ecr get-login --no-include-email)
                - BUILDTAG="$(echo $CODEBUILD_RESOLVED_SOURCE_VERSION | head -c 8)"
                - TAGGED_URI="${REPOSITORY_URI}:${TAG}"
                - BUILD_URI="${REPOSITORY_URI}:${BUILDTAG}"
                - LATEST_URI="${REPOSITORY_URI}:latest"
            build:
              commands:
                - cd "$DOCKERFILE_LOCATION"
                - docker build --tag "$TAGGED_URI" .
                - docker build --tag "$BUILD_URI" .
                - docker build --tag "$LATEST_URI" .
            post_build:
              commands:
                - docker push "$TAGGED_URI"
                - docker push "$BUILD_URI"
                - docker push "$LATEST_URI"
        
      Environment:
        ComputeType: BUILD_GENERAL1_LARGE
        Image: aws/codebuild/docker:17.09.0
        Type: LINUX_CONTAINER
        EnvironmentVariables:
          - Name: AWS_DEFAULT_REGION
            Value: !Ref AWS::Region
          - Name: REPOSITORY_URI
            Value: !Sub ${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/{{.ContainerRegistry}}
          - Name: DOCKERFILE_LOCATION
            Value: {{ .DockerfileLocation }}
          - Name: TAG
            Value: {{ .Tag }}
      Name: !Sub ${PipelineName}-build-{{.Name }}
      ServiceRole: !Ref CodeBuildServiceRole
{{end}}
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
            - Name: fetch-sources
              ActionTypeId:
                Category: Source
                Owner: AWS
                Version: 1
                Provider: CodeCommit
              Configuration:
                RepositoryName: {{.Source.CodeCommitRepo}}
                BranchName: {{.Source.Branch}}
              OutputArtifacts:
              - Name: Sources
              RunOrder: 1
{{if .Builds}}
        - Name: Build
          Actions:
{{range .Builds }}
            - Name: build-{{.Name}}
              ActionTypeId:
                Category: Build
                Owner: AWS
                Version: 1
                Provider: CodeBuild
              Configuration:
                ProjectName: !Ref CodeBuildProject{{.Name | CloudFormationName}}
              InputArtifacts:
                - Name: Sources
              RunOrder: 2
{{end}}
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

	// Also watch for changes to any build steps where we are the owner
	//
	/*
			err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisBuildStep{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &polarisv1alpha1.PolarisBuildPipeline{},
			})
			if err != nil {
				return err
		  }
	*/

	// And also watch for changes to any stacks where we are the owner
	//
	err = c.Watch(&source.Kind{Type: &polarisv1alpha1.PolarisStack{}}, &handler.EnqueueRequestForOwner{
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

	// Create the definition of the stack for this instance
	//
	stack, err := getStackForInstance(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

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

	// Stack already exists - do an update
	//
	reqLogger.Info("Skip reconcile: Stack already exists", "Stack.Namespace", found.Namespace, "Stack.Name", found.Name)

	// Merge the spec from the new one
	//
	stack.Spec.DeepCopyInto(&found.Spec)
	err = r.client.Update(context.TODO(), found)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func getStackForInstance(instance *polarisv1alpha1.PolarisBuildPipeline) (*polarisv1alpha1.PolarisStack, error) {
	templateBody, err := utils.RenderCloudformationTemplate(formationTemplate, instance.Spec)
	if err != nil {
		return nil, err
	}

	labels := map[string]string{
		"app":               instance.Name,
		"polaris-project":   instance.Labels["polaris-project"],
		"polaris-component": instance.Labels["polaris-component"],
	}
	return &polarisv1alpha1.PolarisStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-ci-stack",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: polarisv1alpha1.PolarisStackSpec{
			Nickname: "ci",
			Finalizers: []string{
				"polaris.cleanup.aws.stack.buckets",
				"polaris.cleanup.aws.stack",
			},
			Parameters: map[string]string{
				"PipelineName": instance.Namespace + "-" + instance.Name,
			},
			Template: templateBody,
		},
	}, nil
}
