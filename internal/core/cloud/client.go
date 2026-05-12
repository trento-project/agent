// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package cloud

import "net/http"

// Extract the client creation for UT purposes

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
