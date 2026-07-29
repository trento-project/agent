// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package version

import "github.com/carlmjohnson/versioninfo"

// We explicitly set them via ldflags at build time.
var (
	version            string
	installationSource string
)

// Version returns the version set via ldflags,
// if not set, just the go debug vcs info.
func Version() string {
	if version != "" {
		return version
	}

	return versioninfo.Short()
}

// InstallationSource returns the installation source set via ldflags,
// if not set, "devel".
func InstallationSource() string {
	if installationSource != "" {
		return installationSource
	}

	return "devel"
}
