package operator

import (
	"fmt"
	"sort"
	"strings"
)

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("operator %s not found", e.Name)
}

type Builder func(operationID string, arguments Arguments) Operator

// map[operatorName]map[operatorVersion]OperatorBuilder
type BuildersTree map[string]map[string]Builder

func extractOperatorNameAndVersion(operatorName string) (string, string, error) {
	parts := strings.Split(operatorName, "@")
	if len(parts) == 1 {
		// no version found, just operator name
		return parts[0], "", nil
	}
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"could not extract the operator version from %s, version should follow <operatorName>@<version> syntax",
			operatorName,
		)
	}
	return parts[0], parts[1], nil
}

type Registry struct {
	operators BuildersTree
}

func NewRegistry(operators BuildersTree) *Registry {
	return &Registry{
		operators: operators,
	}
}

func (m *Registry) GetOperatorBuilder(name string) (Builder, error) {
	operatorName, version, err := extractOperatorNameAndVersion(name)
	if err != nil {
		return nil, err
	}
	if version == "" {
		latestVersion, err := m.getLatestVersionForOperator(name)
		if err != nil {
			return nil, err
		}
		version = latestVersion
	}

	if g, found := m.operators[operatorName][version]; found {
		return g, nil
	}
	return nil, &NotFoundError{Name: name}
}

func (m *Registry) AvailableOperators() []string {
	operatorList := []string{}

	for operatorName, versions := range m.operators {
		operatorVersions := []string{}
		for v := range versions {
			operatorVersions = append(operatorVersions, v)
		}
		sort.Strings(operatorVersions)
		operatorList = append(
			operatorList,
			fmt.Sprintf("%s - %s", operatorName, strings.Join(operatorVersions, "/")),
		)
	}

	return operatorList
}

func (m *Registry) getLatestVersionForOperator(name string) (string, error) {
	availableOperators, found := m.operators[name]
	if !found {
		return "", &NotFoundError{Name: name}
	}
	versions := []string{}
	for v := range availableOperators {
		versions = append(versions, v)
	}

	sort.Strings(versions)

	return versions[len(versions)-1], nil
}

func StandardRegistry(options ...BaseOperatorOption) *Registry {
	return &Registry{
		operators: BuildersTree{
			ClusterMaintenanceChangeOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewClusterMaintenanceChange(arguments, operationID, Options[ClusterMaintenanceChange]{
						BaseOperatorOptions: options,
					})
				},
			},
			ClusterResourceRefreshOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewClusterResourceRefresh(arguments, operationID, Options[ClusterResourceRefresh]{
						BaseOperatorOptions: options,
					})
				},
			},
			CrmClusterStartOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewCrmClusterStart(arguments, operationID, Options[CrmClusterStart]{
						BaseOperatorOptions: options,
					})
				},
			},
			CrmClusterStopOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewCrmClusterStop(arguments, operationID, Options[CrmClusterStop]{
						BaseOperatorOptions: options,
					})
				},
			},
			HostRebootOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewHostReboot(arguments, operationID, Options[HostReboot]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapInstanceStartOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSAPInstanceStart(arguments, operationID, Options[SAPInstanceStart]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapInstanceStopOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSAPInstanceStop(arguments, operationID, Options[SAPInstanceStop]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapSystemStartOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSAPSystemStart(arguments, operationID, Options[SAPSystemStart]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapSystemStopOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSAPSystemStop(arguments, operationID, Options[SAPSystemStop]{
						BaseOperatorOptions: options,
					})
				},
			},
			SaptuneApplySolutionOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSaptuneApplySolution(arguments, operationID, Options[SaptuneApplySolution]{
						BaseOperatorOptions: options,
					})
				},
			},
			SaptuneChangeSolutionOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewSaptuneChangeSolution(arguments, operationID, Options[SaptuneChangeSolution]{
						BaseOperatorOptions: options,
					})
				},
			},
			PacemakerEnableOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewServiceEnable(PacemakerEnableOperatorName, arguments, operationID, Options[ServiceEnable]{
						BaseOperatorOptions: options,
						OperatorOptions: []Option[ServiceEnable]{
							Option[ServiceEnable](WithServiceToEnable(pacemakerServiceName)),
						},
					})
				},
			},
			PacemakerDisableOperatorName: map[string]Builder{
				"v1": func(operationID string, arguments Arguments) Operator {
					return NewServiceDisable(PacemakerDisableOperatorName, arguments, operationID, Options[ServiceDisable]{
						BaseOperatorOptions: options,
						OperatorOptions: []Option[ServiceDisable]{
							Option[ServiceDisable](WithServiceToDisable(pacemakerServiceName)),
						},
					})
				},
			},
		},
	}
}
