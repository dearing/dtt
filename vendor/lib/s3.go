package lib

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Upload(file, bucket, key string) (url string, err error) {
	// read out file
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	body := string(bytes)

	// sniff out that this is a template be fore we muck with uploading it
	if !strings.Contains(body, "AWSTemplateFormatVersion") {
		return "", fmt.Errorf("%s does not appear to be a cloudformation template.\n", file)
	}

	// create object in the bucket with contents
	_, err = sss.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(body),
		Bucket: &bucket,
		Key:    &key,
	})

	url = fmt.Sprintf(`https://s3.amazonaws.com/%s/%s`, bucket, key)
	return url, err
}

func Delete(file, bucket, key string) (err error) {

	ses := session.New()
	sss := s3.New(ses, &aws.Config{Region: aws.String(region)})

	// delete the object at the end
	_, err = sss.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	return err

}
