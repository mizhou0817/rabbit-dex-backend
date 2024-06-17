package referrals

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLevel0VolumeLevel(t *testing.T) {
	volume := newDecimal(0)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[0], res.Current)
	assert.Equal(t, REFERRAL_LEVELS[1], *res.Next)
	assert.Equal(t, REFERRAL_LEVELS[1].Volume.Sub(volume), *res.NeededVolume)
}

func TestGetLevel1VolumeLevel(t *testing.T) {
	volume := newDecimal(1)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[0], res.Current)
	assert.Equal(t, REFERRAL_LEVELS[1], *res.Next)
	assert.Equal(t, REFERRAL_LEVELS[1].Volume.Sub(volume), *res.NeededVolume)
}

func TestGetLevelMidLevel(t *testing.T) {
	volume := newDecimal(400_000)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[2], res.Current)
	assert.Equal(t, REFERRAL_LEVELS[3], *res.Next)
	assert.Equal(t, REFERRAL_LEVELS[3].Volume.Sub(volume), *res.NeededVolume)
}

func TestGetLevelMaxLevel(t *testing.T) {
	volume := newDecimal(10_000_000_000)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[11], res.Current)
	assert.Nil(t, res.Next)
	assert.Nil(t, res.NeededVolume)
}

func TestGetLevelMaxPlusLevel(t *testing.T) {
	volume := newDecimal(20000000000001)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[11], res.Current)
	assert.Nil(t, res.Next)
	assert.Nil(t, res.NeededVolume)
}

func TestGetLevelMidJSON(t *testing.T) {
	volume := newDecimal(200_001)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[0], *res.Prev)
	assert.Equal(t, REFERRAL_LEVELS[1], res.Current)
	assert.Equal(t, REFERRAL_LEVELS[2], *res.Next)
	assert.Equal(t, REFERRAL_LEVELS[2].Volume.Sub(volume), *res.NeededVolume)

	jsonStr, _ := json.Marshal(res)
	expectedStr := `
		{
			"volume":"200001",
			"needed_volume":"199999",
			"current":{"volume":"200000","commission_percent":"0.2", "milestone_bonus":"5","level":2},
			"next":{"volume":"400000","commission_percent":"0.22","milestone_bonus":"10","level":3},
			"prev":{"volume":"0","commission_percent":"0.2","milestone_bonus":"0","level":1}
		}`
	assert.JSONEq(t, expectedStr, string(jsonStr))
}

func TestGetLevelLastJSON(t *testing.T) {
	volume := newDecimal(10_000_000_000)
	res := GetLevel(volume)

	assert.Equal(t, REFERRAL_LEVELS[10], *res.Prev)
	assert.Equal(t, REFERRAL_LEVELS[11], res.Current)
	assert.Nil(t, res.Next)
	assert.Nil(t, res.NeededVolume)

	jsonStr, _ := json.Marshal(res)

	expectedStr := `
		{
			"volume":"10000000000",
			"current":{"volume":"10000000000","commission_percent":"0.4", "milestone_bonus":"100000","level":12},
			"prev":{"volume":"1000000000","commission_percent":"0.38","milestone_bonus":"25000","level":11}
		}`

	assert.JSONEq(t, expectedStr, string(jsonStr))
}
