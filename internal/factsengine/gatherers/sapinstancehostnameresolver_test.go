package gatherers_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type SapInstanceHostnameResolverTestSuite struct {
	suite.Suite
	mockResolver *mocks.HostnameResolver
	mockPinger   *mocks.HostPinger
}

func TestSapInstanceHostnameResolverTestSuite(t *testing.T) {
	suite.Run(t, new(SapInstanceHostnameResolverTestSuite))
}

func (suite *SapInstanceHostnameResolverTestSuite) SetupTest() {
	suite.mockResolver = new(mocks.HostnameResolver)
	suite.mockPinger = new(mocks.HostPinger)
}

func (suite *SapInstanceHostnameResolverTestSuite) TestSapInstanceHostnameResolverSuccess() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/NWP", 0644)
	suite.NoError(err)

	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ERS10_sapqaser", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/NWP/profile/NWP_ERS10_sapnwper", []byte{}, 0644)
	suite.NoError(err)

	suite.mockResolver.On("LookupHost", "sapqasas").Return([]string{"10.1.1.5"}, nil)
	suite.mockPinger.On("Ping", "sapqasas").Return(true, nil)
	suite.mockResolver.On("LookupHost", "sapqaser").Return([]string{"10.1.1.6"}, nil)
	suite.mockPinger.On("Ping", "sapqaser").Return(true, nil)
	suite.mockResolver.On("LookupHost", "sapnwper").Return([]string{"10.1.1.7"}, nil)
	suite.mockPinger.On("Ping", "sapnwper").Return(false, nil)

	g := gatherers.NewSapInstanceHostnameResolverGatherer(appFS, suite.mockResolver, suite.mockPinger)

	factRequests := []entities.FactRequest{{
		Name:     "sapinstance_hostname_resolver",
		Gatherer: "sapinstance_hostname_resolver",
		CheckID:  "check1",
	}}

	factResults, err := g.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "sapinstance_hostname_resolver",
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
									"reachability":  &entities.FactValueBool{Value: true},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"hostname": &entities.FactValueString{Value: "sapqaser"},
									"addresses": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueString{Value: "10.1.1.6"},
										},
									},
									"instance_name": &entities.FactValueString{Value: "ERS10"},
									"reachability":  &entities.FactValueBool{Value: true},
								},
							},
						},
					},
					"NWP": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"hostname": &entities.FactValueString{Value: "sapnwper"},
									"addresses": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueString{Value: "10.1.1.7"},
										},
									},
									"instance_name": &entities.FactValueString{Value: "ERS10"},
									"reachability":  &entities.FactValueBool{Value: false},
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

func (suite *SapInstanceHostnameResolverTestSuite) TestSapInstanceHostnameResolverNoProfiles() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	g := gatherers.NewSapInstanceHostnameResolverGatherer(appFS, suite.mockResolver, suite.mockPinger)

	factRequests := []entities.FactRequest{{
		Name:     "sapinstance_hostname_resolver",
		Gatherer: "sapinstance_hostname_resolver",
		CheckID:  "check1",
	}}

	factResults, err := g.Gather(factRequests)
	suite.Nil(factResults)
	suite.EqualError(err, "fact gathering error: sapinstance-hostname-resolver-details-error - error gathering details: open /sapmnt/QAS/profile: file does not exist")
}

func (suite *SapInstanceHostnameResolverTestSuite) TestSapInstanceHostnameResolverLookupHostError() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", []byte{}, 0644)
	suite.NoError(err)

	suite.mockResolver.On("LookupHost", "sapqasas").Return([]string{}, errors.New("lookup sapqasas on 169.254.169.254:53: dial udp 169.254.169.254:53: connect: no route to host"))
	suite.mockPinger.On("Ping", "sapqasas").Return(false, nil)

	g := gatherers.NewSapInstanceHostnameResolverGatherer(appFS, suite.mockResolver, suite.mockPinger)

	factRequests := []entities.FactRequest{{
		Name:     "sapinstance_hostname_resolver",
		Gatherer: "sapinstance_hostname_resolver",
		CheckID:  "check1",
	}}

	expectedResults := []entities.Fact{
		{
			Name:    "sapinstance_hostname_resolver",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"QAS": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"hostname": &entities.FactValueString{Value: "sapqasas"},
									"addresses": &entities.FactValueList{
										Value: []entities.FactValue{},
									},
									"instance_name": &entities.FactValueString{Value: "ASCS00"},
									"reachability":  &entities.FactValueBool{Value: false},
								},
							},
						},
					},
				},
			},
		},
	}

	factResults, err := g.Gather(factRequests)

	suite.NoError(err)
	suite.Equal(expectedResults, factResults)
}
