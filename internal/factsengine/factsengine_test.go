// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

//nolint:lll
package factsengine_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FactsEngineTestSuite struct {
	suite.Suite
}

func TestFactsEngineTestSuite(t *testing.T) {
	suite.Run(t, new(FactsEngineTestSuite))
}
