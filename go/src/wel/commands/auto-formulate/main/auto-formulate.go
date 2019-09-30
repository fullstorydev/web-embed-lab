package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"wel/commands"
	"wel/formulas"
	"wel/services/colluder"
	"wel/services/colluder/session"
	"wel/services/proxy"
	"wel/tunnels"
	"wel/webdriver"

	"github.com/logrusorgru/aurora"
	"github.com/sclevine/agouti"
)

var logger = log.New(os.Stdout, "[auto-formulate] ", 0)

func main() {
	commands.EnvOverrideDotEnv(".env")
	err := run()
	if err != nil {
		logger.Println("Error", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 {
		printHelp()
		return errors.New("Incorrect arguments")
	}

	// Read and validate the configuration JSON argument
	configFile, err := os.Open(os.Args[1])
	if err != nil {
		logger.Printf("Could not open configuration JSON: \"%v\"", os.Args[1])
		return err
	}
	defer configFile.Close()
	config, err := formulas.ParseAutoFormulate(configFile)
	if err != nil {
		logger.Printf("Could not parse configuration JSON: \"%v\": %v", os.Args[1], err)
		return err
	}
	if len(config.Captures) == 0 && len(config.Formulations) == 0 {
		return errors.New("Nothing in the config JSON to capture or formulate")
	}

	// Gather the names of all of the sites to capture
	siteNames := []string{}
	for _, capture := range config.Captures {
		for _, site := range capture.Sites {
			siteNames = append(siteNames, site.Name)
		}
	}

	// Check that each formulation references a captured site name
	for _, formulation := range config.Formulations {
		matchesCapture := false
		for _, siteName := range siteNames {
			if formulation.CaptureName == siteName {
				matchesCapture = true
				break
			}
		}
		if matchesCapture == false {
			return errors.New(fmt.Sprintf("Invalid capture name (%v) referenced by formulation (%v)", formulation.CaptureName, formulation.FormulaName))
		}
	}

	// Create the formula destination dir if necessary
	formulaDirPath := os.Args[2]
	if err = os.MkdirAll(formulaDirPath, 0777); err != nil {
		logger.Printf("Could not find or create formula path: \"%v\"", formulaDirPath)
		return err
	}

	// Read the Browserstack configuration info
	browserstackUser := os.Getenv(webdriver.BrowserstackUserVar)
	browserstackAPIKey := os.Getenv(webdriver.BrowserstackAPIKeyVar)
	if browserstackUser == "" || browserstackAPIKey == "" {
		return errors.New("Environment variables " + webdriver.BrowserstackUserVar + " and " + webdriver.BrowserstackAPIKeyVar + " are required")
	}

	// Start the colluder web app, control web socket, and HTTP proxy services
	err = colluder.PrepForCollusion()
	if err != nil {
		logger.Printf("Could not prep for collusion: %s", err)
		return nil
	}
	go colluder.RunHTTP(colluder.ColluderWebPort)
	go colluder.RunWS(colluder.ColluderWebSocketPort)
	go proxy.Run(colluder.ColluderProxyPort)

	/*
		Set up the ngrok TCP tunnel to the colluder proxy and find its endpoint URL
	*/
	ngrokController := tunnels.NewNgrokController()
	err = ngrokController.Start(int64(colluder.ColluderProxyPort), "tcp")
	if err != nil {
		logger.Println("Could not start ngrok", err)
		return err
	}
	defer ngrokController.Stop()
	_, pageHostURL, err := ngrokController.WaitForNgrokTunnels("tcp")
	if err != nil {
		logger.Println("Error", err)
		return err
	}
	pageHostURL = pageHostURL[6:] // remove the tcp:// leaving just the <hostname>:<port>

	proxyConfig := agouti.ProxyConfig{
		ProxyType: "manual",
		HTTPProxy: pageHostURL,
		SSLProxy:  pageHostURL,
	}

	captureDirPaths := map[string]string{} // capture name -> capture dir path

	for _, capture := range config.Captures {
		if len(capture.BrowserConfiguration) == 0 {
			return errors.New(fmt.Sprintf("Capture has no browser configuration: %v", capture))
		}
		if len(capture.Sites) == 0 {
			return errors.New(fmt.Sprintf("Capture has no sites: %v", capture))
		}

		// Open WebDriver connection to the browser
		logger.Println("Connecting to browser")
		capabilities := agouti.NewCapabilities()
		capabilities.With("trustAllSSLCertificates")
		capabilities.Proxy(proxyConfig)
		capabilities["browserstack.user"] = browserstackUser
		capabilities["browserstack.key"] = browserstackAPIKey
		for key, value := range capture.BrowserConfiguration {
			capabilities[key] = value
		}
		capabilities["name"] = "WEL auto-formulate"
		page, err := agouti.NewPage(webdriver.BrowserstackURL, []agouti.Option{agouti.Desired(capabilities)}...)
		if err != nil {
			return err
		}
		defer page.Destroy() // Close the WebDriver session

		hasNavigated := false // true after the WebDriver session has navigated once
		for _, site := range capture.Sites {
			siteHost, err := parseHostname(site.URL)
			if err != nil {
				logger.Println("Failed to parse URL", site.URL)
				return err
			}
			logger.Println("Capturing", site.Name, site.URL)

			if hasNavigated {
				// Empty the page to stop all previous network connections
				err = page.Navigate("about:blank")
				if err != nil {
					logger.Println("Failed to blank", err)
					return err
				}
			}

			// tell the colluder to start a capture session with the site `name`
			session.CurrentCaptureSession, err = session.NewCaptureSession(siteHost)
			if err != nil {
				logger.Printf("Error toggling on", err)
			}
			session.CurrentCaptureSession.Modifiers = site.Modifiers

			// tell the page to load the URL and wait for successful load or failure
			err = page.Navigate(site.URL)
			if err != nil {
				logger.Println("Failed to navigate to", site.URL, err)
				return err
			}
			hasNavigated = true

			// Pause if the config has a close-pause
			if site.ClosePause > 0 {
				time.Sleep(time.Duration(site.ClosePause) * time.Second)
			}

			// Get the hostname in case it was changed via redirect
			newURL, err := page.URL()
			if err != nil {
				logger.Println("Failed to get the current URL for", site.URL, err)
				return err
			}
			newSiteHost, err := parseHostname(newURL)
			if err != nil {
				logger.Println("Failed to parse the current URL", newURL, err)
				return err
			}
			session.CurrentCaptureSession.Timeline.Hostname = newSiteHost

			// tell the colluder to stop the capture session
			err = session.CurrentCaptureSession.WriteTimeline()
			if err != nil {
				logger.Printf("Error writing timeline %v", err)
				return err
			}
			captureDirPaths[site.Name] = session.CurrentCaptureSession.DirectoryPath
			session.CurrentCaptureSession = nil
		}
	}

	for _, formulation := range config.Formulations {
		captureDirPath, ok := captureDirPaths[formulation.CaptureName]
		if ok == false {
			return errors.New(fmt.Sprintf("Invalid capture name (%v) in formulation (%v)", formulation.CaptureName, formulation.FormulaName))
		}
		logger.Println("Formulating", formulation.FormulaName)
		formulaPath := path.Join(formulaDirPath, formulation.FormulaName)
		err = formulas.Formulate(captureDirPath, formulaPath, formulation.Modifiers, formulation.ProbeBasis)
		if err != nil {
			logger.Println("Error formulating:", err)
			return err
		}
	}

	return nil
}

func printHelp() {
	logger.Println("usage:")
	logger.Println(aurora.Bold("auto-formulate <configuration json path> <formula destination dir>"))
	logger.Println("Example:")
	logger.Println("auto-formulate ./examples/auto-formulate/external-auto-formulate.json ../pf/\n")
}

func parseHostname(siteURL string) (string, error) {
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		logger.Println("Failed to parse URL", siteURL)
		return "", err
	}
	siteHost := parsedURL.Host
	if colonIndex := strings.Index(siteHost, ":"); colonIndex > 0 {
		siteHost = siteHost[:colonIndex]
	}
	return siteHost, nil
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
