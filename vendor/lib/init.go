package lib

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

// need to export
var bucket = "drone-cform-validate"
var region = "us-east-1"
var fail = false
var svc *cloudformation.CloudFormation
var sss *s3.S3

func init() {
	log.SetLevel(log.DebugLevel)

	ses := session.New()
	sss = s3.New(ses, &aws.Config{Region: aws.String(region)})
	svc = cloudformation.New(ses, &aws.Config{Region: aws.String(region)})

}
