package gatherers_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type SapServicesGathererSuite struct {
	suite.Suite
}

func TestSapServicesGathererSuite(t *testing.T) {
	suite.Run(t, new(SapServicesGathererSuite))
}

func (s *SapServicesGathererSuite) TestGatheringFileNotFound() {
	tFs := afero.NewMemMapFs()
	g := gatherers.NewSapServicesGatherer("/usr/sap/sapservices", tFs)

	fr := []entities.FactRequest{
		{
			Name:     "sapservices",
			CheckID:  "check1",
			Gatherer: "sapservices",
		},
	}

	result, err := g.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: sap-services-parsing-error - error reading the sap services file: open /usr/sap/sapservices: file does not exist")
}

func (s *SapServicesGathererSuite) TestGatheringSIDNotIdentifiedSystemd() {
	tFs := afero.NewMemMapFs()
	_ = afero.WriteFile(tFs, "/usr/sap/sapservices", []byte(`
#!/bin/sh
limit.descriptors=1048576
systemctl --no-ask-password start SAPS41_40 
systemctl --no-ask-password start SADS41_41
`), 0777)

	fr := []entities.FactRequest{
		{
			Name:     "sapservices",
			CheckID:  "check1",
			Gatherer: "sapservices",
		},
	}

	g := gatherers.NewSapServicesGatherer("/usr/sap/sapservices", tFs)
	result, err := g.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: sap-services-parsing-error - error parsing the sap services file: could not extract sid from systemd sap services entry: systemctl --no-ask-password start SADS41_41")

}

func (s *SapServicesGathererSuite) TestGatheringSIDNotIdentifiedSapstart() {
	tFs := afero.NewMemMapFs()
	_ = afero.WriteFile(tFs, "/usr/sap/sapservices", []byte(`
#!/bin/sh
limit.descriptors=1048576
LD_LIBRARY_PATH=/usr/sap/HS1/HDB11/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/HS1/HDB11/exe/sapstartsrv pf=/usr/sap/HS1/SYS/profile/HS1HDB11_s41db -D -u hs1adm
LD_LIBRARY_PATH=/usr/sap/S41/ASCS41/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/S41/ASCS41/exe/sapstartsrv pf=/usr/sap/S41/SYS/profile/S41_ASCS41_s41app -D -u s41adm
LD_LIBRARY_PATH=/usr/sap/S41/D40/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/S41/D40/exe/sapstartsrv pf=/usr/sap/S41/SYS/profile/S41_D40_s41app -D -u s41adm
`), 0777)

	fr := []entities.FactRequest{
		{
			Name:     "sapservices",
			CheckID:  "check1",
			Gatherer: "sapservices",
		},
	}

	g := gatherers.NewSapServicesGatherer("/usr/sap/sapservices", tFs)
	result, err := g.Gather(fr)
	s.Nil(result)
	s.EqualError(err, "fact gathering error: sap-services-parsing-error - error parsing the sap services file: could not extract sid from sapstartsrv sap services entry: LD_LIBRARY_PATH=/usr/sap/HS1/HDB11/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/HS1/HDB11/exe/sapstartsrv pf=/usr/sap/HS1/SYS/profile/HS1HDB11_s41db -D -u hs1adm")
}

func (s *SapServicesGathererSuite) TestGatheringSuccessSapstart() {
	tFs := afero.NewMemMapFs()
	_ = afero.WriteFile(tFs, "/usr/sap/sapservices", []byte(`
#!/bin/sh
limit.descriptors=1048576
LD_LIBRARY_PATH=/usr/sap/HS1/HDB11/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/HS1/HDB11/exe/sapstartsrv pf=/usr/sap/HS1/SYS/profile/HS1_HDB11_s41db -D -u hs1adm
LD_LIBRARY_PATH=/usr/sap/S41/ASCS41/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/S41/ASCS41/exe/sapstartsrv pf=/usr/sap/S41/SYS/profile/S41_ASCS41_s41app -D -u s41adm
`), 0777)

	fr := []entities.FactRequest{
		{
			Name:     "sapservices",
			CheckID:  "check1",
			Gatherer: "sapservices",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapservices",
			CheckID: "check1",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"sid":     &entities.FactValueString{Value: "HS1"},
							"kind":    &entities.FactValueString{Value: "sapstartsrv"},
							"content": &entities.FactValueString{Value: "LD_LIBRARY_PATH=/usr/sap/HS1/HDB11/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/HS1/HDB11/exe/sapstartsrv pf=/usr/sap/HS1/SYS/profile/HS1_HDB11_s41db -D -u hs1adm"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"sid":     &entities.FactValueString{Value: "S41"},
							"kind":    &entities.FactValueString{Value: "sapstartsrv"},
							"content": &entities.FactValueString{Value: "LD_LIBRARY_PATH=/usr/sap/S41/ASCS41/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/S41/ASCS41/exe/sapstartsrv pf=/usr/sap/S41/SYS/profile/S41_ASCS41_s41app -D -u s41adm"},
						},
					},
				},
			},
		},
	}
	g := gatherers.NewSapServicesGatherer("/usr/sap/sapservices", tFs)
	result, err := g.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedFacts, result)
}

func (s *SapServicesGathererSuite) TestGatheringSuccessSystemd() {
	tFs := afero.NewMemMapFs()
	_ = afero.WriteFile(tFs, "/usr/sap/sapservices", []byte(`
#!/bin/sh
limit.descriptors=1048576
systemctl --no-ask-password start SAPS41_40
systemctl --no-ask-password start SAPS42_41
`), 0777)

	fr := []entities.FactRequest{
		{
			Name:     "sapservices",
			CheckID:  "check1",
			Gatherer: "sapservices",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapservices",
			CheckID: "check1",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"sid":     &entities.FactValueString{Value: "S41"},
							"kind":    &entities.FactValueString{Value: "systemctl"},
							"content": &entities.FactValueString{Value: "systemctl --no-ask-password start SAPS41_40"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"sid":     &entities.FactValueString{Value: "S42"},
							"kind":    &entities.FactValueString{Value: "systemctl"},
							"content": &entities.FactValueString{Value: "systemctl --no-ask-password start SAPS42_41"},
						},
					},
				},
			},
		},
	}
	g := gatherers.NewSapServicesGatherer("/usr/sap/sapservices", tFs)
	result, err := g.Gather(fr)
	s.NoError(err)
	s.EqualValues(expectedFacts, result)
}
