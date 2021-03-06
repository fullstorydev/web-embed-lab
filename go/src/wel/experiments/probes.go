package experiments

func getBool(data map[string]interface{}, name string) (value bool, ok bool) {
	if interfaceVal, ok := data[name]; ok == true {
		value, ok := interfaceVal.(bool)
		if ok {
			return value, ok
		}
		return false, false
	}
	return false, false
}

func getInt64(data map[string]interface{}, name string) (value int64, ok bool) {
	if interfaceVal, ok := data[name]; ok == true {
		value, ok := interfaceVal.(int64)
		if ok {
			return value, ok
		}
		return 0, false
	}
	return 0, false
}

/*
BaselineData holds information gathered from a page formula sampled without a target embed script.
It is used by test probes during a run of an experiment to compare values with and without a target embed script.
*/
type BaselineData map[string]interface{}

func (baselineData BaselineData) Successful() bool {
	if value, ok := baselineData.GetBool("success"); ok == true {
		return value
	}
	return false
}

func (baselineData BaselineData) GetBool(name string) (value bool, ok bool) {
	return getBool(baselineData, name)
}

func (baselineData BaselineData) GetInt64(name string) (value int64, ok bool) {
	return getInt64(baselineData, name)
}

type BaselineDataList map[string]BaselineData

func (baselineDataList BaselineDataList) Successful() bool {
	for _, baselineData := range baselineDataList {
		if baselineData.Successful() == false {
			return false
		}
	}
	return true
}

/*
ProbeResult holds the final results of a test probe during a run of an experiment
*/
type ProbeResult map[string]interface{}

func (probeResult ProbeResult) Passed() bool {
	if value, ok := probeResult.GetBool("passed"); ok == true {
		return value
	}
	return false
}

func (probeResult ProbeResult) GetBool(name string) (value bool, ok bool) {
	return getBool(probeResult, name)
}

func (probeResult ProbeResult) GetInt64(name string) (value int64, ok bool) {
	return getInt64(probeResult, name)
}

type ProbeResults map[string]ProbeResult

func (probeResults ProbeResults) Passed() bool {
	for _, probeResult := range probeResults {
		if probeResult.Passed() == false {
			return false
		}
	}
	return true
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
