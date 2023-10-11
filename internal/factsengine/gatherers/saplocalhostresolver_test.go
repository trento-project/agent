package gatherers_test

import (
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SapLocalhostResolverTestSuite struct {
	suite.Suite
	mockResolver *utilsMocks.HostnameResolver
}

func TestSapLocalhostResolverTestSuite(t *testing.T) {
	suite.Run(t, new(SapLocalhostResolverTestSuite))
}

func (suite *SapLocalhostResolverTestSuite) SetupTest() {
	suite.mockResolver = new(utilsMocks.HostnameResolver)
}

func (suite *SapLocalhostResolverTestSuite) TestSapLocalhostResolverSuccess() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/PRD", 0644)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	defaultProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.default"))
	defaultProfileContent, _ := io.ReadAll(defaultProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/DEFAULT.PFL", defaultProfileContent, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/DEFAULT.1.PFL", []byte{}, 0644)
	suite.NoError(err)

	ascsProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.ascs"))
	ascsProfileConcent, _ := io.ReadAll(ascsProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", ascsProfileConcent, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas.1", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas.bak", []byte{}, 0644)
	suite.NoError(err)

	minimalProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.minimal"))
	minimalProfileContent, _ := io.ReadAll(minimalProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/DEFAULT.PFL", minimalProfileContent, 0644)
	suite.NoError(err)

	suite.mockResolver.On("LookupHost", "sapnwpas").Return([]string{"1.2.3.4"}, nil)
	suite.mockResolver.On("LookupHost", "sapqasas").Return([]string{"10.1.1.5"}, nil)

	g := gatherers.NewSapLocalhostResolver(appFS, suite.mockResolver)

	factRequests := []entities.FactRequest{{
		Name:     "sap_localhost_resolver",
		Gatherer: "sap_localhost_resolver",
		CheckID:  "check1",
	}}

	factResults, err := g.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "sap_localhost_resolver",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"QAS": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"hostname": &entities.FactValueString{Value: "sapqasas"},
							"addresses": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueString{Value: "10.1.1.5"},
								},
							},
							"instance_name": &entities.FactValueString{Value: "ASCS00"},
						},
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedResults, factResults)
}
