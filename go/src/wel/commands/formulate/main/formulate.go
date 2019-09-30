package main

import (
	"log"
	"os"

	"wel/commands"
	"wel/formulas"
	"wel/modifiers"
)

var logger = log.New(os.Stdout, "[formulate] ", 0)

var urlParameterRegexpFragment = "([\\?].*)?"

// mimetype, file extension, is code
var staticTypes = [...][2]string{
	{"application/octet-stream", "blob"},
	{"application/x-javascript", "js"},
	{"application/javascript", "js"},
	{"application/json", "json"},
	{"application/x-font-woff", "woff"},
	{"application/font-woff2", "woff"},
	{"application/x-font-ttf", "ttf"},
	{"text/css", "css"},
	{"text/json", "json"},
	{"text/plain", "txt"},
	{"text/javascript", "js"},
	{"font/", "font"},
	{"image/", "image"},
}

var codedTypes = [...]string{
	"text/css",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
}

func main() {
	if len(os.Args) != 3 {
		printHelp()
		return
	}
	commands.EnvOverrideDotEnv(".env")

	err := formulas.Formulate(os.Args[1], os.Args[2], []modifiers.FileModifier{}, formulas.ProbeBasis{})
	if err != nil {
		logger.Println("Error:", err)
		os.Exit(1)
	}
}

func printHelp() {
	logger.Println("usage: formulate <source capture directory> <formula destination directory>")
	logger.Println("Example: formulate ./captures/2018-12-28-5C266D4F-1C03/ ./formulas/spiffy-formula/")
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
