/*
Host serves page formulas to browsers during an experiment.
*/
package host

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"wel/formulas"
)

var logger = log.New(os.Stdout, "[host] ", 0)

// The URL for the embedded script that is being tested
var ControlURL = "/__wel_control"

var BlankURL = "/__wel_blank"

type controlState struct {
	BaselineMode bool
}

/*
RunHTTP brings up the page formula host service
This function blocks until the service or process is killed.
*/
func RunHTTP(port int64, frontEndDistPath string, formulasPath string, probesPath string, embeddedScriptPath string) {
	// Check that the front end dist directory exists

	feDistPathInfo, err := os.Stat(frontEndDistPath)
	if err != nil {
		log.Fatal("Could not read the front end dist path:", frontEndDistPath, err)
		return
	}
	if feDistPathInfo.IsDir() == false {
		log.Fatal("The front end dist path does not lead to a directory:", frontEndDistPath)
		return
	}

	// Collect and contatenate the probe scripts
	probeScript, err := GenerateProbesScript(probesPath)
	if err != nil {
		log.Fatal("Could not generate probe script at path ", probesPath, " ", err)
		return
	}

	// Read the embedded script
	embeddedScript := []byte("// empty embedded script \n")
	if embeddedScriptPath != "" {
		embeddedScript, err = ioutil.ReadFile(embeddedScriptPath)
		if err != nil {
			log.Fatal("Could not read the embedded script:", embeddedScriptPath)
			return
		}
	}

	controlState := controlState{
		BaselineMode: false,
	}

	mux := http.NewServeMux()

	// Serve embedded script
	mux.HandleFunc(formulas.EmbeddedScriptURL, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		if controlState.BaselineMode {
			response.Write([]byte("/* baseline mode */ console.log('baseline mode embed script'); "))
		} else {
			response.Write([]byte(embeddedScript))
		}
	})

	// Serve test probes' JS
	mux.HandleFunc(formulas.ProbesURL, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		response.Write([]byte(probeScript))
	})

	// Serve prober JS that runs the tests
	mux.Handle(formulas.ProberDistURL, http.StripPrefix(formulas.ProberDistURL, http.FileServer(http.Dir(frontEndDistPath+"/prober/"))))

	formulaHost, err := NewFormulaHost(formulasPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error starting formula host: %v", err))
		return
	}

	/*
		The BlankURL serves up a mostly empty page.
		It's usually used by the `runner` as a sort of browser state cleanser between test runs in hosted pages.
	*/
	mux.HandleFunc(BlankURL, func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("<html><body><h1>Blank</h1></body></html>"))
	})

	/*
		The control web API is usually called by the runner command to change which page formula is being hosted
	*/
	mux.HandleFunc(ControlURL, func(response http.ResponseWriter, request *http.Request) {
		HandleControlRequest(response, request, formulaHost, &controlState)
	})

	// Serve page formulas
	mux.Handle("/", formulaHost)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
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
