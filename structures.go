package workflowengine

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// JSONObject simple structure for utilizing a JSON Object
type JSONObject map[string]interface{}

// JSONObjectArray simple structure for utilizing a JSON Array of Objects
type JSONObjectArray []JSONObject

// ActionFunction Function that sends the step definition and working payload
type ActionFunction func(JSONObject, *JSONObject)

// DecisionFunction Function that sends the working payload and returns a boolean
type DecisionFunction func(JSONObject) bool

// WorkflowCallback Function that takes the working payload and performs custom actions on it every time the workflow is ran
type WorkflowCallback func(*JSONObject)

// WorkflowDefinition an individual definition of an action and/or decision
type WorkflowDefinition struct {
	Section         string           `json:"section"`
	Method          string           `json:"method"`
	Name            string           `json:"name"`
	Action          ActionFunction   `json:"-"`
	ActionDefined   bool             `json:"action"`
	Decision        DecisionFunction `json:"-"`
	DecisionDefined bool             `json:"decision"`
}

// WorkflowDefinitionSet an indexed set of Workflow Definitions
type WorkflowDefinitionSet map[string]WorkflowDefinition

type workflowResponse struct {
	Route      string `json:"route"`
	StepID     string `json:"stepID"`
	Label      string `json:"label"`
	Icon       string `json:"icon"`
	Style      string `json:"style"`
	FieldCheck string `json:"fieldCheck,omitempty"`
	Comment    string `json:"_comment,omitempty"`
}

// StepStructure is the definition of how steps are used
type StepStructure struct {
	ID            string             `json:"id"`
	Route         string             `json:"route"`
	Template      string             `json:"template"`
	UserInput     bool               `json:"userInput"`
	Payload       json.RawMessage    `json:"payload"`
	Responses     []workflowResponse `json:"response"`
	Label         string             `json:"label,omitempty"`
	Am            string             `json:"am,omitempty"`
	Key           string             `json:"key,omitempty"`
	ReplaceKey    string             `json:"replaceKey,omitempty"`
	Role          string             `json:"role,omitempty"`
	Sender        string             `json:"sender,omitempty"`
	SendTo        string             `json:"sendTo,omitempty"`
	ChildWorkflow string             `json:"childWorkflow,omitempty"`
	ChildContext  string             `json:"childContext,omitempty"`
	IsExternal    bool               `json:"isExternal,omitempty"`
	Debug         bool               `json:"debug,omitempty"`
	Comment       string             `json:"_comment,omitempty"`
	Location      struct {
		X float64 `json:"x"`
		Y float64 `json:"Y"`
	} `json:"location"`
}

// ImportSteps takes the raw JSON and converts it into an array of Steps
func ImportSteps(raw json.RawMessage) []StepStructure {
	var steps []StepStructure

	err := json.Unmarshal(raw, &steps)
	if err != nil {
		describe(err)
	}
	return steps

}

// Export Converts the StepStructure back to a JSONObject
func (obj *StepStructure) Export() JSONObject {
	str, _ := json.Marshal(obj)
	tmp := JSONObject{}
	tmp.ImportRaw(str)
	return tmp
}

// Merge takes a JSONObject and does a shallow merge with another one
func (obj *JSONObject) Merge(s2 JSONObject) {
	tmp := *obj

	for k, v := range s2 {
		tmp[k] = v
	}

	*obj = tmp
}

// String with the key, returns a string value of it
func (obj *JSONObject) String(key string) string {
	if tmp, ok := (*obj)[key]; ok {
		return fmt.Sprintf("%v", tmp)
	}
	return ""
}

// Copy takes a JSONObject and creates an identical one at a new memory address
func (obj *JSONObject) Copy() JSONObject {
	tmp := obj.Export()
	tmp2 := JSONObject{}
	tmp2.ImportRaw(tmp)
	return tmp2
}

// Export takes the JSONObject and converts it back to raw JSON
func (obj *JSONObject) Export() json.RawMessage {
	str, _ := json.Marshal(obj)
	return str
}

// ImportRaw takes the raw JSON and overwrites the existing structure with it.
func (obj *JSONObject) ImportRaw(str json.RawMessage) {
	err := json.Unmarshal(str, &obj)
	if err != nil {
		describe(err)
	}
}

// ImportString takes the string version of JSON and overwrites the existing structure with it.
func (obj *JSONObject) ImportString(str string) {
	obj.ImportRaw(json.RawMessage(str))
}

// Export takes the JSONObjectArray and converts it back to raw JSON
func (obj *JSONObjectArray) Export() json.RawMessage {
	str, _ := json.Marshal(obj)
	return str
}

// ImportRaw takes the raw JSON and overwrites the existing structure with it.
func (obj *JSONObjectArray) ImportRaw(str json.RawMessage) {
	o := *obj
	tmp := [](map[string]interface{}){}
	o = JSONObjectArray{}
	err := json.Unmarshal(str, &tmp)
	if err != nil {
		describe(err)
	}
	for i := 0; i < len(tmp); i++ {
		o = append(o, JSONObject(tmp[i]))
	}
	*obj = o
}

// ImportString takes the string version of JSON and overwrites the existing structure with it.
func (obj *JSONObjectArray) ImportString(str string) {
	obj.ImportRaw(json.RawMessage(str))
}

func describe(i interface{}) {
	pc, fn, line, _ := runtime.Caller(1)
	fmt.Printf("----  %s[%s:%d]\n---- (%+v, %T)\n\n", runtime.FuncForPC(pc).Name(), fn, line, i, i)
}
