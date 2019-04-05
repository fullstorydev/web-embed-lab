/*
Formulas contains utilities for reading, using, and writing page formulas.
Reading and using usually happens in the host service.
Writing usually occurs when converting a colluder session capture to an initial page formula.
*/
package formulas

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[formulas] ", 0)

// Info about the serialized file structure
var FormulaInfoFileName = "formula.json"
var StaticDirName = "static"
var TemplateDirName = "template"

// The URL for rewritten absolute URLs hosted by the FormulaHost
var AbsoluteURLRoot = "/__wel_absolute/"

/*
PageFormula holds a description of a web page and its resources hosted locally and accessed by a web browser during an experiment.
*/
type PageFormula struct {
	Comment      string            `json:"comment"`       // A human readable description
	TemplateData map[string]string `json:"template-data"` // data passed to the formula's go templates
	Routes       []Route           `json:"routes"`        // Determines what to do with incoming URL requests
	InitialPath  string            `json:"initial-path"`  // The URL path that the test runner should use for the main page of the formula
	//ProbeBases   []ProbeBasis      `json:"probe-bases"`   // Expected values for test probes used to compare new embedded scripts
}

func (formula *PageFormula) JSON() ([]byte, error) {
	return json.Marshal(formula)
}

func NewPageFormula() *PageFormula {
	return &PageFormula{
		TemplateData: map[string]string{},
		Routes:       make([]Route, 0),
		InitialPath:  "/",
		//ProbeBases:   make([]ProbeBasis, 0),
	}
}

func ParsePageFormulaInfo(inputFile *os.File) (*PageFormula, error) {
	formula := NewPageFormula()
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, formula)
	if err != nil {
		return nil, err
	}
	return formula, nil
}

// RouteType specifies how a Route will be used
type RouteType int

const (
	TemplateRoute RouteType = iota // Routes to a go template
	MockRoute                      // Routes to a JS class that acts like a web service
	ServiceRoute                   // Routes to a service URL, locally hosted or remote
	StaticRoute                    // Routes to a locally hosted static file
)

type Route struct {
	ID         string            `json:"id"`
	Path       string            `json:"path"` // A regex used route URLs
	Type       RouteType         `json:"type"`
	Value      string            `json:"value"`
	Parameters map[string]string `json:"parameters"`
	Headers    map[string]string `json:"headers"` // HTTP headers to include in the response
}

func NewRoute(id string, path string, routeType RouteType, value string) *Route {
	return &Route{
		ID:         id,
		Path:       path,
		Type:       routeType,
		Value:      value,
		Parameters: make(map[string]string, 0),
		Headers:    make(map[string]string, 0),
	}
}

/*
ProbeBasis holds expected values from test probes.
This information is usually used to check that future probes return similar values.
*/
type ProbeBasis struct {
	/*
		DOM depth & shape
		exceptions
		selector existence
	*/
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
