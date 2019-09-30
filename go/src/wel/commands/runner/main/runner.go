package main

import (
	"log"
	"os"

	"wel/commands"
	"wel/experiments"
	"wel/services/host"
	"wel/tunnels"
	"wel/webdriver"

	"github.com/logrusorgru/aurora"
)

var logger = log.New(os.Stdout, "[runner] ", 0)

var pageHostPort int64 = 9090

var frontEndDistPathVar = "FRONT_END_DIST"

/*
The runner command runs an experiment, using Selenium to run test probes in page formulas.
*/
func main() {
	commands.EnvOverrideDotEnv(".env")
	if run() {
		logger.Println(aurora.Green("*PASSED*"))
		os.Exit(0)
	} else {
		logger.Println(aurora.Red("*FAILED*"))
		os.Exit(1)
	}
}

/*
run does the work of running the experiment
The returned bool is true if all tests were run and passed
*/
func run() bool {
	/*
		Read the path of the front end dist directory
	*/
	frontEndDistPath := os.Getenv(frontEndDistPathVar)
	if frontEndDistPath == "" {
		logger.Println("Environment variable", frontEndDistPathVar, "is required")
		return false
	}

	if len(os.Args) == 3 {
		// Run in developer mode
		logger.Println("Developer mode on port", pageHostPort)
		host.RunHTTP(pageHostPort, frontEndDistPath, os.Args[1], os.Args[2], "")
	} else if len(os.Args) == 4 {
		// Run in developer mode
		logger.Println("Embed mode on port", pageHostPort)
		host.RunHTTP(pageHostPort, frontEndDistPath, os.Args[1], os.Args[2], os.Args[3])
	} else if len(os.Args) != 5 && len(os.Args) != 6 {
		printHelp()
		return false
	}

	/*
		Read the WebDriver configuration
	*/
	browserstackUser := os.Getenv(webdriver.BrowserstackUserVar)
	browserstackAPIKey := os.Getenv(webdriver.BrowserstackAPIKeyVar)
	if browserstackUser == "" || browserstackAPIKey == "" {
		logger.Println("Environment variables", webdriver.BrowserstackUserVar, "and", webdriver.BrowserstackAPIKeyVar, "are required")
		return false
	}

	formulasPath := os.Args[1]
	probesPath := os.Args[2]
	embedScriptPath := os.Args[3]
	experimentPath := os.Args[4]
	soloPageFormulaName := ""
	if len(os.Args) == 6 {
		soloPageFormulaName = os.Args[5]
	}

	/*
		Read and parse the experiment definition
	*/
	experimentFile, err := os.Open(experimentPath)
	if err != nil {
		logger.Println("Error opening experiment JSON:", experimentPath, ":", err)
		printHelp()
		return false
	}
	defer experimentFile.Close()
	experiment, err := experiments.ParseExperiment(experimentFile)
	if err != nil {
		logger.Println("Error parsing experiment JSON:", experimentPath, ":", err)
		printHelp()
		return false
	}
	ok, runnableErrorMessage := experiment.IsRunnable()
	if ok == false {
		logger.Println("Experiment is not runnable:", runnableErrorMessage)
		return false
	}

	if soloPageFormulaName != "" {
		_, ok := experiment.GetPageFormulaConfiguration(soloPageFormulaName)
		if ok == false {
			logger.Println("Unknown page formula:", soloPageFormulaName)
			printHelp()
			return false
		}
	}

	/*
		Set up the ngrok tunnel and find its HTTPS endpoint URL
	*/
	ngrokController := tunnels.NewNgrokController()
	err = ngrokController.Start(pageHostPort, "http")
	if err != nil {
		logger.Println("Could not start ngrok", err)
		return false
	}
	defer ngrokController.Stop()
	_, pageHostURL, err := ngrokController.WaitForNgrokTunnels("https")
	if err != nil {
		logger.Println("Error", err)
		return false
	}

	/*
		Start the page formula host
	*/
	go func() {
		host.RunHTTP(
			pageHostPort,
			frontEndDistPath,
			formulasPath,
			probesPath,
			embedScriptPath,
		)
	}()

	experimentConfig := experiments.ExperimentConfig{
		BrowserstackUser:   browserstackUser,
		BrowserstackAPIKey: browserstackAPIKey,
		FrontEndDistPath:   frontEndDistPath,
		PublicPageHostURL:  pageHostURL,
		PageHostPort:       pageHostPort,
	}

	/*
		Gather the baseline data without the target embed script
	*/
	baselineData, err := experiments.GatherExperimentBaseline(
		experiment,
		&experimentConfig,
		soloPageFormulaName,
	)
	if err != nil {
		logger.Println("Error gathering baseline", err)
		return false
	}
	if len(baselineData) == 0 {
		logger.Println("Zero length baseline data!")
		return false
	}

	/*
		Finally, run the experiment tests
	*/
	return experiments.RunExperimentTests(
		experiment,
		&experimentConfig,
		baselineData,
		soloPageFormulaName,
	)
}

func printHelp() {
	logger.Println("usage (experiment mode): runs the experiment")
	logger.Println(aurora.Bold("runner <formulas dir> <probes dir> <embed script> <experiment json>"))
	logger.Println("Example:")
	logger.Println("runner ../pf/ ./examples/test-probes/ ./examples/embed_scripts/no-op.js ./examples/experiments/external-experiment.json\n")

	logger.Println("usage (single page formula experiment mode): runs only one page formula in the experiment")
	logger.Println(aurora.Bold("runner <formulas dir> <probes dir> <embed script> <experiment json> <page formula name>"))
	logger.Println("Example:")
	logger.Println("runner ../pf/ ./examples/test-probes/ ./examples/embed_scripts/no-op.js ./examples/experiments/external-experiment.json transmutable-light\n")

	logger.Println("usage (development mode): runs the page formula host")
	logger.Println(aurora.Bold("runner <formulas dir> <probes dir>"))
	logger.Println("Example:")
	logger.Println("runner ../pf/ ./examples/test-probes/\n")

	logger.Println("usage (embed mode): runs the page formula host with an embed script")
	logger.Println(aurora.Bold("runner <formulas dir> <probes dir> <embed script>"))
	logger.Println("Example:")
	logger.Println("runner ../pf/ ./examples/test-probes/ ./examples/embed_scripts/no-op.js")
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
