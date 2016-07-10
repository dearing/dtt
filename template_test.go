package dtt

// TODO: adopt testify
import (
	"os"
	"testing"
)

type templateFixture struct {
	Template
	ErrorExpected bool
}

var templates = []templateFixture{
	{
		Template: Template{
			File:   "fixture/doest_exist.template",
			Bucket: "drone-cform-validate",
			Key:    "doest_exist.template",
		},
		ErrorExpected: true,
	}, {
		Template: Template{
			File:   "fixture/passing.template",
			Bucket: "drone-cform-validate",
			Key:    "passing.template",
		},
		ErrorExpected: false,
	}, {
		Template: Template{
			File:   "fixture/failing.template",
			Bucket: "drone-cform-validate",
			Key:    "failing.template",
		},
		ErrorExpected: true,
	}, {
		Template: Template{
			File:   "fixture/not_a.template",
			Bucket: "drone-cform-validate",
			Key:    "not_a.template",
		},
		ErrorExpected: true,
	}, {
		Template: Template{
			File:   "fixture/minified.template",
			Bucket: "drone-cform-validate",
			Key:    "minified.template",
		},
		ErrorExpected: false,
	}, {
		Template: Template{
			File:   "fixture/malformed_json.template",
			Bucket: "drone-cform-validate",
			Key:    "malformed_json.template",
		},
		ErrorExpected: true,
	},
}

func TestValidate(t *testing.T) {
	for _, tt := range templates {

		err := tt.Read()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}

		err = tt.Upload()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}
		err = tt.Validate()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}
		err = tt.Delete()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}

		tt.File = os.TempDir() + "temp.json"
		err = tt.Write()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}

		os.Remove(os.TempDir() + "/temp.json")
	}
}

func TestPrettyPrint(t *testing.T) {
	for _, tt := range templates {

		err := tt.Read()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}

		err = tt.PrettyPrint()
		if err != nil {
			if !tt.ErrorExpected {
				t.Log(tt.File, "\n", err)
				t.Fail()
			}
		}

	}
}
