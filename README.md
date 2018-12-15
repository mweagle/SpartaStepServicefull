# SpartaStepServicefull

Sparta-based application a Lambda-free service that deploys a Docker
image of this application to ECR and then to Fargate.


1. [Install Go](https://golang.org/doc/install)
1. `go get github.com/mweagle/SpartaStepServicefull`
1. `go get -u -d github.com/magefile/mage`
1. `cd ./SpartaStepServicefull`
1. `S3_BUCKET=YOUR_S3_BUCKET mage provision`
1. Visit the AWS Console and test your Step function!

<div align="center">
<img src="https://raw.githubusercontent.com/mweagle/SpartaStepServicefull/master/step_functions_fargate.jpg" />
</div>

## Example

```
$ mage provision
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.8.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 7074f3a
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: ServicefulStepFunction               LinkFlags= Option=provision UTC="2018-12-15T00:35:01Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Using `git` SHA for StampedBuildID            Command="git rev-parse HEAD" SHA=3164d6bdaf4e772be9a1b187659804eebf6044b3
INFO[0000] Provisioning service                          BuildID=3164d6bdaf4e772be9a1b187659804eebf6044b3 CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
WARN[0000] No lambda functions provided to Sparta.Provision()
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=0
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0001] Creating code ZIP archive for upload          TempName=./.sparta/ServicefulStepFunction-code.zip
INFO[0001] Bypassing S3 upload as no Lambda functions were provided
INFO[0001] Calling WorkflowHook                          ServiceDecoratorHook=github.com/mweagle/SpartaStepServicefull/bootstrap.ecrImageBuilderDecorator.func1 WorkflowHookContext="map[]"
INFO[0001] Docker version 18.09.0, build 4d60db4
INFO[0001] Running `go generate`
INFO[0001] Compiling binary                              Name=ServicefulStepFunction-1544834103480464000-docker.lambda.amd64
INFO[0003] Creating Docker image                         Tags="map[servicefulstepfunction:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103]"
 NFO[0004] Sending build context to Docker daemon  38.97MB
INFO[0004] Step 1/5 : FROM alpine:3.8
INFO[0004]  ---> 196d12cf6ab1
INFO[0004] Step 2/5 : RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
INFO[0004]  ---> Using cache
INFO[0004]  ---> 99402375b7f2
INFO[0004] Step 3/5 : ARG SPARTA_DOCKER_BINARY
INFO[0004]  ---> Using cache
INFO[0004]  ---> a44d27522c40
INFO[0004] Step 4/5 : ADD $SPARTA_DOCKER_BINARY /SpartaServicefull
INFO[0004]  ---> 88b3c7760b7c
INFO[0004] Step 5/5 : CMD ["/SpartaServicefull", "fargateTask"]
INFO[0004]  ---> Running in 1a0d36d7eb9a
INFO[0005] Removing intermediate container 1a0d36d7eb9a
INFO[0005]  ---> 65836c009b8c
INFO[0005] Successfully built 65836c009b8c
INFO[0005] Successfully tagged servicefulstepfunction:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103
INFO[0006] The push refers to repository [123412341234.dkr.ecr.us-west-2.amazonaws.com/spartadocker]
INFO[0006] 3f16d5a6bf23: Preparing
INFO[0006] 651f20095279: Preparing
INFO[0006] df64d3292fd6: Preparing
INFO[0006] denied: Your Authorization Token has expired. Please run 'aws ecr get-login --no-include-email' to fetch a new one.
INFO[0006] ECR push failed - reauthorizing               Error="exit status 1"
INFO[0006] Login Succeeded
INFO[0006] The push refers to repository [123412341234.dkr.ecr.us-west-2.amazonaws.com/spartadocker]
INFO[0007] 3f16d5a6bf23: Preparing
INFO[0007] 651f20095279: Preparing
INFO[0007] df64d3292fd6: Preparing
INFO[0007] df64d3292fd6: Layer already exists
INFO[0007] 651f20095279: Layer already exists
INFO[0010] 3f16d5a6bf23: Pushed
INFO[0011] 3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103: digest: sha256:d9563438eed9c72064759010f33e0f90b4a46c2922236e96df767724daf0b019 size: 949
INFO[0011] Docker image pushed                           ECRUrl="123412341234.dkr.ecr.us-west-2.amazonaws.com/spartadocker:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103"
INFO[0011] Calling WorkflowHook                          ServiceDecoratorHook="github.com/mweagle/Sparta/aws/step.(*StateMachine).StateMachineNamedDecorator.func1" WorkflowHookContext="map[imageTag:servicefulstepfunction:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103 imageURL:123412341234.dkr.ecr.us-west-2.amazonaws.com/spartadocker:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103]"
INFO[0011] Calling WorkflowHook                          ServiceDecoratorHook=github.com/mweagle/SpartaStepServicefull/bootstrap.fargateClusterDecorator.func1 WorkflowHookContext="map[imageTag:servicefulstepfunction:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103 imageURL:123412341234.dkr.ecr.us-west-2.amazonaws.com/spartadocker:3164d6bdaf4e772be9a1b187659804eebf6044b3.1544834103]"
INFO[0011] Uploading local file to S3                    Bucket=weagle Key=ServicefulStepFunction/ServicefulStepFunction-cftemplate-b51a9d2705a96ffdfaf56ff96e294d4b40a2fd9c.json Path=./.sparta/ServicefulStepFunction-cftemplate.json Size="7.0 kB"
INFO[0012] Issued CreateChangeSet request                StackName=ServicefulStepFunction
INFO[0019] Issued ExecuteChangeSet request               StackName=ServicefulStepFunction
INFO[0088] CloudFormation Metrics ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0088]     Operation duration                        Duration=61.11s Resource=ServicefulStepFunction Type="AWS::CloudFormation::Stack"
INFO[0088]     Operation duration                        Duration=8.85s Resource=ECSRunTaskSyncExecutionRole1bfabbcf3865ec2714d24bd657f707d2cf716985 Type="AWS::IAM::Role"
INFO[0088]     Operation duration                        Duration=1.75s Resource=StateMachine59f153f18068faa0b7fb588350be79df422ba5ef Type="AWS::StepFunctions::StateMachine"
INFO[0088]     Operation duration                        Duration=0.45s Resource=ECSTaskDefinitiona6848dd088648ebe9a15a6cf6a9b42c707b9a80d Type="AWS::ECS::TaskDefinition"
INFO[0088] Stack provisioned                             CreationTime="2018-12-13 06:36:25.423 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/ServicefulStepFunction/69d1ff80-fea1-11e8-849b-500c32c86c29" StackName=ServicefulStepFunction
INFO[0088] ════════════════════════════════════════════════
INFO[0088] ServicefulStepFunction Summary
INFO[0088] ════════════════════════════════════════════════
INFO[0088] Verifying IAM roles                           Duration (s)=0
INFO[0088] Verifying AWS preconditions                   Duration (s)=0
INFO[0088] Creating code bundle                          Duration (s)=2
INFO[0088] Uploading code                                Duration (s)=0
INFO[0088] Ensuring CloudFormation stack                 Duration (s)=87
INFO[0088] Total elapsed time                            Duration (s)=89
```