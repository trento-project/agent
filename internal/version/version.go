// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package version

// We explicitly set them via ldflags at build time.
var (
	version            string
	installationSource string
)

func Version() string {
	return version
}

func InstallationSource() string {
	return installationSource
}
