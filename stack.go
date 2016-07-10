package dtt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var registry = make(map[string]*cloudformation.DescribeStackResourcesOutput)

type Assertion struct {
	Target string `json:"Target"`
	Test   string `json:"Test"`
	Op     string `json:"Op"`
}

// Test represents a relationship of templates and parameters
// called in branches, concurrently
type Stack struct {
	File    string
	Body    []byte
	Name    string
	Comment string `json:"comment"`
	ID      string `json:"id"`
	Timeout int

	Template Template `json:"template"`

	Parameters []struct {
		ParameterKey     string `json:"ParameterKey"`
		ParameterValue   string `json:"ParameterValue"`
		UsePreviousValue bool   `json:"UsePreviousValue"`
	} `json:"parameters"`

	Children []Stack     `json:"children"`
	Tests    []Assertion `json:"tests"`

	Params *cloudformation.CreateStackInput
	Events *cloudformation.DescribeStackEventsOutput
}

func (s *Stack) Read() (err error) {
	s.Body, err = ioutil.ReadFile(s.File)
	return
}

func (s *Stack) Create() (err error) {
	_, err = svc.CreateStack(s.Params)
	return
}

func (s *Stack) Kill() (err error) {
	s.Events, err = svc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.Name),
	})

	_, err = svc.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(s.Name),
	})

	return
}

// iterate over the parameters from the test package
// and replace for items in the registry
// TODO: clean up
func (s *Stack) Parse() (slice []*cloudformation.Parameter) {
	for i := 0; i < len(s.Parameters); i++ {

		for k, v := range registry {

			for _, b := range v.StackResources {

				y := fmt.Sprintf("%s.%s", k, *b.LogicalResourceId)
				z := *b.PhysicalResourceId
				replacement := strings.Replace(s.Parameters[i].ParameterValue, y, z, -1)
				if replacement != s.Parameters[i].ParameterValue {
					log.Debugf("substitution '%s' => '%s'", y, replacement)
					s.Parameters[i].ParameterValue = replacement
				}

			}

		}

		slice = append(slice,
			&cloudformation.Parameter{
				ParameterKey:     &s.Parameters[i].ParameterKey,
				ParameterValue:   &s.Parameters[i].ParameterValue,
				UsePreviousValue: &s.Parameters[i].UsePreviousValue},
		)
	}
	return
}

// recursively create stacks of children and wait
func (s *Stack) Execute() (err error) {

	var wg sync.WaitGroup

	s.Template.Key = uuid.NewV4().String()
	s.Template.Bucket = "drone-cform-validate"

	err = s.Template.Read()
	if err != nil {
		return
	}

	err = s.Template.Upload()
	if err != nil {
		return
	}

	defer s.Template.Delete()

	err = s.Template.Validate()
	if err != nil {
		return
	}

	// lifted Docker's container naming because I'm lazy
	s.Name = strings.ToUpper(name())

	// parameters to create this stack by
	s.Params = &cloudformation.CreateStackInput{
		StackName: &s.Name,
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
		},
		DisableRollback: aws.Bool(true),
		Parameters:      s.Parse(),
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("drone-testing"),
				Value: aws.String(s.Name),
			},
		},
		TemplateURL:      &s.Template.URL,
		TimeoutInMinutes: aws.Int64(10),
	}

	s.Create()

	defer s.Kill()

	// useful to see the end results for debugging
	_, err = svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		return
	}

	// idle while the stack cooks
	err = svc.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		return
	}

	// the stack should be good, fetch the resources and store them in our registry
	resources, err := svc.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}
	registry[s.ID] = resources

	events, err := svc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		log.Error(err.Error())
		return
	}

	for i := 0; i < len(s.Tests); i++ {
		result := assert(events.StackEvents, s.Tests[i])
		if !result {
			fail = true
		}
	}

	for i := 0; i < len(s.Children); i++ {
		wg.Add(1)
		go func(s *Stack) {
			defer wg.Done()
			s.Execute()
		}(&s.Children[i])
	}

	wg.Wait()

	return
}

func assert(events []*cloudformation.StackEvent, assert Assertion) (result bool) {

	var v interface{}

	result = true

	data := strings.SplitN(assert.Target, ".", 2)
	log.Debugf("split %+v", data)

	for i := 0; i < len(events); i++ {
		if *events[i].ResourceStatus != "CREATE_COMPLETE" {
			continue
		}

		logicalid := events[i].LogicalResourceId
		properties := events[i].ResourceProperties

		if strings.Compare(*logicalid, data[0]) == 0 {

			err := json.Unmarshal([]byte(*properties), &v)
			if err != nil {
				log.Error(err)
			}

			m := v.(map[string]interface{})
			log.Debugf("DECODED : %+v", m)

			for k, v := range m {
				log.Debug("Key:", k, "Value:", v)
			}

			switch t := m[data[1]].(type) {
			default:
				log.Warn("Assuming map")
				switch assert.Op {
				case "in":
					if v, ok := t.(map[string]interface{}); ok {
						log.Warn(v)
					}
				}
			case string:
				switch assert.Op {
				case "eq":
					if strings.Compare(t, assert.Test) != 0 {
						result = false
					}
				case "ne":

					if strings.Compare(t, assert.Test) == 0 {
						result = false
					}
				}
				if result {
					log.Infof("PASS: '%s.%s' %s '%s' ", data[0], data[1], assert.Op, assert.Test)
				} else {
					log.Infof("FAIL: '%s.%s' %s '%s' \t GOT %v", data[0], data[1], assert.Op, assert.Test, t)
				}
			}

		}

	}

	return
}
