package gatherers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/moby/sys/mountinfo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type MountInfoTestSuite struct {
	suite.Suite
	mockMountParser *mocks.MountParserInterface
	mockExecutor    *utilsMocks.CommandExecutor
}

func TestMountInfoTestSuite(t *testing.T) {
	suite.Run(t, new(MountInfoTestSuite))
}

func (suite *MountInfoTestSuite) SetupTest() {
	suite.mockMountParser = new(mocks.MountParserInterface)
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *MountInfoTestSuite) TestMountInfoParsingSuccess() {
	suite.mockMountParser.On("GetMounts", mock.Anything).Return([]*mountinfo.Info{
		{
			Mountpoint: "/sapmnt",
			Source:     "10.1.1.10:/sapmnt",
			FSType:     "nfs4",
			Options:    "rw,relatime",
		},
		{
			Mountpoint: "/hana/data",
			Source:     "/dev/mapper/vg_hana-lv_data",
			FSType:     "xfs",
			Options:    "rw,relatime",
		},
		{
			Mountpoint: "/hana/log",
			Source:     "/dev/mapper/vg_hana-lv_log",
			FSType:     "xfs",
			Options:    "rw,relatime",
		},
	}, nil)

	blkidOutput := []byte(`
DEVNAME=/dev/mapper/vg_hana-lv_data
UUID=82e3685c-6c2a-432b-bf8c-a71286248fae
BLOCK_SIZE=4096
TYPE=xfs
`)

	blkidOutputNoUUID := []byte(`
DEVNAME=/dev/mapper/vg_hana-lv_log
PARTUUID=82e3685c-6c2a-432b-bf8c-a71286248fae
BLOCK_SIZE=4096
TYPE=xfs
`)

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "blkid", "10.1.1.10:/sapmnt", "-o", "export").
		Return(nil, fmt.Errorf("blkid error")).
		On("ExecContext", mock.Anything, "blkid", "/dev/mapper/vg_hana-lv_data", "-o", "export").
		Return(blkidOutput, nil).
		On("ExecContext", mock.Anything, "blkid", "/dev/mapper/vg_hana-lv_log", "-o", "export").
		Return(blkidOutputNoUUID, nil)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "not_mounted",
			Gatherer: "mount_info",
			CheckID:  "check1",
			Argument: "/usr/sap",
		},
		{
			Name:     "shared",
			Gatherer: "mount_info",
			CheckID:  "check1",
			Argument: "/sapmnt",
		},
		{
			Name:     "mounted_locally",
			Gatherer: "mount_info",
			CheckID:  "check1",
			Argument: "/hana/data",
		},
		{
			Name:     "no_uuid",
			Gatherer: "mount_info",
			CheckID:  "check1",
			Argument: "/hana/log",
		},
	}

	gatherer := gatherers.NewMountInfoGatherer(suite.mockMountParser, suite.mockExecutor)

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedResults := []entities.Fact{
		{
			Name:    "not_mounted",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"block_uuid":  &entities.FactValueString{Value: ""},
					"fs_type":     &entities.FactValueString{Value: ""},
					"mount_point": &entities.FactValueString{Value: ""},
					"options":     &entities.FactValueString{Value: ""},
					"source":      &entities.FactValueString{Value: ""},
				},
			},
		},
		{
			Name:    "shared",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"block_uuid":  &entities.FactValueString{Value: ""},
					"fs_type":     &entities.FactValueString{Value: "nfs4"},
					"mount_point": &entities.FactValueString{Value: "/sapmnt"},
					"options":     &entities.FactValueString{Value: "rw,relatime"},
					"source":      &entities.FactValueString{Value: "10.1.1.10:/sapmnt"},
				},
			},
		},
		{
			Name:    "mounted_locally",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"block_uuid":  &entities.FactValueString{Value: "82e3685c-6c2a-432b-bf8c-a71286248fae"},
					"fs_type":     &entities.FactValueString{Value: "xfs"},
					"mount_point": &entities.FactValueString{Value: "/hana/data"},
					"options":     &entities.FactValueString{Value: "rw,relatime"},
					"source":      &entities.FactValueString{Value: "/dev/mapper/vg_hana-lv_data"},
				},
			},
		},
		{
			Name:    "no_uuid",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"block_uuid":  &entities.FactValueString{Value: ""},
					"fs_type":     &entities.FactValueString{Value: "xfs"},
					"mount_point": &entities.FactValueString{Value: "/hana/log"},
					"options":     &entities.FactValueString{Value: "rw,relatime"},
					"source":      &entities.FactValueString{Value: "/dev/mapper/vg_hana-lv_log"},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *MountInfoTestSuite) TestMountInfoParsingNoArgument() {
	suite.mockMountParser.On("GetMounts", mock.Anything).Return([]*mountinfo.Info{}, nil)

	requestedFacts := []entities.FactRequest{
		{
			Name:     "no_argument",
			Gatherer: "mount_info",
			CheckID:  "check1",
		},
	}

	gatherer := gatherers.NewMountInfoGatherer(suite.mockMountParser, suite.mockExecutor)

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)

	expectedResults := []entities.Fact{{
		Name:    "no_argument",
		CheckID: "check1",
		Value:   nil,
		Error: &entities.FactGatheringError{
			Type:    "mount-info-missing-argument",
			Message: "missing required argument",
		},
	}}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *MountInfoTestSuite) TestMountInfoParsingError() {
	suite.mockMountParser.On("GetMounts", mock.Anything).Return(nil, fmt.Errorf("some error"))

	requestedFacts := []entities.FactRequest{
		{
			Name:     "parsing_error",
			Gatherer: "mount_info",
			CheckID:  "check1",
		},
	}

	gatherer := gatherers.NewMountInfoGatherer(suite.mockMountParser, suite.mockExecutor)

	factResults, err := gatherer.Gather(context.Background(), requestedFacts)

	suite.Empty(factResults)
	suite.EqualError(err, "fact gathering error: mount-info-parsing-error - "+
		"error parsing mount information: some error")
}

func (suite *MountInfoTestSuite) TestMountInfoParsingGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := gatherers.NewSapHostCtrlGatherer(utils.Executor{})
	factRequests := []entities.FactRequest{
		{Name: "shared",
			Gatherer: "mount_info",
			CheckID:  "check1",
			Argument: "/sapmnt",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
