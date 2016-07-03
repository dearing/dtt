package lib

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

// need to export
var bucket = "drone-cform-validate"
var region = "us-east-1"

var svc *cloudformation.CloudFormation
var sss *s3.S3

func init() {
	log.SetLevel(log.DebugLevel)

	ses := session.New()
	sss = s3.New(ses, &aws.Config{Region: aws.String(region)})
	svc = cloudformation.New(ses, &aws.Config{Region: aws.String(region)})

}

func Validate(url string) (err error) {

	delay := uint(1)
	for {
		_, err = svc.ValidateTemplate(&cloudformation.ValidateTemplateInput{TemplateURL: &url})
		if err != nil {
			if strings.Contains(err.Error(), `Throttling: Rate exceeded`) {
				delay = delay + 1
				time.Sleep(1 << delay * time.Millisecond)
				continue
			} else {
				return err
			}
		} else {
			break
		}
	}

	return err
}

// destroy the stack and print out the events for the curious
func Kill(stackName string) (err error) {

	events, err := svc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return
	}

	resp, err := svc.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}

}
