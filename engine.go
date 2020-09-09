package workflowengine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"git.aax.dev/agora-altx/apiarcade/sockets"
	"git.aax.dev/agora-altx/models-go/networth"
	"git.aax.dev/agora-altx/utils-go/util"
	"git.aax.dev/agora-altx/workflow-engine/workflowstructs"
	"github.com/labstack/echo"
	"github.com/tidwall/gjson"
)

// HandleRoute takes the step and sends it to the correct engine
func HandleRoute(routes WorkflowDefinitionSet, stepRaw, payload json.RawMessage, callback workflowCallback, c echo.Context) (*networth.WorkflowStep, networth.Activity, bool) {
	var step stepStructure
	var payloadObject workflowstructs.Payload

	err := json.Unmarshal(stepRaw, &step)

	// Take the existing payload item and merge it with the last item in the activity's thread
	if len(act.ThreadJSON) > 0 {
		newestEntry := act.ThreadJSON[len(act.ThreadJSON)-1].Envelope.Copy()
		newestEntry.EmptyVolatileArrays()
		tmp := workflowstructs.Payload{
			ActivityID:       payloadObject.ActivityID,
			ActivityMetaData: newestEntry,
		}
		tmp.Waterfall = []util.JSONObject{}
		payloadObject.Merge(tmp)
		//util.Describe(*payloadObject.ActivityMetaData)
		//*payloadObject.ActivityMetaData = *tmp.ActivityMetaData
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
		blankStep := networth.WorkflowStep{
			UserInput: true,
			Route:     "STANDARD.BYPASS",
			Argument: networth.WorkflowArgument{
				Payload: str,
			},
		}
		return &blankStep, act, false
	}
	payloadObject.ExecuteType = networth.ETDefault

	action := routes[step.Route].Action
	action(step, &payloadObject, routes, c)

	if step.Sender != "" {
		if step.Sender == "aax" {
			entities := networth.FindEntityWithStatus(networth.EntityStatusAdmin)
			payloadObject.Sender = entities[0].ID
		} else {
			r := gjson.GetBytes(payload, `context.`+step.Sender)
			payloadObject.Sender = r.String()
		}
	} else {
		payloadObject.Sender = payloadObject.From
	}
	if step.SendTo != "" {
		if step.SendTo == "aax" {
			entities := networth.FindEntityWithStatus(networth.EntityStatusAdmin)
			payloadObject.To = append(payloadObject.To, entities[0].ID)
		} else {
			r := gjson.GetBytes(payload, `context.`+step.SendTo)
			payloadObject.To = append(payloadObject.To, r.String())
		}
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

	act.Find(payloadObject.ActivityID)
	act.AttachMessage(&msg)
	act.Update()
	act.Populate()

	// Deal with the payload in the activity thread

	sockets.CreateActivitySocketMessage(act, "")

	payloadObject.InvokedBy = ""
	step.Argument.Payload, _ = json.Marshal(payloadObject)

	if step.UserInput {
		return &step, act, true
	}

	if len(step.Argument.Responses) > 0 {
		newPayload, _ := json.Marshal(payloadObject)
		resp := handleEvaluation(routes, step, newPayload, c)
		payloadObject.RouteID = resp.Route
		newPayload, _ = json.Marshal(payloadObject)
		nextStep := networth.FindWorkflowStepByID(resp.StepID, "")
		if nextStep.ID == "" {
			return nil, act, true
		}
		return HandleRoute(routes, nextStep, newPayload, callback, c)
	}

	return nil, act, true

}

// HandleEvaluation takes the step and returns the correct response
func handleEvaluation(routes workflowstructs.WorkflowDefinitionSet, step networth.WorkflowStep, payload json.RawMessage, c echo.Context) (response networth.WorkflowResponse) {
	var payloadObject workflowstructs.Payload

	err := json.Unmarshal(payload, &payloadObject)
	if err != nil {
		util.Describe(err)
	}

	for _, resp := range step.Argument.Responses {

		action := routes[resp.Route].Decision
		if action(resp, payloadObject, c) {
			return resp
		}

	}

	return networth.WorkflowResponse{}
}

func createRequestHeaders(c echo.Context) (resp util.JSONObject) {
	resp["Remote-Host"] = c.Request().RemoteAddr
	resp["User-Agent"] = c.Request().UserAgent()
	return
}
