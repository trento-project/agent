package gatherers_test

import (
	"errors"
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

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	ascsProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.ascs"))
	ascsProfileConcent, _ := io.ReadAll(ascsProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", ascsProfileConcent, 0644)
	suite.NoError(err)

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
					"QAS": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
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
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedResults, factResults)
}

func (suite *SapLocalhostResolverTestSuite) TestSapLocalhostResolverNoProfiles() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	g := gatherers.NewSapLocalhostResolver(appFS, suite.mockResolver)

	factRequests := []entities.FactRequest{{
		Name:     "sap_localhost_resolver",
		Gatherer: "sap_localhost_resolver",
		CheckID:  "check1",
	}}

	factResults, err := g.Gather(factRequests)
	suite.Nil(factResults)
	suite.EqualError(err, "fact gathering error: saplocalhost_resolver-file-system-error - error reading the sap profiles file system: open /sapmnt/QAS/profile: file does not exist")
}

func (suite *SapLocalhostResolverTestSuite) TestSapLocalhostResolverLookupHostError() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	ascsProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.ascs"))
	ascsProfileConcent, _ := io.ReadAll(ascsProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", ascsProfileConcent, 0644)
	suite.NoError(err)

	suite.mockResolver.On("LookupHost", "sapqasas").Return([]string{}, errors.New("lookup sapqasas on 169.254.169.254:53: dial udp 169.254.169.254:53: connect: no route to host"))

	g := gatherers.NewSapLocalhostResolver(appFS, suite.mockResolver)

	factRequests := []entities.FactRequest{{
		Name:     "sap_localhost_resolver",
		Gatherer: "sap_localhost_resolver",
		CheckID:  "check1",
	}}

	factResults, err := g.Gather(factRequests)

	suite.Nil(factResults)
	suite.EqualError(err, "fact gathering error: saplocalhost_resolver-resolution-error - error resolving hostname: lookup sapqasas on 169.254.169.254:53: dial udp 169.254.169.254:53: connect: no route to host")
}
