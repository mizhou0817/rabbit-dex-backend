package model

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	PROFILE_CREATE_AIRDROP           = "airdrop.create_airdrop"
	PROFILE_SET_PROFILE_TOTAL        = "airdrop.set_profile_total"
	PROFILE_UPDATE_PROFILE_CLAIMABLE = "airdrop.update_profile_claimable"
	PROFILE_CLAIM_ALL                = "airdrop.claim_all"
	GET_PROFILE_AIRDROPS             = "airdrop.get_profile_airdrops"
	UPDATE_ALL_PROFILE_AIRDROPS      = "airdrop.update_all_profile_airdrops"
	PENDING_CLAIM                    = "airdrop.pending_claim"
	FINISH_CLAIM                     = "airdrop.finish_claim"
	DELETE_ALL_AIRDROPS              = "airdrop.delete_all_airdrops"
	TEST_CREATE_CLAIM_OPS            = "airdrop.test_create_claim_ops"
	TEST_GET_CLAIM_OPS               = "airdrop.test_get_claim_ops"
)

type Airdrop struct {
	Title          string `msgpack:"title" json:"last_processed_block"`
	StartTimestamp int64  `msgpack:"start_timestamp" json:"start_timestamp"`
	EndTimestamp   int64  `msgpack:"end_timestamp" json:"end_timestamp"`
	ShardId        string `msgpack:"shard_id" json:"-"`
	ArchiveId      int    `msgpack:"archive_id" json:"-"`
}

type ProfileAirdrop struct {
	ProfileId               uint             `msgpack:"profile_id" json:"profile_id"`
	AirdropTitle            string           `msgpack:"airdrop_title" json:"airdrop_title"`
	Status                  string           `msgpack:"status" json:"status"`
	TotalVolumeForAirdrop   tdecimal.Decimal `msgpack:"total_volume_for_airdrop" json:"total_volume_for_airdrop"`
	TotalVolumeAfterAirdrop tdecimal.Decimal `msgpack:"total_volume_after_airdrop" json:"total_volume_after_airdrop"`
	TotalRewards            tdecimal.Decimal `msgpack:"total_rewards" json:"total_rewards"`
	Claimable               tdecimal.Decimal `msgpack:"claimable" json:"claimable"`
	Claimed                 tdecimal.Decimal `msgpack:"claimed" json:"claimed"`
	LastFillTimestamp       map[string]int64 `msgpack:"last_fill_timestamp" json:"last_fill_timestamp"`
	InitialRewards          tdecimal.Decimal `msgpack:"initial_rewards" json:"-"`
	ShardId                 string           `msgpack:"shard_id" json:"-"`
	ArchiveId               int              `msgpack:"archive_id" json:"-"`
}

type AirdropClaimOps struct {
	Id           uint             `msgpack:"id" json:"id"`
	AirdropTitle string           `msgpack:"airdrop_title" json:"airdrop_title"`
	ProfileId    uint             `msgpack:"profile_id" json:"profile_id"`
	Status       string           `msgpack:"status" json:"status"`
	Amount       tdecimal.Decimal `msgpack:"amount" json:"amount"`
	Timestamp    int64            `msgpack:"timestamp" json:"timestamp"`
	ShardId      string           `msgpack:"shard_id" json:"-"`
	ArchiveId    int              `msgpack:"archive_id" json:"-"`
}

func (api *ApiModel) CreateAirdrop(ctx context.Context, title string, start_timestamp, end_timestamp int64) (*Airdrop, error) {
	return DataResponse[*Airdrop]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_CREATE_AIRDROP, []interface{}{
		title,
		start_timestamp,
		end_timestamp,
	})
}

func (api *ApiModel) GetProfileAirdrops(ctx context.Context, profile_id uint) ([]*ProfileAirdrop, error) {
	return DataResponse[[]*ProfileAirdrop]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_PROFILE_AIRDROPS, []interface{}{
		profile_id,
	})
}

func (api *ApiModel) SetProfileTotal(ctx context.Context, profile_id uint, airdrop_title string, total_rewards, claimable float64) (*ProfileAirdrop, error) {

	// Will fix it later
	d_total_rewards := tdecimal.NewDecimal(decimal.NewFromFloat(total_rewards))
	d_claimable := tdecimal.NewDecimal(decimal.NewFromFloat(claimable))

	return DataResponse[*ProfileAirdrop]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_SET_PROFILE_TOTAL, []interface{}{
		profile_id,
		airdrop_title,
		d_total_rewards,
		d_claimable,
	})
}

func (api *ApiModel) UpdateProfileClaimable(ctx context.Context, profile_id uint, airdrop_title string) (*ProfileAirdrop, error) {
	return DataResponse[*ProfileAirdrop]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_UPDATE_PROFILE_CLAIMABLE, []interface{}{
		profile_id,
		airdrop_title,
	})
}

func (api *ApiModel) ProfileClaimAll(ctx context.Context, profile_id uint, airdrop_title string) (*AirdropClaimOps, error) {
	return DataResponse[*AirdropClaimOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_CLAIM_ALL, []interface{}{
		profile_id,
		airdrop_title,
	})
}

func (api *ApiModel) PendingClaimOps(ctx context.Context, profile_id uint) (*AirdropClaimOps, error) {
	return DataResponse[*AirdropClaimOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PENDING_CLAIM, []interface{}{
		profile_id,
	})
}

func (api *ApiModel) FinishClaim(ctx context.Context, profile_id uint) (*AirdropClaimOps, error) {
	return DataResponse[*AirdropClaimOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, FINISH_CLAIM, []interface{}{
		profile_id,
	})
}

func (api *ApiModel) UpdateAllProfileAirdrops(ctx context.Context, profile_id uint) error {
	_, err := DataResponse[*interface{}]{}.Request(ctx, PROFILE_INSTANCE, api.broker, UPDATE_ALL_PROFILE_AIRDROPS, []interface{}{
		profile_id,
	})

	return err
}

func (api *ApiModel) TestAirdropResetAll(ctx context.Context, title string, start_timestamp, end_timestamp int64, profile_id uint, total_rewards, claimable float64) error {

	d_total_rewards := tdecimal.NewDecimal(decimal.NewFromFloat(total_rewards))
	d_claimable := tdecimal.NewDecimal(decimal.NewFromFloat(claimable))

	_, err := DataResponse[*interface{}]{}.Request(ctx, PROFILE_INSTANCE, api.broker, DELETE_ALL_AIRDROPS, []interface{}{
		title,
		start_timestamp,
		end_timestamp,
		profile_id,
		d_total_rewards,
		d_claimable,
	})

	return err
}

// test_create_claim_ops(profile_id, airdrop_title, claimable
func (api *ApiModel) TestCreateClaimOps(ctx context.Context, profile_id uint, airdrop_title string, claimable float64) (*AirdropClaimOps, error) {
	return DataResponse[*AirdropClaimOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, TEST_CREATE_CLAIM_OPS, []interface{}{
		profile_id,
		airdrop_title,
		claimable,
	})
}

func (api *ApiModel) TestGetClaimOps(ctx context.Context, ops_id uint) (*AirdropClaimOps, error) {
	return DataResponse[*AirdropClaimOps]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, TEST_GET_CLAIM_OPS, []interface{}{
		ops_id,
	})
}
