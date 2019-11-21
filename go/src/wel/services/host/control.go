package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"wel/formulas"
)

/*
HandleControlRequest is used as an HTTP responder function in host.RunHTTP
It implements a control API to check and set which page formula is hosted.
*/
func HandleControlRequest(response http.ResponseWriter, request *http.Request, formulaHost *FormulaHost, ctrlState *controlState) {
	if request.Method == "PUT" {
		requestBodyData, err := ioutil.ReadAll(request.Body)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}
		controlRequest := &ControlRequest{}
		err = json.Unmarshal(requestBodyData, controlRequest)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}
		if controlRequest.CurrentFormula != "" {
			formulaHost.SetCurrentFormula(controlRequest.CurrentFormula)
		}
		if controlRequest.HeadSnippet != "" {
			formulaHost.HeadSnippet = controlRequest.HeadSnippet
		}
		ctrlState.BaselineMode = controlRequest.BaselineMode
	} else if request.Method != "GET" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	controlResponse := &ControlResponse{
		Formulas:       []string{},
		CurrentFormula: formulaHost.CurrentFormula,
		InitialPath:    formulaHost.PageFormulas[formulaHost.CurrentFormula].InitialPath,
		ProbeBasis:     formulaHost.PageFormulas[formulaHost.CurrentFormula].ProbeBasis,
		BaselineMode:   ctrlState.BaselineMode,
		HeadSnippet:    formulaHost.HeadSnippet,
	}

	for formulaName := range formulaHost.PageFormulas {
		controlResponse.Formulas = append(controlResponse.Formulas, formulaName)
	}
	responseData, err := json.Marshal(controlResponse)
	if err != nil {
		logger.Println("Error", err)
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}
	response.Write([]byte(responseData))
}

/*
A parsable data structure for reading control API PUTs
*/
type ControlRequest struct {
	CurrentFormula string `json:"current-formula"`
	BaselineMode   bool   `json:"baseline-mode"`
	HeadSnippet    string `json:"head-snippet"`
}

/*
A serializable data structure for responding from the control API
*/
type ControlResponse struct {
	Formulas       []string            `json:"formulas"`
	CurrentFormula string              `json:"current-formula"`
	InitialPath    string              `json:"initial-path"`
	ProbeBasis     formulas.ProbeBasis `json:"probe-basis"`
	BaselineMode   bool                `json:"baseline-mode"`
	HeadSnippet    string              `json:"head-snippet"`
}

/*
Uses the HTTP control API to request that the host change to a new page formula
*/
func RequestPageFormulaChange(httpPort int64, formulaName string, baselineMode bool, headSnippet string) (bool, *ControlResponse, error) {

	url := fmt.Sprintf("http://127.0.0.1:%v%v", httpPort, ControlURL)

	data, err := json.Marshal(&ControlRequest{
		CurrentFormula: formulaName,
		BaselineMode:   baselineMode,
		HeadSnippet:    headSnippet,
	})
	if err != nil {
		return false, nil, err
	}

	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return false, nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return false, nil, err
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, nil, err
	}
	controlResponse := &ControlResponse{}
	err = json.Unmarshal(responseData, controlResponse)
	if err != nil {
		return false, nil, err
	}
	if controlResponse.CurrentFormula == formulaName && controlResponse.BaselineMode == baselineMode {
		return true, controlResponse, nil
	}
	return false, controlResponse, nil
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
