package api

import (
	"time"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

func IsAllowedProfileId(profileID uint) bool {

	if profileID == 0 || profileID == model.MAX_PROFILE_ID {
		return false
	}

	return true
}

func NextHourTimestamp() int64 {
	now := time.Now().UTC()
	// Round up to the next hour
	nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(time.Hour)

	return nextHour.Unix()
}
