package workflowengine

import (
	"encoding/json"
	"fmt"
)

type JSONObject map[string]interface{}
type JSONObjectArray []JSONObject

type ActionFunction func(JSONObject, *JSONObject, WorkflowDefinitionSet)
type DecisionFunction func(JSONObject, JSONObject) bool

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
	Comment       string             `json:"_comment,omitempty"`
	Location      struct {
		X float64 `json:"x"`
		Y float64 `json:"Y"`
	} `json:"location"`
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

func (obj *JSONObject) ImportRaw(str json.RawMessage) {
	err := json.Unmarshal(str, &obj)
	if err != nil {
		Describe(err)
	}
}

func (obj *JSONObject) ImportString(str string) {
	obj.ImportRaw(json.RawMessage(str))
}

func (obj *JSONObjectArray) ImportRaw(str json.RawMessage) {
	o := *obj
	tmp := [](map[string]interface{}){}
	o = JSONObjectArray{}
	err := json.Unmarshal(str, &tmp)
	if err != nil {
		Describe(err)
	}
	for i := 0; i < len(tmp); i++ {
		o = append(o, JSONObject(tmp[i]))
	}
	*obj = o
}

func (obj *JSONObjectArray) ImportString(str string) {
	obj.ImportRaw(json.RawMessage(str))
}
