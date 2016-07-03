package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Template struct {
	Body   []byte
	Bucket string
	Differ bool
	File   string
	Key    string
	Pretty bool
	URL    string
}

// Read a file from disk into BODY
func (t *Template) Read() (err error) {
	t.Body, err = ioutil.ReadFile(t.File)
	return
}

// Write a file to disk of BODY
func (t *Template) Write() (err error) {
	err = ioutil.WriteFile(t.File, t.Body, 0644)
	return
}

// Upload this template to S3 based on Key and Bucket
func (t *Template) Upload() (err error) {

	// sniff out that this is a template be fore we muck with uploading it
	if !strings.Contains(string(t.Body), "AWSTemplateFormatVersion") {
		return fmt.Errorf("%s does not appear to be a cloudformation template.\n", t.File)
	}
	// create object in the bucket with contents
	_, err = sss.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(string(t.Body)),
		Bucket: &t.Bucket,
		Key:    &t.Key,
	})

	t.URL = fmt.Sprintf(`https://s3.amazonaws.com/%s/%s`, t.Bucket, t.Key)

	return err
}

// Delete this S3 object based on Bucket and Key
func (t *Template) Delete() (err error) {

	// delete the object at the end
	_, err = sss.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &t.Bucket,
		Key:    &t.Key,
	})

	return err

}

// PrettyPrint JSON to BODY from BODY
func (t *Template) PrettyPrint() (err error) {

	var data bytes.Buffer

	err = json.Indent(&data, t.Body, "", "  ")
	if err != nil {
		t.Pretty = false
		return err
	}

	t.Pretty = true
	return err
}

// Validate this template by URL
func (t *Template) Validate() (err error) {

	delay := uint(1)
	for {
		_, err = svc.ValidateTemplate(&cloudformation.ValidateTemplateInput{TemplateURL: &t.URL})
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
