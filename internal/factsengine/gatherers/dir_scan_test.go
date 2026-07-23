// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package gatherers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	"github.com/trento-project/agent/v3/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/v3/pkg/factsengine/entities"
)

const dirScanTestBasePath = "/var/test"

type DirScanGathererSuite struct {
	suite.Suite

	testFS     afero.Fs
	basePathFS string
}

func TestDirScanGathererSuite(t *testing.T) {
	suite.Run(t, new(DirScanGathererSuite))
}

func (s *DirScanGathererSuite) SetupSuite() {
	bfs := afero.NewOsFs()

	s.basePathFS = afero.GetTempDir(bfs, "")
	tFs := afero.NewBasePathFs(bfs, s.basePathFS)

	for i := range 3 {
		dirPath := fmt.Sprintf("%s/%d/", dirScanTestBasePath, i)
		filePath := fmt.Sprintf("%s/%d/ASCS%d", dirScanTestBasePath, i, i)
		_ = tFs.MkdirAll(dirPath, 0777)
		_, _ = tFs.Create(filePath)
	}

	_, _ = tFs.Create(dirScanTestBasePath + "/1/ASCS3")
	_, _ = tFs.Create(dirScanTestBasePath + "/1/ASDX2")
	_, _ = tFs.Create(dirScanTestBasePath + "/2/ASDX1")

	s.testFS = tFs
}

func (s *DirScanGathererSuite) TearDownSuite() {
	err := s.testFS.RemoveAll(dirScanTestBasePath)
	s.Require().NoError(err)
}

func (s *DirScanGathererSuite) TestDirScanningErrorDirScanningWithoutGlob() {
	groupSearcher := mocks.NewMockGroupSearcher(s.T())
	groupSearcher.On("GetGroupByID", mock.AnythingOfType("string")).Return("trento", nil)

	userSearcher := mocks.NewMockUserSearcher(s.T())
	userSearcher.On("GetUsernameByID", mock.AnythingOfType("string")).Return("trento", nil)

	g := gatherers.NewDirScanGatherer(s.testFS, userSearcher, groupSearcher)

	fr := []entities.FactRequest{{
		Argument: dirScanTestBasePath + "/1/ASCS3",
		CheckID:  "check1",
		Gatherer: "dir_scan",
		Name:     "dir_scan",
	}}

	result, _ := g.Gather(context.Background(), fr)
	expectedResult := []entities.Fact{{
		Name:    "dir_scan",
		CheckID: "check1",
		Value: &entities.FactValueList{
			Value: []entities.FactValue{
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueString{Value: "trento"},
						"group": &entities.FactValueString{Value: "trento"},
						"name":  &entities.FactValueString{Value: "/var/test/1/ASCS3"},
					},
				},
			},
		},
	}}

	s.Equal(expectedResult, result)
}

func (s *DirScanGathererSuite) TestDirScanningErrorNoArgument() {
	g := gatherers.NewDirScanGatherer(s.testFS, &gatherers.CredentialsFetcher{}, &gatherers.CredentialsFetcher{})

	fr := []entities.FactRequest{{
		CheckID:  "check1",
		Gatherer: "dir_scan",
		Name:     "dir_scan",
	}}

	expectedResult := []entities.Fact{{
		Name:    "dir_scan",
		CheckID: "check1",
		Value:   nil,
		Error: &entities.FactGatheringError{
			Type:    "dir-scan-missing-argument",
			Message: "missing required argument",
		},
	}}

	result, _ := g.Gather(context.Background(), fr)
	s.Equal(expectedResult, result)
}

func (s *DirScanGathererSuite) TestDirScanningSuccess() {
	dirScanTestGlobPattern := "/var/test/*/ASCS*"

	groupSearcher := mocks.NewMockGroupSearcher(s.T())
	groupSearcher.On("GetGroupByID", mock.AnythingOfType("string")).Return("trento", nil)

	userSearcher := mocks.NewMockUserSearcher(s.T())
	userSearcher.On("GetUsernameByID", mock.AnythingOfType("string")).Return("trento", nil)

	g := gatherers.NewDirScanGatherer(s.testFS, userSearcher, groupSearcher)
	fr := []entities.FactRequest{{
		Argument: dirScanTestGlobPattern,
		CheckID:  "check1",
		Gatherer: "dir_scan",
		Name:     "dir_scan",
	}}

	expectedResult := []entities.Fact{{
		Name:    "dir_scan",
		CheckID: "check1",
		Value: &entities.FactValueList{
			Value: []entities.FactValue{
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueString{Value: "trento"},
						"group": &entities.FactValueString{Value: "trento"},
						"name":  &entities.FactValueString{Value: "/var/test/0/ASCS0"},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueString{Value: "trento"},
						"group": &entities.FactValueString{Value: "trento"},
						"name":  &entities.FactValueString{Value: "/var/test/1/ASCS1"},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueString{Value: "trento"},
						"group": &entities.FactValueString{Value: "trento"},
						"name":  &entities.FactValueString{Value: "/var/test/1/ASCS3"},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"owner": &entities.FactValueString{Value: "trento"},
						"group": &entities.FactValueString{Value: "trento"},
						"name":  &entities.FactValueString{Value: "/var/test/2/ASCS2"},
					},
				},
			},
		},
	}}

	result, err := g.Gather(context.Background(), fr)
	s.Require().NoError(err)
	s.Equal(expectedResult, result)
}

func (s *DirScanGathererSuite) TestDirScanningGathererContextCancelled() {
	dirScanTestGlobPattern := "/var/test/*/ASCS*"

	groupSearcher := mocks.NewMockGroupSearcher(s.T())
	groupSearcher.On("GetGroupByID", mock.AnythingOfType("string")).Return("trento", nil).Maybe()

	userSearcher := mocks.NewMockUserSearcher(s.T())
	userSearcher.On("GetUsernameByID", mock.AnythingOfType("string")).Return("trento", nil).Maybe()

	c := gatherers.NewDirScanGatherer(s.testFS, userSearcher, groupSearcher)
	factRequests := []entities.FactRequest{{
		Argument: dirScanTestGlobPattern,
		CheckID:  "check1",
		Gatherer: "dir_scan",
		Name:     "dir_scan",
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factResults, err := c.Gather(ctx, factRequests)

	s.Require().Error(err)
	s.Empty(factResults)
}

// When the context is already cancelled, Gather returns as soon as it observes ctx.Done(),
// while the background goroutine keeps scanning and then tries to send its result on `results`.
// If that channel isn't buffered, nobody is left to receive and the goroutine leaks forever.
func (s *DirScanGathererSuite) TestDirScanningGathererContextCancelledDoesNotLeakGoroutine() {
	dirScanTestGlobPattern := "/var/test/*/ASCS*"

	groupSearcher := mocks.NewMockGroupSearcher(s.T())
	groupSearcher.On("GetGroupByID", mock.AnythingOfType("string")).Return("trento", nil).Maybe()

	userSearcher := mocks.NewMockUserSearcher(s.T())
	userSearcher.On("GetUsernameByID", mock.AnythingOfType("string")).Return("trento", nil).Maybe()

	c := gatherers.NewDirScanGatherer(s.testFS, userSearcher, groupSearcher)
	factRequests := []entities.FactRequest{{
		Argument: dirScanTestGlobPattern,
		CheckID:  "check1",
		Gatherer: "dir_scan",
		Name:     "dir_scan",
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _ = c.Gather(ctx, factRequests)

	// goleak.Find already retries internally for a short window; polling it from another
	// goroutine (e.g. via require.Eventually) would make that polling goroutine itself show
	// up as "unexpected", so check it directly in this goroutine instead.
	s.NoError(goleak.Find(), "background gathering goroutine leaked after context cancellation")
}
