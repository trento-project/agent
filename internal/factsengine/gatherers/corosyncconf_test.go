package gatherers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CorosyncConfTestSuite struct {
	suite.Suite
}

func TestCorosyncConfTestSuite(t *testing.T) {
	suite.Run(t, new(CorosyncConfTestSuite))
}

func (suite *CorosyncConfTestSuite) SetupTest() {
	fileSystem = afero.NewMemMapFs()

	err := fileSystem.MkdirAll("/etc/corosync", 0644)
	if err != nil {
		panic(err)
	}
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfBasic() {
	testFile, _ := os.Open("../../../test/fixtures/gatherers/corosync.conf.basic")
	confFile, _ := ioutil.ReadAll(testFile)
	err := afero.WriteFile(fileSystem, "/etc/corosync/corosync.conf", confFile, 0644)
	assert.NoError(suite.T(), err)
	c := NewCorosyncConfGatherer()

	factRequests := []FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
		{
			Name:     "corosync_join",
			Gatherer: "corosync.conf",
			Argument: "totem.join",
		},
		{
			Name:     "corosync_node1id",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node.0.nodeid",
		},
		{
			Name:     "corosync_node2id",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node.1.nodeid",
		},
		{
			Name:     "corosync_nodes",
			Gatherer: "corosync.conf",
			Argument: "nodelist.node",
		},
		{
			Name:     "corosync_not_found",
			Gatherer: "corosync.conf",
			Argument: "totem.not_found",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:  "corosync_token",
			Value: "30000",
		},
		{
			Name:  "corosync_join",
			Value: "60",
		},
		{
			Name:  "corosync_node1id",
			Value: "1",
		},
		{
			Name:  "corosync_node2id",
			Value: "2",
		},
		{
			Name: "corosync_nodes",
			Value: []interface{}{
				map[string]interface{}{
					"ring0_addr": "10.0.0.119",
					"ring1_addr": "10.0.0.120",
					"nodeid":     "1",
				},
				map[string]interface{}{
					"ring0_addr": "10.0.1.153",
					"ring1_addr": "10.0.1.154",
					"nodeid":     "2",
				},
			},
		},
		{
			Name:  "corosync_not_found",
			Value: nil,
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfFileNotExists() {
	c := NewCorosyncConfGatherer()

	factRequests := []FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	_, err := c.Gather(factRequests)

	suite.EqualError(err, "could not open corosync.conf file: open /etc/corosync/corosync.conf: file does not exist")
}

func (suite *CorosyncConfTestSuite) TestCorosyncConfInvalid() {
	testFile, _ := os.Open("../../../test/fixtures/gatherers/corosync.conf.invalid")
	confFile, _ := ioutil.ReadAll(testFile)
	err := afero.WriteFile(fileSystem, "/etc/corosync/corosync.conf", confFile, 0644)
	assert.NoError(suite.T(), err)
	c := NewCorosyncConfGatherer()

	factRequests := []FactRequest{
		{
			Name:     "corosync_token",
			Gatherer: "corosync.conf",
			Argument: "totem.token",
		},
	}

	_, err = c.Gather(factRequests)

	suite.EqualError(err, "invalid corosync file structure. some section is not closed properly")
}
