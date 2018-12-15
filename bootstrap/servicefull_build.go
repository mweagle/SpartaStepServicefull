// +build !lambdabinary

package bootstrap

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	iamBuilder "github.com/mweagle/Sparta/aws/iam/builder"
	spartaStep "github.com/mweagle/Sparta/aws/step"
	spartaDocker "github.com/mweagle/Sparta/docker"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// contextKeyImageTag is the docker image tag to use for the task provider
	contextKeyImageTag string = "imageTag"
	// contextKeyImageURL is the docker URL in the context to use for the fargate task
	contextKeyImageURL string = "imageURL"
)

// Utility function to create logical
// CloudFormation Resource names that are stable
// across builds.
func resourceName(baseName string) string {
	return sparta.CloudFormationResourceName(baseName, baseName)
}

type stackResourceNames struct {
	StepFunction              string
	SNSTopic                  string
	ECSCluster                string
	ECSRunTaskRole            string
	ECSTaskDefinition         string
	ECSTaskDefinitionLogGroup string
	ECSTaskDefinitionRole     string
	VPC                       string
	InternetGateway           string
	AttachGateway             string
	RouteViaIgw               string
	PublicRouteViaIgw         string
	ECSSecurityGroup          string
	PublicSubnetAzs           []string
}

func newStackResourceNames() *stackResourceNames {
	return &stackResourceNames{
		StepFunction:              resourceName("ServicefulStepFunction"),
		SNSTopic:                  resourceName("SNSTopic"),
		ECSCluster:                resourceName("ECSCluster"),
		ECSRunTaskRole:            resourceName("ECSRunTaskSyncExecutionRole"),
		ECSTaskDefinition:         resourceName("ECSTaskDefinition"),
		ECSTaskDefinitionLogGroup: resourceName("ECSTaskDefinitionLogGroup"),
		ECSTaskDefinitionRole:     resourceName("ECSTaskDefinitionRole"),
		VPC:                       resourceName("VPC"),
		InternetGateway:           resourceName("InternetGateway"),
		AttachGateway:             resourceName("AttachGateway"),
		RouteViaIgw:               resourceName("RouteViaIgw"),
		PublicRouteViaIgw:         resourceName("PublicRouteViaIgw"),
		ECSSecurityGroup:          resourceName("ECSSecurityGroup"),
		PublicSubnetAzs: []string{
			resourceName("PubSubnetAz1"),
			resourceName("PubSubnetAz2")},
	}
}

func ecrImageBuilderDecorator(ecrRepositoryName string) sparta.ServiceDecoratorHookHandler {
	decorator := func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {
		dockerServiceName := strings.ToLower(serviceName)
		dockerTags := make(map[string]string, 0)

		// We use a nonce build id since normally this wouldn't cause a
		// new image URL because the build ID is a stable git SHA. Using a stable
		// ID causes CloudFormation to understandably reject the update request,
		// so add a bit of sugar.
		nonceBuildID := fmt.Sprintf("%s.%d", buildID, time.Now().Unix())
		dockerTags[dockerServiceName] = nonceBuildID
		buildTag := fmt.Sprintf("%s:%s", dockerServiceName, nonceBuildID)
		context[contextKeyImageTag] = buildTag

		// Always build the image
		buildErr := spartaDocker.BuildDockerImage(serviceName,
			"",
			dockerTags,
			logger)
		if buildErr != nil {
			return buildErr
		}
		var ecrURL string
		if !noop {
			// Push the image to ECR & store the URL s.t. we can properly annotate
			// the CloudFormation template
			ecrURLPush, pushImageErr := spartaDocker.PushDockerImageToECR(buildTag,
				ecrRepositoryName,
				awsSession,
				logger)
			if nil != pushImageErr {
				return pushImageErr
			}
			ecrURL = ecrURLPush
			logger.WithFields(logrus.Fields{
				"ECRUrl": ecrURL,
			}).Info("Docker image pushed")
		} else {
			ecrURL = fmt.Sprintf("https://123412341234.dkr.ecr.aws-region.amazonaws.com/%s", buildTag)
			logger.WithFields(logrus.Fields{
				"ECRUrl": ecrURL,
			}).Info("Using Docker mock ECR URL due to -n/--noop argument")
		}
		// Save the URL
		context[contextKeyImageURL] = ecrURL
		return nil
	}
	return sparta.ServiceDecoratorHookFunc(decorator)

}

// fargateClusterDecorator returns a ServiceDecoratorHookHandler that
// that provisions an ECS cluster that can run the Fargate task
func fargateClusterDecorator(resourceNames *stackResourceNames) sparta.ServiceDecoratorHookHandler {
	decorator := func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {

		// Let's make them all...
		ecsRunTaskRole := &gocf.IAMRole{}
		ecsRunTaskRole.AssumeRolePolicyDocument = map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []spartaIAM.PolicyStatement{
				iamBuilder.Allow("sts:AssumeRole").
					ForPrincipals("states.amazonaws.com").
					ToPolicyStatement(),
			},
		}
		ecsRunTaskRole.Path = gocf.String("/")
		ecsRunTaskRole.Policies = &gocf.IAMRolePolicyList{
			gocf.IAMRolePolicy{
				PolicyName: gocf.String("FargateTaskNotificationAccessPolicy"),
				PolicyDocument: map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []spartaIAM.PolicyStatement{
						iamBuilder.Allow("sns:Publish").
							ForResource().
							Ref(resourceNames.SNSTopic).
							ToPolicyStatement(),
						iamBuilder.Allow("ecs:RunTask").
							ForResource().
							Ref(resourceNames.ECSTaskDefinition).
							ToPolicyStatement(),
						iamBuilder.Allow("iam:PassRole",
							"ecs:StopTask",
							"ecs:DescribeTasks").
							ForResource().
							Literal("*").
							ToPolicyStatement(),
						iamBuilder.Allow("events:PutTargets",
							"events:PutRule",
							"events:DescribeRule").
							ForResource().
							Literal("arn:").
							Partition().
							Literal(":events:").
							Region(":").
							AccountID().
							Literal(":rule/StepFunctionsGetEventsForECSTaskRule").
							ToPolicyStatement(),
						// Ref: https://docs.aws.amazon.com/AmazonECR/latest/userguide/ECR_on_ECS.html
					},
				},
			},
		}
		template.AddResource(resourceNames.ECSRunTaskRole, ecsRunTaskRole)
		// SNS Topic
		template.AddResource(resourceNames.SNSTopic, &gocf.SNSTopic{})
		// ECS Cluster
		template.AddResource(resourceNames.ECSCluster, &gocf.ECSCluster{})
		// ECS TaskDefinition
		logger.WithField("context", fmt.Sprintf("%#v", context)).Debug("ECS TASK CONTEXT")

		// Get the imageURL from the context. This is the ECR URL to which we
		// previously pushed the locally built image in the ecrImageBuilderDecorator
		// step
		imageURL, _ := context[contextKeyImageURL].(string)
		if imageURL == "" {
			return errors.Errorf("Failed to get image URL from context with key %s", contextKeyImageURL)
		}
		// We need an IAM role to pull images from ECR...
		ecsTaskDefRole := &gocf.IAMRole{}
		ecsTaskDefRole.AssumeRolePolicyDocument = map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []spartaIAM.PolicyStatement{
				iamBuilder.Allow("sts:AssumeRole").
					ForPrincipals("ecs-tasks.amazonaws.com").
					ToPolicyStatement(),
			},
		}
		ecsTaskDefRole.Policies = &gocf.IAMRolePolicyList{
			gocf.IAMRolePolicy{
				PolicyName: gocf.String("FargateTaskNotificationAccessPolicy"),
				PolicyDocument: map[string]interface{}{
					"Version": "2012-10-17",
					// Ref: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html
					"Statement": []spartaIAM.PolicyStatement{
						iamBuilder.Allow("ecr:GetAuthorizationToken",
							"ecr:BatchCheckLayerAvailability",
							"ecr:GetDownloadUrlForLayer",
							"ecr:BatchGetImage",
							"logs:CreateLogStream",
							"logs:CreateLogGroup",
							"logs:PutLogEvents").
							ForResource().
							Literal("*").
							ToPolicyStatement(),
					},
				},
			},
		}
		template.AddResource(resourceNames.ECSTaskDefinitionRole, ecsTaskDefRole)

		// Create the ECS task definition
		ecsTaskDefinition := &gocf.ECSTaskDefinition{
			ExecutionRoleArn:        gocf.GetAtt(resourceNames.ECSTaskDefinitionRole, "Arn"),
			RequiresCompatibilities: gocf.StringList(gocf.String("FARGATE")),
			CPU:                     gocf.String("256"),
			Memory:                  gocf.String("512"),
			NetworkMode:             gocf.String("awsvpc"),
			ContainerDefinitions: &gocf.ECSTaskDefinitionContainerDefinitionList{
				gocf.ECSTaskDefinitionContainerDefinition{
					Image:     gocf.String(imageURL),
					Name:      gocf.String("sparta-servicefull"),
					Essential: gocf.Bool(true),
					LogConfiguration: &gocf.ECSTaskDefinitionLogConfiguration{
						LogDriver: gocf.String("awslogs"),
						// Options Ref: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html
						Options: map[string]interface{}{
							"awslogs-region": gocf.Ref("AWS::Region"),
							"awslogs-group": strings.Join([]string{"",
								sparta.ProperName,
								serviceName}, "/"),
							"awslogs-stream-prefix": serviceName,
							"awslogs-create-group":  "true",
						},
					},
				},
			},
		}
		template.AddResource(resourceNames.ECSTaskDefinition, ecsTaskDefinition)

		// VPC...
		vpc := &gocf.EC2VPC{
			CidrBlock:          gocf.String("10.0.0.0/16"),
			EnableDNSSupport:   gocf.Bool(true),
			EnableDNSHostnames: gocf.Bool(true),
		}
		template.AddResource(resourceNames.VPC, vpc)

		// Subnets
		for i := 0; i != len(resourceNames.PublicSubnetAzs); i++ {
			subnet := &gocf.EC2Subnet{
				VPCID:            gocf.Ref(resourceNames.VPC).String(),
				CidrBlock:        gocf.String(fmt.Sprintf("10.0.%d.0/24", i)),
				AvailabilityZone: gocf.Select("0", gocf.GetAZs(gocf.String(""))),
				Tags: &gocf.TagList{
					gocf.Tag{
						Key: gocf.String("Name"),
						Value: gocf.Join("/",
							gocf.Ref(resourceNames.ECSCluster),
							gocf.String("Public")),
					},
				},
			}
			template.AddResource(resourceNames.PublicSubnetAzs[i], subnet)

			// Wire them up...
			subnetRouteTableAssociation := &gocf.EC2SubnetRouteTableAssociation{
				SubnetID:     gocf.Ref(resourceNames.PublicSubnetAzs[i]).String(),
				RouteTableID: gocf.Ref(resourceNames.RouteViaIgw).String(),
			}
			subnetRouteTableAssociationName := resourceName(fmt.Sprintf("PubSubnet%dRouteTableAssociation", i+1))
			template.AddResource(subnetRouteTableAssociationName, subnetRouteTableAssociation)
		}
		// InternetGateway
		template.AddResource(resourceNames.InternetGateway, &gocf.EC2InternetGateway{})

		// AttachGateway
		template.AddResource(resourceNames.AttachGateway, &gocf.EC2VPCGatewayAttachment{
			VPCID:             gocf.Ref(resourceNames.VPC).String(),
			InternetGatewayID: gocf.Ref(resourceNames.InternetGateway).String(),
		})
		// RouteViaIgw
		template.AddResource(resourceNames.RouteViaIgw, &gocf.EC2RouteTable{
			VPCID: gocf.Ref(resourceNames.VPC).String(),
		})

		// PublicRouteViaIgw
		routeResource := template.AddResource(resourceNames.PublicRouteViaIgw, &gocf.EC2Route{
			RouteTableID:         gocf.Ref(resourceNames.RouteViaIgw).String(),
			DestinationCidrBlock: gocf.String("0.0.0.0/0"),
			GatewayID:            gocf.Ref(resourceNames.InternetGateway).String(),
		})
		routeResource.DependsOn = []string{resourceNames.AttachGateway}
		// Security Group
		template.AddResource(resourceNames.ECSSecurityGroup, &gocf.EC2SecurityGroup{
			GroupDescription: gocf.String("ECS Allowed Ports"),
			VPCID:            gocf.Ref(resourceNames.VPC).String(),
		})
		return nil
	}
	return sparta.ServiceDecoratorHookFunc(decorator)
}

// Run the bootstrap
func Run(logger *logrus.Logger) (*sparta.WorkflowHooks, error) {
	resourceNames := newStackResourceNames()

	// Make the states
	fargateParams := spartaStep.FargateTaskParameters{
		LaunchType:     "FARGATE",
		Cluster:        gocf.Ref(resourceNames.ECSCluster).String(),
		TaskDefinition: gocf.Ref(resourceNames.ECSTaskDefinition).String(),
		NetworkConfiguration: &spartaStep.FargateNetworkConfiguration{
			AWSVPCConfiguration: &gocf.ECSServiceAwsVPCConfiguration{
				Subnets: gocf.StringList(
					gocf.Ref(resourceNames.PublicSubnetAzs[0]).String(),
					gocf.Ref(resourceNames.PublicSubnetAzs[1]).String(),
				),
				AssignPublicIP: gocf.String("ENABLED"),
			},
		},
	}
	fargateState := spartaStep.NewFargateTaskState("Run Fargate Task", fargateParams)

	snsSuccessParams := spartaStep.SNSTaskParameters{
		Message:  "AWS Fargate Task started by Step Functions succeeded",
		TopicArn: gocf.Ref(resourceNames.SNSTopic),
	}
	snsSuccessState := spartaStep.NewSNSTaskState("Notify Success", snsSuccessParams)
	fargateState.Next(snsSuccessState)

	snsFailParams := spartaStep.SNSTaskParameters{
		Message:  "AWS Fargate Task started by Step Functions failed",
		TopicArn: gocf.Ref(resourceNames.SNSTopic).String(),
	}
	snsFailState := spartaStep.NewSNSTaskState("Notify Failure", snsFailParams)
	fargateState.WithCatchers(spartaStep.NewTaskCatch(
		snsFailState,
		spartaStep.StatesAll,
	))

	// Startup the machine
	stateMachineName := spartaCF.UserScopedStackName("TestStepServicesMachine")
	stateMachine := spartaStep.NewStateMachine(stateMachineName, fargateState).
		WithRoleArn(gocf.GetAtt(resourceNames.ECSRunTaskRole, "Arn"))

	// Add the state machine to the deployment...
	workflowHooks := &sparta.WorkflowHooks{
		ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
			ecrImageBuilderDecorator("spartadocker"),
			// Then build the state machine
			stateMachine.StateMachineDecorator(),
			// Then the ECS cluster that supports the Fargate task
			fargateClusterDecorator(resourceNames),
		},
	}
	return workflowHooks, nil
}
