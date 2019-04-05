package experiments

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

/*
TestProbe holds the indentifier and basis (aka config) for a JS test run in a hosted page formula
*/
type TestProbe struct {
	Name string `json:"name"`
	//ProbeBasis formulas.ProbeBasis
}

type PageFormulaConfiguration struct {
	Name string `json:"name"`
	//TemplateData
}

/*
Experiment pulls together a set of page formulas, test probes, and browser configurations.

An experiment is handed to the runner process which will:
- host the page formulas
- inject the test probes
- use WebDriver to run tests on specific browsers.
*/
type Experiment struct {
	PageFormulaConfigurations []PageFormulaConfiguration `json:"page-formulas"`
	TestProbes                []TestProbe                `json:"test-probes"`
	BrowserConfigurations     []map[string]string        `json:"browser-configurations"`
}

func NewExperiment() *Experiment {
	return &Experiment{
		PageFormulaConfigurations: make([]PageFormulaConfiguration, 0),
		TestProbes:                make([]TestProbe, 0),
		BrowserConfigurations:     make([]map[string]string, 0),
	}
}

func ParseExperiment(inputFile *os.File) (*Experiment, error) {
	experiment := NewExperiment()
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, experiment)
	if err != nil {
		return nil, err
	}
	return experiment, nil
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
