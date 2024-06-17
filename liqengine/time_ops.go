package liqengine

import (
	"time"
)

const (
	INSURANCE_WATERFALL_INTERVAL = time.Duration(6 * time.Second)
	WATERFALL1_INTERVAL          = time.Duration(6 * time.Second) //service runs every 6 seconds and times are to nearest second
)

func IsIntervalPassedForMicroseconds(last_check int64, interval time.Duration) bool {
	t_now := time.Now().UnixMicro()
	passed := time.Duration(t_now-last_check) * 1e3

	diff := passed - interval
	return diff >= 0
}
