package entities

type OperationRequestedTarget struct {
	AgentID   string
	Name      string
	Operator  string
	Arguments map[string]interface{}
}

type OperationRequested struct {
	OperationID string
	GroupID     string
	StepNumber  int32
	Targets     []OperationRequestedTarget
}
