package host

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleControlRequest(response http.ResponseWriter, request *http.Request, formulaHost *FormulaHost) {
	logger.Println("Handling control")

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
		logger.Println("PUT", controlRequest)
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
