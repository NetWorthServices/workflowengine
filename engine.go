package workflowengine

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"git.aax.dev/agora-altx/models-go/networth"
	"git.aax.dev/agora-altx/workflow-engine/workflowstructs"
	"github.com/tidwall/gjson"
)

// HandleRoute takes the step and sends it to the correct engine
// routes is the definition of functions for both actions and decisions
// steps is the raw JSON structure of all of the workflow steps.
// payload is the current JSON payload of the resulting workflow
// nextStep is the id of the step that is going to be ran. Can be blank (which it will then get the first workflow step)
// callback is a function that gets called after the action has been ran for the step.
func HandleRoute(routes WorkflowDefinitionSet, stepsRaw json.RawMessage, payload json.RawMessage, nextStep string, callback workflowCallback) (response json.RawMessage, err *error) {
	var step StepStructure
	var steps []StepStructure
	var payloadObject JSONObject

	json.Unmarshal(stepsRaw, &steps)
	if nextStep == "" {
		step = steps[0]
	} else {
		for _, s := range steps {
			if s.ID == nextStep {
				step = s
			}
		}
		if step.ID == "" {
			err = errors.New("No new additional step ID")
			return payload, err
		}
	}

	// Take the existing payload item and merge it with the last item in the activity's thread
	if len(act.ThreadJSON) > 0 {
		newestEntry := act.ThreadJSON[len(act.ThreadJSON)-1].Envelope.Copy()
		newestEntry.EmptyVolatileArrays()
		tmp := workflowstructs.Payload{
			ActivityID:       payloadObject.ActivityID,
			ActivityMetaData: newestEntry,
		}
		payloadObject.Merge(tmp)
	}

	msg := networth.ActivityMessage{
		Created:  time.Now(),
		Message:  "",
		Template: step.Template,
	}

	msg.Prepare()
	payloadObject.PathchainID = msg.ID

	//payloadObject.Session = createRequestHeaders(c)

	if routes[step.Route].Name == "" {
		if len(payloadObject.To) < 1 {
			payloadObject.To = []string{payloadObject.From}
		}

		str, _ := json.Marshal(payloadObject)
		blankStep := StepStructure{
			UserInput: true,
			Route:     "STANDARD.BYPASS",
			Payload:   str,
		}
		return &blankStep, act, false
	}

	action := routes[step.Route].Action
	action(step, &payloadObject, routes, c)

	if step.Sender != "" {
		r := gjson.GetBytes(payload, `context.`+step.Sender)
		payloadObject.Sender = r.String()
	} else {
		payloadObject.Sender = payloadObject.From
	}
	if step.SendTo != "" {
		r := gjson.GetBytes(payload, `context.`+step.SendTo)
		payloadObject.To = append(payloadObject.To, r.String())
	}

	if len(payloadObject.To) < 1 {
		payloadObject.To = []string{payloadObject.From}
	}

	if payloadObject.Sender == "" {
		payloadObject.Sender = payloadObject.From
	}

	payloadObject.WorkflowID = step.ID

	msg.Envelope = payloadObject.ToEnvelopeData()
	// A commit

	if payloadObject.Debug {
		str, _ := json.Marshal(msg.Envelope)
		fmt.Println("================ DEBUG  MODE =================")
		fmt.Println("============== CURRENT PAYLOAD ===============")
		fmt.Printf("%v\n", string(str))
		fmt.Printf("\n\nContinue (Y/n) ")

		reader := bufio.NewReader(os.Stdin)
		char, _, _ := reader.ReadRune()
		if string(char) == "N" || string(char) == "n" {
			return nil, act, false
		}
	}
	// Do the Blockchain add right here!
	callback(&msg)

	// Deal with the payload in the activity thread

	payloadObject.InvokedBy = ""
	step.Argument.Payload, _ = json.Marshal(payloadObject)

	if step.UserInput {
		return &step, act, true
	}

	if len(step.Argument.Responses) > 0 {
		newPayload, _ := json.Marshal(payloadObject)
		resp := handleEvaluation(routes, step, newPayload)
		payloadObject.RouteID = resp.Route
		newPayload, _ = json.Marshal(payloadObject)
		nextStep := networth.FindWorkflowStepByID(resp.StepID, "")
		if nextStep.ID == "" {
			return nil, act, true
		}
		return HandleRoute(routes, nextStep, newPayload, callback)
	}

	return nil, act, true

}

// HandleEvaluation takes the step and returns the correct response
func handleEvaluation(routes WorkflowDefinitionSet, step networth.WorkflowStep, payload json.RawMessage) (response networth.WorkflowResponse) {
	var payloadObject Payload

	json.Unmarshal(payload, &payloadObject)

	for _, resp := range step.Argument.Responses {

		action := routes[resp.Route].Decision
		if action(resp, payloadObject, c) {
			return resp
		}

	}

	return networth.WorkflowResponse{}
}
