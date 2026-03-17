package agent

import "time"

const (
	HeartbeatDefaultInterval = 5 * time.Second
	HeartbeatMinInterval     = 3 * time.Second
)
