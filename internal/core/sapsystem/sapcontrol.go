package sapsystem

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	sapcontrol "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
)

type SAPControl struct {
	Processes  []*sapcontrol.OSProcess
	Instances  []*sapcontrol.SAPInstance
	Properties []*sapcontrol.InstanceProperty
}

func NewSAPControl(ctx context.Context, w sapcontrol.WebService) (*SAPControl, error) {
	properties, err := w.GetInstanceProperties(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "SAPControl web service error")
	}

	processes, err := w.GetProcessList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "SAPControl web service error")
	}

	instances, err := w.GetSystemInstanceList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "SAPControl web service error")
	}

	return &SAPControl{
		Properties: properties.Properties,
		Processes:  processes.Processes,
		Instances:  instances.Instances,
	}, nil
}

func (s *SAPControl) findProperty(key string) (string, error) {
	for _, item := range s.Properties {
		if item.Property == key {
			return item.Value, nil
		}
	}

	return "", fmt.Errorf("Property %s not found", key)
}
