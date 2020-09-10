package workflowengine

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// HandleRoute takes the step and sends it to the correct engine
// routes is the definition of functions for both actions and decisions
// steps is the raw JSON structure of all of the workflow steps.
// paypayloadJSONload is the current JSON payload of the resulting workflow
// nextStep is the id of the step that is going to be ran. Can be blank (which it will then get the first workflow step)
// callback is a function that gets called after the action has been ran for the step.
func HandleRoute(routes WorkflowDefinitionSet, stepsRaw json.RawMessage, payloadJSON json.RawMessage, nextStep string, callback WorkflowCallback) (response json.RawMessage, err *error) {
	var step StepStructure
	var payload JSONObjectArray

	payload.ImportRaw(payloadJSON)
	steps := ImportSteps(stepsRaw)

	if nextStep == "" {
		step = steps[0]
	} else {
		for _, s := range steps {
			if s.ID == nextStep {
				step = s
			}
		}
		if step.ID == "" {
			err := errors.New("No new additional step ID")
			return payload.Export(), &err
		}
	}

	// Take the existing payload item and merge it with the last item in the activity's thread
	if len(payload) > 0 {
		newestEntry := payload[len(payload)-1].Copy()
		payload = append(payload, newestEntry)
	} else {
		payload = JSONObjectArray{JSONObject{}}
	}

	msg := payload[len(payload)-1]

	msg["id"] = uuid.New()
	if step.Debug {
		msg["debug"] = true
	}

	action := routes[step.Route].Action
	action(step.Export(), &msg)

	if step.Sender != "" {
		r := gjson.GetBytes(msg.Export(), `context.`+step.Sender)
		msg["sender"] = r.String()
	} else {
		msg["sender"] = msg["from"]
	}
	if step.SendTo != "" {
		r := gjson.GetBytes(msg.Export(), `context.`+step.SendTo)
		msg["to"] = append(msg["to"].([]string), r.String())
	}

	if len(msg["to"].([]string)) < 1 {
		msg["to"] = []string{msg["from"].(string)}
	}

	if msg["sender"] == "" {
		msg["sender"] = msg["from"]
	}

	msg["workflowID"] = step.ID

	// A commit

	if msg["debug"].(bool) {
		str := msg.Export()
		fmt.Println("================= DEBUG  MODE ==================")
		fmt.Println("=============== CURRENT PAYLOAD ================")
		fmt.Printf("%v\n", string(str))
		fmt.Printf("\n\nContinue (Y/n) ")

		reader := bufio.NewReader(os.Stdin)
		char, _, _ := reader.ReadRune()
		if string(char) == "N" || string(char) == "n" {
			payload[len(payload)-1] = msg
			err := errors.New("Debug was told to stop")
			return payload.Export(), &err
		}
	}
	// Do the Blockchain add right here!
	callback(&msg)

	// Deal with the payload in the activity thread

	msg["invokedBy"] = ""
	payload[len(payload)-1] = msg

	if step.UserInput {
		return payload.Export(), nil
	}

	if len(step.Responses) > 0 {
		resp := handleEvaluation(routes, step, msg.Export())
		msg["routeID"] = resp.Route

		payload[len(payload)-1] = msg
		nextStep := findWorkflowStepByID(stepsRaw, resp.StepID)
		if nextStep.ID == "" {
			err := errors.New("The next step given does not exist in the list of steps")
			return payload.Export(), &err
		}
		return HandleRoute(routes, stepsRaw, payload.Export(), nextStep.ID, callback)
	}

	return payload.Export(), nil

}

// HandleEvaluation takes the step and returns the correct response
func handleEvaluation(routes WorkflowDefinitionSet, step StepStructure, payload json.RawMessage) (response workflowResponse) {
	var payloadObject JSONObject

	payloadObject.ImportRaw(payload)

	for _, resp := range step.Responses {

		decision := routes[resp.Route].Decision
		if decision(payloadObject) {
			return resp
		}

	}

	return workflowResponse{}
}

func findWorkflowStepByID(stepsRaw json.RawMessage, id string) StepStructure {
	steps := ImportSteps(stepsRaw)
	for _, s := range steps {
		if s.ID == id {
			return s
		}
	}
	return StepStructure{}
}
