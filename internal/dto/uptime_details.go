package dto

import "time"

type UptimeDetails struct {
	TotalUptime        time.Duration            `json:"total_uptime"`
	PerContainerUptime map[string]time.Duration `json:"per_container_uptime"`
}