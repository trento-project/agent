package factsengine

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
)

const (
	FactsGatheringRequested     = "Trento.Checks.V1.FactsGatheringRequested"
	OperationExecutionRequested = "Trento.Operations.V1.OperationExecutionRequested"
)

func (c *FactsEngine) handleEvent(_ string, request []byte) error {
	eventType, err := events.EventType(request)
	if err != nil {
		return errors.Wrap(err, "Error getting event type")
	}
	switch eventType {
	case FactsGatheringRequested:
		err := c.handleFactsGatheringRequestedEvent(request)
		if err != nil {
			return errors.Wrap(err, "Error handling facts request")
		}
	case OperationExecutionRequested:
		err := c.handleOperationRequestedEvent(request)
		if err != nil {
			return errors.Wrap(err, "Error handling operation request")
		}
	default:
		return fmt.Errorf("Invalid event type: %s", eventType)
	}
	return nil
}

func (c *FactsEngine) handleFactsGatheringRequestedEvent(factsRequestByte []byte) error {
	factsRequest, err := FactsGatheringRequestedFromEvent(factsRequestByte)
	if err != nil {
		return err
	}

	agentFactsRequest := getAgentFacts(c.agentID, factsRequest)

	if agentFactsRequest == nil {
		log.Infof("FactsGatheringRequested is not for this agent. Discarding facts gathering execution")
		return nil
	}

	gatheredFacts, err := gatherFacts(
		factsRequest.ExecutionID,
		c.agentID,
		factsRequest.GroupID,
		agentFactsRequest,
		c.gathererRegistry,
	)
	if err != nil {
		log.Errorf("Error gathering facts: %s", err)
		return err
	}

	if err := c.publishFacts(gatheredFacts); err != nil {
		log.Errorf("Error publishing facts: %s", err)
		return err
	}

	return nil
}

func (c *FactsEngine) handleOperationRequestedEvent(operationRequestByte []byte) error {
	operationRequest, err := OperationRequestedFromEvent(operationRequestByte)
	if err != nil {
		return err
	}

	log.Infof("Received operation request: %v", operationRequest)

	agentOperationRequest := getAgentOperations(c.agentID, operationRequest)

	if agentOperationRequest == nil {
		log.Infof("OperationExecutionRequested is not for this agent. Discarding operation execution")
		return nil
	}

	dummyEvent, _ := OperationResultToEvent(entities.OperationCompleted{
		OperationID: operationRequest.OperationID,
		GroupID:     operationRequest.GroupID,
		StepNumber:  operationRequest.StepNumber,
		AgentID:     c.agentID,
		Phase:       "COMMIT",
		Diff:        make(map[string]string),
	})

	if err := c.factsServiceAdapter.Publish(
		exchange, executionsRoutingKey, events.ContentType(), dummyEvent); err != nil {

		log.Error(err)
		return err
	}

	return nil
}

func getAgentFacts(
	agentID string,
	factsRequest *entities.FactsGatheringRequested) *entities.FactsGatheringRequestedTarget {

	for _, agentRequests := range factsRequest.Targets {
		if agentRequests.AgentID == agentID {
			return &agentRequests
		}
	}

	return nil
}

func getAgentOperations(
	agentID string,
	operationRequest *entities.OperationRequested) *entities.OperationRequestedTarget {

	for _, agentRequests := range operationRequest.Targets {
		if agentRequests.AgentID == agentID {
			return &agentRequests
		}
	}

	return nil
}

func (c *FactsEngine) publishFacts(facts entities.FactsGathered) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	event, err := FactsGatheredToEvent(facts)
	if err != nil {
		return err
	}

	if err := c.factsServiceAdapter.Publish(
		exchange, executionsRoutingKey, events.ContentType(), event); err != nil {

		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
