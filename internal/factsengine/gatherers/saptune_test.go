//nolint:dupl
package gatherers_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SaptuneTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestSaptuneTestSuite(t *testing.T) {
	suite.Run(t, new(SaptuneTestSuite))
}

func (suite *SaptuneTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererStatus() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/saptune-status.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "status", "--non-compliance-check").Return(mockOutput, nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_status",
			Gatherer: "saptune",
			Argument: "status",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_status",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"$schema":      &entities.FactValueString{Value: "file:///usr/share/saptune/schemas/1.0/saptune_status.schema.json"},
					"publish_time": &entities.FactValueString{Value: "2023-09-15 15:15:14.599"},
					"argv":         &entities.FactValueString{Value: "saptune --format json status"},
					"pid":          &entities.FactValueInt{Value: 6593},
					"command":      &entities.FactValueString{Value: "status"},
					"exit_code":    &entities.FactValueInt{Value: 1},
					"result": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"services": &entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"saptune": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueString{Value: "disabled"},
											&entities.FactValueString{Value: "inactive"},
										},
									},
									"sapconf": &entities.FactValueList{Value: []entities.FactValue{}},
									"tuned":   &entities.FactValueList{Value: []entities.FactValue{}},
								},
							},
							"systemd_system_state":      &entities.FactValueString{Value: "degraded"},
							"tuning_state":              &entities.FactValueString{Value: "compliant"},
							"virtualization":            &entities.FactValueString{Value: "kvm"},
							"configured_version":        &entities.FactValueString{Value: "3"},
							"package_version":           &entities.FactValueString{Value: "3.1.0"},
							"solution_enabled":          &entities.FactValueList{Value: []entities.FactValue{}},
							"notes_enabled_by_solution": &entities.FactValueList{Value: []entities.FactValue{}},
							"solution_applied":          &entities.FactValueList{Value: []entities.FactValue{}},
							"notes_applied_by_solution": &entities.FactValueList{Value: []entities.FactValue{}},
							"notes_enabled_additionally": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1410736"},
								},
							},
							"notes_enabled": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1410736"},
								},
							},
							"notes_applied": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1410736"},
								},
							},
							"staging": &entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"staging_enabled":  &entities.FactValueBool{Value: false},
									"notes_staged":     &entities.FactValueList{Value: []entities.FactValue{}},
									"solutions_staged": &entities.FactValueList{Value: []entities.FactValue{}},
								},
							},
							"remember_message": &entities.FactValueString{Value: "This is a reminder"},
						},
					},
					"messages": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "actions.go:85: ATTENTION: You are running a test version"},
								},
							},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererNoteVerify() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/saptune-note-verify.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "note", "verify").Return(mockOutput, nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_note_verify",
			Gatherer: "saptune",
			Argument: "note-verify",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_note_verify",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"$schema": &entities.FactValueString{
						Value: "file:///usr/share/saptune/schemas/1.0/saptune_note_verify.schema.json",
					},
					"publish_time": &entities.FactValueString{
						Value: "2023-04-24 15:49:43.399",
					},
					"argv": &entities.FactValueString{
						Value: "saptune --format json note verify",
					},
					"pid": &entities.FactValueInt{
						Value: 25202,
					},
					"command": &entities.FactValueString{
						Value: "note verify",
					},
					"exit_code": &entities.FactValueInt{
						Value: 1,
					},
					"result": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"verifications": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":        &entities.FactValueString{Value: "1771258"},
											"note_version":   &entities.FactValueString{Value: "6"},
											"parameter":      &entities.FactValueString{Value: "LIMIT_@dba_hard_nofile"},
											"compliant":      &entities.FactValueBool{Value: true},
											"expected_value": &entities.FactValueString{Value: "@dba hard nofile 1048576"},
											"actual_value":   &entities.FactValueString{Value: "@dba hard nofile 1048576"},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":        &entities.FactValueString{Value: "1771258"},
											"note_version":   &entities.FactValueString{Value: "6"},
											"parameter":      &entities.FactValueString{Value: "LIMIT_@dba_soft_nofile"},
											"compliant":      &entities.FactValueBool{Value: true},
											"expected_value": &entities.FactValueString{Value: "@dba soft nofile 1048576"},
											"actual_value":   &entities.FactValueString{Value: "@dba soft nofile 1048576"},
										},
									},
								},
							},
							"attentions": &entities.FactValueList{
								Value: []entities.FactValue{},
							},
							"notes_enabled": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1771258"},
								},
							},
							"system_compliance": &entities.FactValueBool{Value: false},
						},
					},
					"messages": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "actions.go:85 You are running a test version"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "WARNING"},
									"message":  &entities.FactValueString{Value: "sysctl.go:73: Parameter 'kernel.shmmax' redefined "},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "WARNING"},
									"message":  &entities.FactValueString{Value: "sysctl.go:73: Parameter 'kernel.shmall' redefined"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "ini.go:308: block device related section settings detected"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "ERROR"},
									"message":  &entities.FactValueString{Value: "system.go:148: The parameters have deviated from recommendations"},
								},
							},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererSolutionVerify() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/saptune-solution-verify.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "solution", "verify").Return(mockOutput, nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_solution_verify",
			Gatherer: "saptune",
			Argument: "solution-verify",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_solution_verify",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"$schema":      &entities.FactValueString{Value: "file:///usr/share/saptune/schemas/1.0/saptune_solution_verify.schema.json"},
					"publish_time": &entities.FactValueString{Value: "2023-04-27 17:17:23.743"},
					"argv":         &entities.FactValueString{Value: "saptune --format json solution verify"},
					"pid":          &entities.FactValueInt{Value: 2538},
					"command":      &entities.FactValueString{Value: "solution verify"},
					"exit_code":    &entities.FactValueInt{Value: 1},
					"result": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"verifications": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":        &entities.FactValueString{Value: "1771258"},
											"note_version":   &entities.FactValueString{Value: "6"},
											"parameter":      &entities.FactValueString{Value: "LIMIT_@dba_hard_nofile"},
											"compliant":      &entities.FactValueBool{Value: true},
											"expected_value": &entities.FactValueString{Value: "@dba hard nofile 1048576"},
											"actual_value":   &entities.FactValueString{Value: "@dba hard nofile 1048576"},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":        &entities.FactValueString{Value: "1771258"},
											"note_version":   &entities.FactValueString{Value: "6"},
											"parameter":      &entities.FactValueString{Value: "LIMIT_@dba_soft_nofile"},
											"compliant":      &entities.FactValueBool{Value: true},
											"expected_value": &entities.FactValueString{Value: "@dba soft nofile 1048576"},
											"actual_value":   &entities.FactValueString{Value: "@dba soft nofile 1048576"},
										},
									},
								},
							},
							"attentions": &entities.FactValueList{
								Value: []entities.FactValue{},
							},
							"notes_enabled": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1771258"},
								},
							},
							"system_compliance": &entities.FactValueBool{Value: false},
						},
					},
					"messages": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "actions.go:85 You are running a test version"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "WARNING"},
									"message":  &entities.FactValueString{Value: "sysctl.go:73: Parameter 'kernel.shmmax' redefined "},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "WARNING"},
									"message":  &entities.FactValueString{Value: "sysctl.go:73: Parameter 'kernel.shmall' redefined"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "ini.go:308: block device related section settings detected"},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "ERROR"},
									"message":  &entities.FactValueString{Value: "system.go:148: The parameters have deviated from recommendations"},
								},
							},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererSolutionList() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/saptune-solution-list.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "solution", "list").Return(mockOutput, nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_solution_list",
			Gatherer: "saptune",
			Argument: "solution-list",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_solution_list",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"$schema":      &entities.FactValueString{Value: "file:///usr/share/saptune/schemas/1.0/saptune_solution_list.schema.json"},
					"publish_time": &entities.FactValueString{Value: "2023-04-27 17:21:27.926"},
					"argv":         &entities.FactValueString{Value: "saptune --format json solution list"},
					"pid":          &entities.FactValueInt{Value: 2582},
					"command":      &entities.FactValueString{Value: "solution list"},
					"exit_code":    &entities.FactValueInt{Value: 0},
					"result": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"solutions_available": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"solution_id": &entities.FactValueString{Value: "BOBJ"},
											"note_list": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueString{Value: "1771258"},
												},
											},
											"solution_enabled":         &entities.FactValueBool{Value: false},
											"solution_override_exists": &entities.FactValueBool{Value: false},
											"custom_solution":          &entities.FactValueBool{Value: false},
											"solution_deprecated":      &entities.FactValueBool{Value: false},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"solution_id": &entities.FactValueString{Value: "DEMO"},
											"note_list": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueString{Value: "demo"},
												},
											},
											"solution_enabled":         &entities.FactValueBool{Value: false},
											"solution_override_exists": &entities.FactValueBool{Value: false},
											"custom_solution":          &entities.FactValueBool{Value: true},
											"solution_deprecated":      &entities.FactValueBool{Value: false},
										},
									},
								},
							},
							"remember_message": &entities.FactValueString{Value: ""},
						},
					},
					"messages": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "actions.go:85 You are running a test version"},
								},
							},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererNoteList() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/saptune-note-list.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "note", "list").Return(mockOutput, nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_note_list",
			Gatherer: "saptune",
			Argument: "note-list",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_note_list",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"$schema":      &entities.FactValueString{Value: "file:///usr/share/saptune/schemas/1.0/saptune_note_list.schema.json"},
					"publish_time": &entities.FactValueString{Value: "2023-04-27 17:28:53.073"},
					"argv":         &entities.FactValueString{Value: "saptune --format json note list"},
					"pid":          &entities.FactValueInt{Value: 2604},
					"command":      &entities.FactValueString{Value: "note list"},
					"exit_code":    &entities.FactValueInt{Value: 0},
					"result": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"notes_available": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":          &entities.FactValueString{Value: "1410736"},
											"note_description": &entities.FactValueString{Value: "TCP/IP: setting keepalive interval"},
											"note_reference": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueString{Value: "https://launchpad.support.sap.com/#/notes/1410736"},
												},
											},
											"note_version":             &entities.FactValueString{Value: "6"},
											"note_release_date":        &entities.FactValueString{Value: "13.01.2020"},
											"note_enabled_manually":    &entities.FactValueBool{Value: false},
											"note_enabled_by_solution": &entities.FactValueBool{Value: false},
											"note_reverted_manually":   &entities.FactValueBool{Value: false},
											"note_override_exists":     &entities.FactValueBool{Value: false},
											"custom_note":              &entities.FactValueBool{Value: false},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"note_id":          &entities.FactValueString{Value: "1656250"},
											"note_description": &entities.FactValueString{Value: "SAP on AWS: prerequisites - only Linux"},
											"note_reference": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueString{Value: "https://launchpad.support.sap.com/#/notes/1656250"},
												},
											},
											"note_version":             &entities.FactValueString{Value: "46"},
											"note_release_date":        &entities.FactValueString{Value: "11.05.2022"},
											"note_enabled_manually":    &entities.FactValueBool{Value: false},
											"note_enabled_by_solution": &entities.FactValueBool{Value: true},
											"note_reverted_manually":   &entities.FactValueBool{Value: false},
											"note_override_exists":     &entities.FactValueBool{Value: false},
											"custom_note":              &entities.FactValueBool{Value: false},
										},
									},
								},
							},
							"notes_enabled": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "1656250"},
								},
							},
							"remember_message": &entities.FactValueString{Value: ""},
						},
					},
					"messages": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"priority": &entities.FactValueString{Value: "NOTICE"},
									"message":  &entities.FactValueString{Value: "actions.go:85 You are running a test version"},
								},
							},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererNoArgumentProvided() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "saptune",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "saptune",
			Argument: "",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "no_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "saptune-missing-argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "saptune-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererUnsupportedArgument() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "unsupported_argument_fact",
			Gatherer: "saptune",
			Argument: "unsupported",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "unsupported_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "the requested argument is not currently supported: unsupported",
				Type:    "saptune-unsupported-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererVersionUnsupported() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("2.0.0"), nil,
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_status",
			Gatherer: "saptune",
			Argument: "status",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{}

	suite.EqualError(err, "fact gathering error: saptune-version-not-supported - currently installed version of saptune is not supported")
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererNotInstalled() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		nil, errors.New("exit status 1"),
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_status",
			Gatherer: "saptune",
			Argument: "status",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{}

	suite.EqualError(err, "fact gathering error: saptune-not-installed - saptune is not installed: could not determine saptune version: exit status 1")
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererCommandError() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "status", "--non-compliance-check").Return(
		nil, errors.New("exit status 1"),
	)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_status",
			Gatherer: "saptune",
			Argument: "status",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "saptune_status",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error executing saptune command: unexpected end of JSON input",
				Type:    "saptune-cmd-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SaptuneTestSuite) TestSaptuneGathererCommandCaching() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)
	suite.mockExecutor.On("Exec", "saptune", "--format", "json", "status", "--non-compliance-check").Return([]byte("{\"some_json_key\": \"some_value\"}"), nil)
	c := gatherers.NewSaptuneGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "saptune_repeated_argument_1",
			Gatherer: "saptune",
			Argument: "status",
		},
		{
			Name:     "saptune_repeated_argument_2",
			Gatherer: "saptune",
			Argument: "status",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "saptune_repeated_argument_1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"some_json_key": &entities.FactValueString{Value: "some_value"},
				},
			},
		},
		{
			Name: "saptune_repeated_argument_2",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"some_json_key": &entities.FactValueString{Value: "some_value"},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
	suite.mockExecutor.AssertNumberOfCalls(suite.T(), "Exec", 2) // 1 for rpm, 1 for saptune
}
