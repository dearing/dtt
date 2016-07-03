package lib

/*
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

var fail = false

var registry = make(map[string]*cloudformation.DescribeStackResourcesOutput)

// Test represents a relationship of templates and parameters
// called in branches, concurrently
type Test struct {
	Comment    string `json:"comment"`
	ID         string `json:"id"`
	Template   string `json:"template"`
	Parameters []struct {
		ParameterKey     string `json:"ParameterKey"`
		ParameterValue   string `json:"ParameterValue"`
		UsePreviousValue bool   `json:"UsePreviousValue"`
	} `json:"parameters"`
	Children []Test      `json:"children"`
	Tests    []Assertion `json:"tests"`
}

type Assertion struct {
	Target string `json:"Target"`
	Test   string `json:"Test"`
	Op     string `json:"Op"`
}

func TestPack(args ...string) bool {

	log.Info(args)

	if len(args) < 1 {
		log.Error("require at least one test package to continue")
		fail = true
	}

	for _, file := range args[:] {

		var wg sync.WaitGroup

		registry, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error(err.Error())
			fail = true
			continue
		}

		var tests []Test

		err = json.Unmarshal(registry, &tests)
		if err != nil {
			log.Error(err.Error())
			fail = true
			continue
		}

		for i := 0; i < len(tests); i++ {
			wg.Add(1)
			go func(test *Test) {
				defer wg.Done()
				run(test)
			}(&tests[i])
		}

		wg.Wait()

	}

	if fail {
		log.Error("F̶̵̣̝̬͙͕͇̤̏ͯ̾ͣ͛͗̎͛͟A̴͚̗̒̉͌͂̎ͫI̻̤̝̖ͭ̈́̑͘͠ͅL̠̩̝͇͙ͯ͂̇̅͒")

	}
	log.Info("PASS")

	return fail

}

// recursively create stacks of children and wait
func run(test *Test) {

	var wg sync.WaitGroup

	if len(registry) != 0 {
		log.Debug(registry)
	}

	key := uuid.NewV4().String()
	log.Infof("uploading %s to %s/%s", test.Template, bucket, key)
	url, err := Upload(test.Template, bucket, key)
	if err != nil {
		log.Error(err)
		fail = true
		return
	}
	defer Delete(test.Template, bucket, key)
	defer log.Infof("deleting %s", url)

	log.Infof("validating %s", url)
	err = Validate(url)
	if err != nil {
		log.Error(err)
		fail = true
		return
	}

	// lifted Docker's container naming because I'm lazy
	stackName := strings.ToUpper(name())

	// parameters to create this stack by
	params := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
		},
		DisableRollback: aws.Bool(true),
		Parameters:      parse(test),
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("drone-testing"),
				Value: aws.String(stackName),
			},
		},
		TemplateURL:      &url,
		TimeoutInMinutes: aws.Int64(10),
	}

	log.Infof("creating stack %s from %s // %s", stackName, url, test.Comment)
	resp, err := svc.CreateStack(params)
	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}

	log.Debugf("%s\n%s", stackName, resp)
	defer Kill(stackName)
	defer log.Infof("deleting stack %s // %s", stackName, test.Comment)

	// useful to see the end results for debugging
	input, err := svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}
	log.Debugf("%s\n%s", stackName, input)

	// idle while the stack cooks
	err = svc.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}

	log.Infof("PASS: %s // %s", stackName, test.Comment)

	// the stack should be good, fetch the resources and store them in our registry
	resources, err := svc.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Error(err.Error())
		fail = true
		return
	}
	registry[test.ID] = resources

	events, err := svc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Error(err.Error())
		return
	}

	for i := 0; i < len(test.Tests); i++ {
		log.Debugf("TEST: %s : %+v", test.ID, test.Tests[i])
		result := assert(events.StackEvents, test.Tests[i])
		if result == false {
			fail = true
		}
	}

	for i := 0; i < len(test.Children); i++ {
		wg.Add(1)
		go func(test *Test) {
			defer wg.Done()
			run(test)
		}(&test.Children[i])
	}

	wg.Wait()

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

// iterate over the parameters from the test package
// and replace for items in the registry
// TODO: clean up
func parse(test *Test) (slice []*cloudformation.Parameter) {
	for i := 0; i < len(test.Parameters); i++ {

		log.Debug("Evaluating: ", test.Parameters[i].ParameterValue)

		for k, v := range registry {

			for _, b := range v.StackResources {

				y := fmt.Sprintf("%s.%s", k, *b.LogicalResourceId)
				z := *b.PhysicalResourceId
				replacement := strings.Replace(test.Parameters[i].ParameterValue, y, z, -1)
				if replacement != test.Parameters[i].ParameterValue {
					log.Debugf("substitution '%s' => '%s'", y, replacement)
					test.Parameters[i].ParameterValue = replacement
				}

			}

		}

		slice = append(slice,
			&cloudformation.Parameter{
				ParameterKey:     &test.Parameters[i].ParameterKey,
				ParameterValue:   &test.Parameters[i].ParameterValue,
				UsePreviousValue: &test.Parameters[i].UsePreviousValue},
		)
	}
	return
}
*/
