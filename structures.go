package workflowengine

import "encoding/json"

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

type stepStructure struct {
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
