// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cloud

import "net/http"

// Extract the client creation for UT purposes

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
