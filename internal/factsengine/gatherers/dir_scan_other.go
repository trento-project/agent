// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

//go:build windows

// This file usually won't be compiled, as we solely support Linux builds.
// However, developers may want to build the agent on Windows for testing purposes,
// and this file is needed to avoid build errors when building on Windows.

package gatherers

import (
	"context"
	"fmt"

	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const DirScanGathererName = "dir_scan"

type DirScanGatherer struct{}

func NewDefaultDirScanGatherer() *DirScanGatherer {
	return &DirScanGatherer{}
}

func (d *DirScanGatherer) Gather(_ context.Context, _ []entities.FactRequest) ([]entities.Fact, error) {
	return nil, fmt.Errorf("%s gatherer not supported on Windows", DirScanGathererName)
}
