package gatherers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type FSUsageGathererTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.MockCommandExecutor
}

func (s *FSUsageGathererTestSuite) SetupTest() {
	s.mockExecutor = new(utilsMocks.MockCommandExecutor)
}

func TestFSUsageGathererSuite(t *testing.T) {
	suite.Run(t, new(FSUsageGathererTestSuite))
}

func (s *FSUsageGathererTestSuite) TestFstabGatheringSuccess() {
	dfOutputFile := []byte(`Filesystem       1024-blocks      Used Available Capacity Mounted on
/dev/mapper/toot   927310848 346117896 579451320      38% /
`)

	dfOutputAll := []byte(`Filesystem       1024-blocks      Used Available Capacity Mounted on
/dev/mapper/toot   927310848 346119956 579450348      38% /
devtmpfs                4096         0      4096       0% /dev
tmpfs               30672592         0  30672592       0% /dev/shm
efivarfs                 248       117       127      48% /sys/firmware/efi/efivars
tmpfs               12269040      2548  12266492       1% /run
tmpfs                   1024         0      1024       0% /run/credentials/systemd-cryptsetup@root.service
tmpfs                   1024         0      1024       0% /run/credentials/systemd-journald.service
tmpfs               30672592       444  30672148       1% /tmp
/dev/nvme0n1p1       1569776   1233596    336180      79% /usr/local/etc/boot.d with space
tmpfs                   1024         0      1024       0% /run/credentials/getty@tty1.service
tmpfs                6134516        68   6134448       1% /run/user/1000
coolfs auto             1024      1024        -1     101% /run/user/2000
`)

	s.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/df", "-k", "-P", "--", "/usr/sap").Return(dfOutputFile, nil).On("ExecContext", mock.Anything, "/usr/bin/df", "-k", "-P", "--").Return(dfOutputAll, nil)

	gatherer := gatherers.NewFSUsageGatherer(s.mockExecutor)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "specified_file",
			Gatherer: gatherers.FSUsageGathererName,
			CheckID:  "check1",
			Argument: "/usr/sap",
		},
		{
			Name:     "all",
			Gatherer: gatherers.FSUsageGathererName,
			CheckID:  "check1",
		},
	}

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedResults := []entities.Fact{
		{
			Name:    "specified_file",
			CheckID: "check1",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "/dev/mapper/toot"},
							"blocks":     &entities.FactValueInt{Value: 927310848},
							"used":       &entities.FactValueInt{Value: 346117896},
							"available":  &entities.FactValueInt{Value: 579451320},
							"capacity":   &entities.FactValueInt{Value: 38},
							"mountpoint": &entities.FactValueString{Value: "/"},
						},
					},
				},
			},
		},
		{
			Name:    "all",
			CheckID: "check1",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "/dev/mapper/toot"},
							"blocks":     &entities.FactValueInt{Value: 927310848},
							"used":       &entities.FactValueInt{Value: 346119956},
							"available":  &entities.FactValueInt{Value: 579450348},
							"capacity":   &entities.FactValueInt{Value: 38},
							"mountpoint": &entities.FactValueString{Value: "/"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "devtmpfs"},
							"blocks":     &entities.FactValueInt{Value: 4096},
							"used":       &entities.FactValueInt{Value: 0},
							"available":  &entities.FactValueInt{Value: 4096},
							"capacity":   &entities.FactValueInt{Value: 0},
							"mountpoint": &entities.FactValueString{Value: "/dev"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 30672592},
							"used":       &entities.FactValueInt{Value: 0},
							"available":  &entities.FactValueInt{Value: 30672592},
							"capacity":   &entities.FactValueInt{Value: 0},
							"mountpoint": &entities.FactValueString{Value: "/dev/shm"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "efivarfs"},
							"blocks":     &entities.FactValueInt{Value: 248},
							"used":       &entities.FactValueInt{Value: 117},
							"available":  &entities.FactValueInt{Value: 127},
							"capacity":   &entities.FactValueInt{Value: 48},
							"mountpoint": &entities.FactValueString{Value: "/sys/firmware/efi/efivars"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 12269040},
							"used":       &entities.FactValueInt{Value: 2548},
							"available":  &entities.FactValueInt{Value: 12266492},
							"capacity":   &entities.FactValueInt{Value: 1},
							"mountpoint": &entities.FactValueString{Value: "/run"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 1024},
							"used":       &entities.FactValueInt{Value: 0},
							"available":  &entities.FactValueInt{Value: 1024},
							"capacity":   &entities.FactValueInt{Value: 0},
							"mountpoint": &entities.FactValueString{Value: "/run/credentials/systemd-cryptsetup@root.service"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 1024},
							"used":       &entities.FactValueInt{Value: 0},
							"available":  &entities.FactValueInt{Value: 1024},
							"capacity":   &entities.FactValueInt{Value: 0},
							"mountpoint": &entities.FactValueString{Value: "/run/credentials/systemd-journald.service"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 30672592},
							"used":       &entities.FactValueInt{Value: 444},
							"available":  &entities.FactValueInt{Value: 30672148},
							"capacity":   &entities.FactValueInt{Value: 1},
							"mountpoint": &entities.FactValueString{Value: "/tmp"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "/dev/nvme0n1p1"},
							"blocks":     &entities.FactValueInt{Value: 1569776},
							"used":       &entities.FactValueInt{Value: 1233596},
							"available":  &entities.FactValueInt{Value: 336180},
							"capacity":   &entities.FactValueInt{Value: 79},
							"mountpoint": &entities.FactValueString{Value: "/usr/local/etc/boot.d with space"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 1024},
							"used":       &entities.FactValueInt{Value: 0},
							"available":  &entities.FactValueInt{Value: 1024},
							"capacity":   &entities.FactValueInt{Value: 0},
							"mountpoint": &entities.FactValueString{Value: "/run/credentials/getty@tty1.service"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "tmpfs"},
							"blocks":     &entities.FactValueInt{Value: 6134516},
							"used":       &entities.FactValueInt{Value: 68},
							"available":  &entities.FactValueInt{Value: 6134448},
							"capacity":   &entities.FactValueInt{Value: 1},
							"mountpoint": &entities.FactValueString{Value: "/run/user/1000"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"filesystem": &entities.FactValueString{Value: "coolfs auto"},
							"blocks":     &entities.FactValueInt{Value: 1024},
							"used":       &entities.FactValueInt{Value: 1024},
							"available":  &entities.FactValueInt{Value: -1},
							"capacity":   &entities.FactValueInt{Value: 101},
							"mountpoint": &entities.FactValueString{Value: "/run/user/2000"},
						},
					},
				},
			},
		},
	}

	s.NoError(err)
	s.ElementsMatch(expectedResults, factResults)
}

func (s *FSUsageGathererTestSuite) TestFstabGatheringCommandFailedSingle() {
	s.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/df", "-k", "-P", "--", "/usr/sap").Return([]byte{}, errors.New("test error"))

	gatherer := gatherers.NewFSUsageGatherer(s.mockExecutor)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "specified_file",
			Gatherer: gatherers.FSUsageGathererName,
			CheckID:  "check1",
			Argument: "/usr/sap",
		},
	}

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)
	s.NoError(err)

	s.Len(factResults, 1)
	result := factResults[0]

	s.Equal(result.Name, "specified_file")
	s.Equal(result.CheckID, "check1")

	usageErr := new(entities.FactGatheringError)
	s.True(s.ErrorAs(result.Error, &usageErr))
	s.Equal(usageErr.Type, gatherers.FSUsageExecutionError.Type)
}

func (s *FSUsageGathererTestSuite) TestFstabGatheringCommandFailedAll() {
	s.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/df", "-k", "-P", "--").Return([]byte{}, errors.New("test error"))

	gatherer := gatherers.NewFSUsageGatherer(s.mockExecutor)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "all",
			Gatherer: gatherers.FSUsageGathererName,
			CheckID:  "check1",
		},
	}

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)
	s.NoError(err)

	s.Len(factResults, 1)
	result := factResults[0]

	s.Equal(result.Name, "all")
	s.Equal(result.CheckID, "check1")

	usageErr := new(entities.FactGatheringError)
	s.True(s.ErrorAs(result.Error, &usageErr))
	s.Equal(usageErr.Type, gatherers.FSUsageExecutionError.Type)
}

func (s *FSUsageGathererTestSuite) TestFstabGatheringInvalidFormat() {
	dfOutputFile := []byte(`Filesystem       1024-blocks      Used Available Capacity Mounted on
/dev/mapper/toot   927310848 346117896 579451320      38%
`)

	s.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/df", "-k", "-P", "--", "/usr/sap").Return(dfOutputFile, nil)

	gatherer := gatherers.NewFSUsageGatherer(s.mockExecutor)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "single",
			Gatherer: gatherers.FSUsageGathererName,
			CheckID:  "check1",
			Argument: "/usr/sap",
		},
	}

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)
	s.NoError(err)

	s.Len(factResults, 1)
	result := factResults[0]

	s.Equal(result.Name, "single")
	s.Equal(result.CheckID, "check1")

	usageErr := new(entities.FactGatheringError)
	s.True(s.ErrorAs(result.Error, &usageErr))
	s.Equal(usageErr.Type, gatherers.FSUsageInvalidFormatError.Type)
}
