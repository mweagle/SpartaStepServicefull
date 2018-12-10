# SpartaStepServicefull

Sparta-based application a Lambda-free service that deploys a Docker
image of this application to ECR and then to Fargate.


1. [Install Go](https://golang.org/doc/install)
1. `go get github.com/mweagle/SpartaStepServicefull`
1. `go get -u -d github.com/magefile/mage`
1. `cd ./SpartaStepServicefull`
1. `S3_BUCKET=YOUR_S3_BUCKET mage provision`
1. Visit the AWS Console and test your Step function!