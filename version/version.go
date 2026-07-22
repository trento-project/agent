// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package version

// We exclude these variables from linting because
// we explicitly set them via ldflags at build time.
//
//nolint:gochecknoglobals
var (
	Version            string
	InstallationSource string
)
