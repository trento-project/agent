// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package agent

import "time"

const (
	HeartbeatDefaultInterval = 5 * time.Second
	HeartbeatMinInterval     = 3 * time.Second
)
