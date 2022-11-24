package entities

type FactsGatheringRequestedTarget struct {
	AgentID      string
	FactRequests []FactRequest
}

type FactRequest struct {
	Argument string
	CheckID  string
	Gatherer string
	Name     string
}

type FactsGatheringRequested struct {
	ExecutionID string
	GroupID     string
	Targets     []FactsGatheringRequestedTarget
}

type GroupedByGathererRequestedTarget struct {
	FactRequests map[string][]FactRequest
}
