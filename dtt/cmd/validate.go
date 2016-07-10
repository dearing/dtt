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
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/dearing/dtt"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate template(s)",
	Long:  `Concurrently reads, uploads, validates and then deletes a series of templates.`,
	Run: func(cmd *cobra.Command, args []string) {

		var wg sync.WaitGroup
		for i := 0; i < len(args); i++ {

			wg.Add(1)

			go func(arg string) {
				defer wg.Done()
				validateCmdRun(arg)

			}(args[i])

		}
		wg.Wait()

	},
}

func validateCmdRun(args ...string) {

	for _, arg := range args {

		t := &dtt.Template{
			File:   arg,
			Bucket: "drone-cform-validate",
			Key:    uuid.NewV4().String(),
		}

		log.Debugf("READ %s", t.File)
		err := t.Read()
		if err != nil {
			log.Errorf("%s\n%s", t.File, err)
			continue
		}

		log.Debugf("UPLOAD %s", t.File)
		err = t.Upload()
		if err != nil {
			log.Errorf("%s\n%s", t.File, err)
			continue
		}

		defer func() {
			log.Debugf("DELETE %s", t.File)
			err = t.Delete()
			if err != nil {
				log.Errorf("%s\n%s", t.File, err)
			}
		}()

		log.Debugf("VALIDATE %s", t.File)
		err = t.Validate()
		if err != nil {
			log.Errorf("%s\n%s", t.File, err)
			continue
		}

		log.Info("PASS ", t.File)
	}
}

func init() {
	RootCmd.AddCommand(validateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// validateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
