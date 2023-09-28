package gatherers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type FstabGathererTestSuite struct {
	suite.Suite
}

func TestFstabGathererSuite(t *testing.T) {
	suite.Run(t, new(FstabGathererTestSuite))
}

func (s *FstabGathererTestSuite) TestFstabGatheringErrorInvalidFstab() {
	g := gatherers.NewFstabGatherer(helpers.GetFixturePath("gatherers/fstab.invalid"))

	fr := []entities.FactRequest{
		{
			Name:     "fstab",
			CheckID:  "check1",
			Gatherer: "fstab",
		},
	}

	result, err := g.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: fstab-file-error - error reading /etc/fstab file: Syntax error at line 4: ao is not a number")
}

func (s *FstabGathererTestSuite) TestFstabGatheringErrorFstabFileNotFound() {
	g := gatherers.NewFstabGatherer("not found")

	fr := []entities.FactRequest{
		{
			Name:     "fstab",
			CheckID:  "check1",
			Gatherer: "fstab",
		},
	}

	result, err := g.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: fstab-file-error - error reading /etc/fstab file: open not found: no such file or directory")
}

func (s *FstabGathererTestSuite) TestFstabGatheringSuccess() {
	g := gatherers.NewFstabGatherer(helpers.GetFixturePath("gatherers/fstab.valid"))

	fr := []entities.FactRequest{
		{
			Name:     "fstab",
			CheckID:  "check1",
			Gatherer: "fstab",
		},
	}

	expectedResults := []entities.Fact{{
		Name:    "fstab",
		CheckID: "check1",
		Value: &entities.FactValueList{
			Value: []entities.FactValue{
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "/dev/system/root"},
						"mount_point": &entities.FactValueString{Value: "/"},
						"fs":          &entities.FactValueString{Value: "btrfs"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "defaults"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 0},
						"check_order": &entities.FactValueInt{Value: 0},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "/dev/system/root"},
						"mount_point": &entities.FactValueString{Value: "/root"},
						"fs":          &entities.FactValueString{Value: "btrfs"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "subvol=/@/root"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 1},
						"check_order": &entities.FactValueInt{Value: 1},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "/dev/system/root"},
						"mount_point": &entities.FactValueString{Value: "/home"},
						"fs":          &entities.FactValueString{Value: "btrfs"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "subvol=/@/home"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 1},
						"check_order": &entities.FactValueInt{Value: 0},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "/dev/system/swap"},
						"mount_point": &entities.FactValueString{Value: "swap"},
						"fs":          &entities.FactValueString{Value: "swap"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "defaults"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 0},
						"check_order": &entities.FactValueInt{Value: 0},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "/dev/system/root"},
						"mount_point": &entities.FactValueString{Value: "/.snapshots"},
						"fs":          &entities.FactValueString{Value: "btrfs"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "subvol=/@/.snapshots"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 0},
						"check_order": &entities.FactValueInt{Value: 1},
					},
				},
				&entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"device":      &entities.FactValueString{Value: "DA2F-21CF"},
						"mount_point": &entities.FactValueString{Value: "/boot/efi"},
						"fs":          &entities.FactValueString{Value: "vfat"},
						"options": &entities.FactValueList{
							Value: []entities.FactValue{
								&entities.FactValueString{Value: "utf8"},
							},
						},
						"backup":      &entities.FactValueInt{Value: 0},
						"check_order": &entities.FactValueInt{Value: 0},
					},
				},
			},
		},
	}}

	result, err := g.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedResults, result)
}
