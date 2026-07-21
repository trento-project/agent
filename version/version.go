// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package version

// We exclude that variables from linting
// because we explicitly use that
// in the ldflags at build time
//
//nolint:gochecknoglobals
var (
	Version            string
	InstallationSource string
)
