package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"wel/commands/runner"
	"wel/experiments"
	"wel/services/host"

	"github.com/sclevine/agouti"
)

var logger = log.New(os.Stdout, "[runner] ", 0)

var runnerPort int64 = 8090

var browserstackURL = "http://hub-cloud.browserstack.com/wd/hub"
var browserstackUserVar = "BROWSERSTACK_USER"
var browserstackAPIKeyVar = "BROWSERSTACK_API_KEY"
var frontEndDistPathVar = "FRONT_END_DIST"

/*
The runner command runs an experiment, using Selenium to run test probes in page formulas.
*/
func main() {
	/*
		Read the path of the front end dist directory
	*/
	frontEndDistPath := os.Getenv(frontEndDistPathVar)
	if frontEndDistPath == "" {
		logger.Println("Environment variable", frontEndDistPathVar, "is required")
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		// Run in developer host mode
		host.RunHTTP(runnerPort, frontEndDistPath, os.Args[1], os.Args[2], "")
	} else if len(os.Args) != 5 {
		printHelp()
		os.Exit(1)
		return
	}

	/*
		Read the WebDriver configuration
	*/
	browserstackUser := os.Getenv(browserstackUserVar)
	browserstackAPIKey := os.Getenv(browserstackAPIKeyVar)
	if browserstackUser == "" || browserstackAPIKey == "" {
		logger.Println("Environment variables", browserstackUserVar, "and", browserstackAPIKeyVar, "are required")
		os.Exit(1)
		return
	}

	/*
		Read and parse the experiment definition
	*/
	experimentPath := os.Args[3]
	experimentFile, err := os.Open(experimentPath)
	if err != nil {
		logger.Println("Error opening experiment JSON:", experimentPath, ":", err)
		printHelp()
		os.Exit(1)
		return
	}
	defer experimentFile.Close()
	experiment, err := experiments.ParseExperiment(experimentFile)
	if err != nil {
		logger.Println("Error parsing experiment JSON:", experimentPath, ":", err)
		printHelp()
		os.Exit(1)
		return
	}

	/*
		Start the page formula host
	*/
	go func() {
		host.RunHTTP(runnerPort, frontEndDistPath, os.Args[1], os.Args[2], os.Args[4])
	}()

	/*
		Set up the ngrok tunnel
	*/
	ngrokController := runner.NewNgrokController()
	err = ngrokController.Start(runnerPort)
	if err != nil {
		logger.Println("Could not start ngrok", err)
		os.Exit(1)
		return
	}
	defer ngrokController.Stop()

	logger.Println("Waiting for ngrok tunnels")
	var tunnels *runner.NgrokTunnels = nil
	tryCount := 0
	pageHostURL := ""
	for {
		if tryCount > 100 {
			logger.Println("Could not read ngrok process")
			os.Exit(1)
		}
		tryCount += 1
		// wait for ngrok to start or fail
		time.Sleep(100 * time.Millisecond)
		if ngrokController.Command == nil {
			continue
		}
		if ngrokController.Command.ProcessState != nil {
			logger.Println("ngrok process ended")
			os.Exit(1)
			return
		}
		tunnels, err = runner.FetchNgrokTunnels()
		if err != nil {
			continue
		}
		if len(tunnels.Tunnels) == 2 {
			if tunnels.Tunnels[0].Protocol == "https" {
				pageHostURL = tunnels.Tunnels[0].PublicURL
			} else if tunnels.Tunnels[1].Protocol == "https" {
				pageHostURL = tunnels.Tunnels[1].PublicURL
			} else {
				logger.Println("No ngrok tunnel is https")
				os.Exit(1)
				return
			}
			logger.Println("Found ngrok tunnel:", pageHostURL)
			break
		}
	}

	/*
		For each {browser config, page formula} combination:
		- point the browser at the host
		- run the test probes
	*/
	gatheredResults := []*runner.ProbeResults{}
	gatheredReturnValues := []string{}
	for _, browserConfiguration := range experiment.BrowserConfigurations {
		logger.Println("Spinning up", browserConfiguration["browserName"], "via selenium")
		capabilities := agouti.NewCapabilities()
		capabilities["browserstack.user"] = browserstackUser
		capabilities["browserstack.key"] = browserstackAPIKey
		for key, value := range browserConfiguration {
			capabilities[key] = value
		}
		page, err := agouti.NewPage(browserstackURL, []agouti.Option{agouti.Desired(capabilities)}...)
		if err != nil {
			logger.Println("Failed to open selenium:", err)
			os.Exit(1)
			return
		}
		logger.Println("Opened", browserConfiguration["browserName"])

		hasNavigated := false
		for _, pageFormulaConfig := range experiment.PageFormulaConfigurations {
			logger.Println("Hosting page formula:", pageFormulaConfig.Name)

			// Tell the host which page formula to use
			formulaSet, initialPath, err := host.RequestPageFormulaChange(runnerPort, pageFormulaConfig.Name)
			if err != nil {
				logger.Println("Failed to reach host control API", err)
				os.Exit(1)
				return
			}
			if formulaSet == false {
				logger.Println("Failed to host page formula", pageFormulaConfig.Name)
				os.Exit(1)
				return
			}

			logger.Println("Testing...")
			if hasNavigated {
				err = page.Reset()
				if err != nil {
					logger.Println("Failed to reset page", err)
					os.Exit(1)
					return
				}
			}
			logger.Println("Navigating to:", pageHostURL+initialPath)
			err = page.Navigate(pageHostURL + initialPath)
			if err != nil {
				logger.Println("Failed to navigate to hosted page formula", err)
				os.Exit(1)
				return
			}
			hasNavigated = true

			var returnValue string
			page.RunScript("return JSON.stringify(runWebEmbedLabProbes());", map[string]interface{}{}, &returnValue)
			probeResults := &runner.ProbeResults{}
			err = json.Unmarshal([]byte(returnValue), probeResults)
			if err != nil {
				logger.Println("Error parsing probes result", err, returnValue)
				time.Sleep(100 * time.Minute)
				os.Exit(1)
				return
			} else {
				gatheredResults = append(gatheredResults, probeResults)
				gatheredReturnValues = append(gatheredReturnValues, returnValue)
			}
		}
	}

	logger.Println("Gathered return values", gatheredReturnValues)

	hasFailure := false
	for index, probeResults := range gatheredResults {
		if probeResults.Passed() == false {
			logger.Println("Failed:", gatheredReturnValues[index])
			hasFailure = true
		}
	}
	if hasFailure {
		os.Exit(1)
	}
	os.Exit(0)
}

func printHelp() {
	logger.Println("usage: runner <formulas dir> <probes dir> <experiment json> <embed script>")
	logger.Println("Example: runner ./examples/page-formulas/ ./examples/test-probes/ ./examples/experiments/hello-world.json ./examples/embed_scripts/no-op.js\n")
	logger.Println("usage (development mode): runner <formulas dir> <probes dir>")
	logger.Println("Example: runner ./examples/page-formulas/ ./examples/test-probes/")
}

/*
Copyright 2019 FullStory, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software
and associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
