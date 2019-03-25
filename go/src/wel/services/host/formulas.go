package host

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"wel/formulas"
)

type FormulaHost struct {
	FormulasPath   string
	CurrentFormula string
	PageFormulas   map[string]*formulas.PageFormula
}

func NewFormulaHost(formulasPath string) (*FormulaHost, error) {
	formulasStat, err := os.Stat(formulasPath)
	if err != nil {
		return nil, err
	}
	if formulasStat.IsDir() == false {
		return nil, errors.New("Formulas path does not lead to a directory")
	}

	files, err := ioutil.ReadDir(formulasPath)
	if err != nil {
		return nil, err
	}

	host := &FormulaHost{
		FormulasPath:   formulasPath,
		CurrentFormula: "",
		PageFormulas:   map[string]*formulas.PageFormula{},
	}

	var firstName = ""

	for _, file := range files {
		if file.IsDir() == false {
			continue
		}
		infoPath := path.Join(formulasPath, file.Name(), formulas.FormulaInfoFileName)
		infoStat, err := os.Stat(infoPath)
		if err != nil {
			logger.Printf("Found no %v", infoPath)
			continue
		}
		if infoStat.IsDir() {
			logger.Printf("Found a dir at %v", infoPath)
			continue
		}
		infoFile, err := os.Open(infoPath)
		if err != nil {
			logger.Printf("Could not open %v", infoPath)
			continue
		}
		defer func() {
			infoFile.Close()
		}()
		pageFormula, err := formulas.ParsePageFormulaInfo(infoFile)
		if err != nil {
			logger.Printf("Could not parse %v: %v", file.Name(), err)
			continue
		}
		host.PageFormulas[file.Name()] = pageFormula
		if firstName == "" {
			firstName = file.Name()
		}
	}

	if len(host.PageFormulas) == 0 {
		return nil, errors.New("Read no page formulas")
	}
	host.CurrentFormula = firstName
	logger.Println("Hosting", firstName)

	return host, nil
}

func (host *FormulaHost) SetCurrentFormula(name string) bool {
	for formulaName := range host.PageFormulas {
		if name == formulaName {
			host.CurrentFormula = formulaName
			return true
		}
	}
	return false
}

func (host *FormulaHost) MatchRoute(path string) *formulas.Route {
	if host.CurrentFormula == "" {
		return nil
	}
	for _, route := range host.PageFormulas[host.CurrentFormula].Routes {
		matched, err := regexp.MatchString(route.Path, path)
		if err != nil {
			logger.Printf("Error in regexp \"%v\": %v", route.Path, err)
			return nil
		}
		if matched {
			return &route
		}
	}
	return nil
}

func (host *FormulaHost) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	route := host.MatchRoute(request.URL.Path)
	if route == nil {
		if strings.HasPrefix(request.URL.Path, fmt.Sprintf("/%v/", formulas.StaticDirName)) && strings.Contains(request.URL.Path, "..") == false {
			host.handleStaticRequest(writer, request)
			return
		}
		logger.Println("No route", request.URL.Path)
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(fmt.Sprintf("No route: %v", request.URL.Path)))
		return
	}
	switch route.Type {
	case formulas.TemplateRoute:
		host.handleTemplateRoute(route, writer, request)
	case formulas.StaticRoute:
		host.handleStaticRoute(route, writer, request)
	default:
		logger.Println("Unknown route type", route.Type)
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(fmt.Sprintf("Not found: %v", request.URL.Path)))
	}
}

func (host *FormulaHost) handleStaticRequest(writer http.ResponseWriter, request *http.Request) {
	logger.Println("Static request", request.URL.Path)
	blob, blobStat, err := host.openBlob(request.URL.Path)
	if err != nil {
		logger.Println("No static blob", request.URL.Path, err)
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Blob error"))
		return
	}
	defer func() {
		blob.Close()
	}()

	writer.WriteHeader(http.StatusOK)
	if _, err := io.CopyN(writer, blob, blobStat.Size()); err != nil {
		logger.Printf("Error writing blob: %s", err)
	}
}

func (host *FormulaHost) handleStaticRoute(route *formulas.Route, writer http.ResponseWriter, request *http.Request) {
	blob, blobStat, err := host.openBlob(route.Value)
	if err != nil {
		logger.Println("No routed static blob", route.Value, err)
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Blob error"))
		return
	}
	defer func() {
		blob.Close()
	}()

	for key, value := range route.Headers {
		writer.Header().Add(key, value)
	}
	writer.WriteHeader(http.StatusOK)

	if _, err := io.CopyN(writer, blob, blobStat.Size()); err != nil {
		logger.Printf("Error writing blob: %s", err)
	}
}

func (host *FormulaHost) handleTemplateRoute(route *formulas.Route, writer http.ResponseWriter, request *http.Request) {
	routeTemplate, err := template.ParseFiles(path.Join(host.FormulasPath, host.CurrentFormula, route.Value))
	if err != nil {
		logger.Println("Template error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Template error"))
		return
	}
	err = routeTemplate.Execute(writer, route.Parameters)
	if err != nil {
		logger.Println("Template execution error", err)
		writer.Write([]byte("Template failure"))
	}
}

func (host *FormulaHost) openBlob(pathValue string) (*os.File, os.FileInfo, error) {
	blobPath := path.Join(host.FormulasPath, host.CurrentFormula, pathValue)
	blobStat, err := os.Stat(blobPath)
	if err != nil {
		return nil, nil, err
	}
	if blobStat.IsDir() {
		return nil, nil, errors.New("Unexpectedly found a directory")
	}
	blob, err := os.Open(blobPath)
	if err != nil {
		return nil, nil, err
	}
	return blob, blobStat, nil
}
