// Copyright © 2016 Jacob Dearing <jacob.dearing@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/spf13/cobra"

	log "github.com/Sirupsen/logrus"

	"encoding/json"
	"io/ioutil"
	"sync"

	"lib"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		testCmdRun(args...)
	},
}

func testCmdRun(args ...string) {

	var fail = false

	for _, arg := range args {

		var wg sync.WaitGroup

		registry, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Error(err.Error())
			fail = true
			continue
		}

		var tests []lib.Suite

		err = json.Unmarshal(registry, &tests)
		if err != nil {
			log.Error(err.Error())
			fail = true
			continue
		}

		//log.Debugf("%+v", tests)

		for i := 0; i < len(tests); i++ {
			wg.Add(1)
			go func(s *lib.Suite) {
				defer wg.Done()
				err := s.Execute()
				if err != nil {
					log.Error(err)
				}
			}(&tests[i])
		}

		wg.Wait()

	}

	if fail {
		log.Error("F̶̵̣̝̬͙͕͇̤̏ͯ̾ͣ͛͗̎͛͟A̴͚̗̒̉͌͂̎ͫI̻̤̝̖ͭ̈́̑͘͠ͅL̠̩̝͇͙ͯ͂̇̅͒")

	}
	log.Info("PASS")

}

func run() {

}

func init() {
	RootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
