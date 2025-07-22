package corosync_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster/corosync"
	"github.com/trento-project/agent/test/helpers"
)

type ParserTestSuite struct {
	suite.Suite
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
func (suite *ParserTestSuite) TestParse() {
	file := helpers.GetFixturePath("discovery/cluster/corosync.conf")

	p := corosync.NewCorosyncParser(file)
	data, err := p.Parse()
	suite.NoError(err)
	suite.Equal("hana_cluster", data.Totem["cluster_name"])
}

func (suite *ParserTestSuite) TestFailingParse() {
	file := helpers.GetFixturePath("NOT_EXISTING_FILE")

	p := corosync.NewCorosyncParser(file)
	_, err := p.Parse()
	suite.Error(err)
}
