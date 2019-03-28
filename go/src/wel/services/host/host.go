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
	//weltls "wel/tls"
)

var logger = log.New(os.Stdout, "[host] ", 0)

// The scripts for the tests
var ProbesURL = "/__wel_probes.js"

// The URL for the embedded script that is being tested
var EmbeddedScriptURL = "/__wel_embed.js"

// The URL for the embedded script that is being tested
var ControlURL = "/__wel_control"

// The resources for the prober script that runs the tests
var ProberDistPath = "fe/dist/prober"
var ProberDistURL = "/__wel/prober/"
var ProberURL = fmt.Sprintf("%vprober.js", ProberDistURL)

/*
RunHTTP brings up the page formula host service
This function blocks until the service or process is killed.
*/
func RunHTTP(port int64, formulasPath string, probesPath string, embeddedScriptPath string) {
	// Collect and contatenate the probe scripts
	probeScript, err := GenerateProbesScript(probesPath)
	if err != nil {
		log.Fatal("Could not generate probe script at path", probesPath, err)
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

	mux := http.NewServeMux()

	// Serve embedded script
	mux.HandleFunc(EmbeddedScriptURL, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		response.Write([]byte(embeddedScript))
	})

	// Serve test probes' JS
	mux.HandleFunc(ProbesURL, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		response.Write([]byte(probeScript))
	})

	// Serve prober JS that runs the tests
	mux.Handle(ProberDistURL, http.StripPrefix(ProberDistURL, http.FileServer(http.Dir(ProberDistPath))))

	formulaHost, err := NewFormulaHost(formulasPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error starting formula host: %v", err))
		return
	}

	/*
		The control web API is usually called by the runner command to change which page formula is being hosted
	*/
	mux.HandleFunc(ControlURL, func(response http.ResponseWriter, request *http.Request) {
		HandleControlRequest(response, request, formulaHost)
	})

	// Serve page formulas
	mux.Handle("/", formulaHost)

	logger.Println("Listening on", port)
	//log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), weltls.LocalhostCertPath, weltls.LocalhostKeyPath, mux))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
