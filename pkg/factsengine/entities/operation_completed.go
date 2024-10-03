package entities

type OperationCompleted struct {
	OperationID string
	GroupID     string
	StepNumber  int32
	AgentID     string
	Phase       string
	Diff        map[string]string
}
