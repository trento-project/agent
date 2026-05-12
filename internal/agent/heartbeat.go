// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package agent

import "time"

const (
	HeartbeatDefaultInterval = 5 * time.Second
	HeartbeatMinInterval     = 3 * time.Second
)
