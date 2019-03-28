package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
HandleControlRequest is used as an HTTP responder function in host.RunHTTP
It implements a control API to check and set which page formula is hosted.
*/
func HandleControlRequest(response http.ResponseWriter, request *http.Request, formulaHost *FormulaHost) {
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
	} else if request.Method != "GET" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	controlResponse := &ControlResponse{
		Formulas:       []string{},
		CurrentFormula: formulaHost.CurrentFormula,
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
}

/*
A serializable data structure for responding from the control API
*/
type ControlResponse struct {
	Formulas       []string `json:"formulas"`
	CurrentFormula string   `json:"current-formula"`
}

/*
Uses the HTTP control API to request that the host change to a new page formula
*/
func RequestPageFormulaChange(httpPort int64, formulaName string) (bool, error) {

	url := fmt.Sprintf("http://127.0.0.1:%v%v", httpPort, ControlURL)

	data, err := json.Marshal(&ControlRequest{
		CurrentFormula: formulaName,
	})
	if err != nil {
		return false, err
	}

	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return false, err
	}
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	controlResponse := &ControlResponse{}
	err = json.Unmarshal(responseData, controlResponse)
	if err != nil {
		return false, err
	}
	return controlResponse.CurrentFormula == formulaName, nil
}
