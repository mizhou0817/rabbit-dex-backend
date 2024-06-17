package profile

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/alitto/pond"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

var (
	DefaultMarketsClientOptions = MarketsClientOptions{
		Workers: 100,
		Retries: 1,
	}
)

type MarketId = string

type MarketsClientOptions struct {
	Workers    uint
	Retries    uint
	RetryDelay time.Duration
}

type MarketsClient struct {
	api        *model.ApiModel
	marketsIds []MarketId
	options    MarketsClientOptions
	backoff    *retrier.Retrier
	pool       *pond.WorkerPool
}

func NewMarketsClient(api *model.ApiModel, marketsIds []MarketId, options MarketsClientOptions) *MarketsClient {
	return &MarketsClient{
		api:        api,
		marketsIds: marketsIds,
		options:    options,
		backoff:    retrier.New(retrier.ConstantBackoff(int(options.Retries), options.RetryDelay), nil),
		pool:       pond.New(int(options.Workers), 0),
	}
}

type (
	ProfileMarketsMeta = map[MarketId]*model.ProfileMeta
	ProfilesMeta       = map[ProfileId]ProfileMarketsMeta
	MarketsLastTs      = map[MarketId]int64
)

func (c *MarketsClient) GetUpdatedProfilesMeta(ctx context.Context, marketsLastTs MarketsLastTs) (ProfilesMeta, MarketsLastTs, error) {
	var (
		totalMetas  ProfilesMeta  = make(ProfilesMeta)
		totalLastTs MarketsLastTs = make(MarketsLastTs)
		mu          sync.Mutex
	)

	group, ctx := c.pool.GroupContext(ctx)

	for _, marketId := range c.marketsIds {
		marketId := marketId
		group.Submit(func() error {
			lastTs := marketsLastTs[marketId]

			var marketMetas []*model.ProfileMeta
			err := c.backoff.Run(func() error {
				var err error
				logrus.Infof("Fetching profiles meta for market-id=%s, last-timestamp=%d", marketId, lastTs)
				marketMetas, err = c.api.GetProfilesMetaAfterTs(ctx, marketId, lastTs)
				if err != nil {
					return TarantoolError{errors.Wrapf(err, "GetProfilesMetaAfterTs: market-id=%s, last-ts=%d", marketId, lastTs)}
				}
				return nil
			})
			if err != nil {
				return err
			}

			for _, profileMeta := range marketMetas {
				if lastTs < profileMeta.Timestamp {
					lastTs = profileMeta.Timestamp
				}
				profileId := profileMeta.ProfileID

				mu.Lock()

				if metas, ok := totalMetas[profileMeta.ProfileID]; !ok {
					metas = make(map[MarketId]*model.ProfileMeta, len(c.marketsIds))
					metas[marketId] = profileMeta
					totalMetas[profileId] = metas
				} else {
					if _, ok := metas[marketId]; !ok {
						metas[marketId] = profileMeta
					} else {
						return ErrMetaAlreadyExists
					}
				}

				mu.Unlock()
			}

			{
				mu.Lock()
				totalLastTs[marketId] = lastTs
				mu.Unlock()
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, nil, err
	}

	return totalMetas, totalLastTs, nil
}

func (c *MarketsClient) GetMarketsIds() []MarketId {
	return c.marketsIds
}

func (c *MarketsClient) UpdateProfilesToTiers(ctx context.Context, profilesToTiers []model.ProfileTier) error {
	group, ctx := c.pool.GroupContext(ctx)

	for _, marketId := range c.marketsIds {
		marketId := marketId
		group.Submit(func() error {
			err := c.backoff.Run(func() error {
				logrus.Infof("Updating profiles tiers on market-id=%s", marketId)
				err := c.api.UpdateMarketProfilesToTiers(ctx, marketId, profilesToTiers)
				if err != nil {
					return TarantoolError{errors.Wrapf(err, "UpdateMarketProfilesToTiers: market-id=%s", marketId)}
				}
				return nil
			})
			if err != nil {
				return err
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

var (
	ErrMetaAlreadyExists = errors.New("profile meta already exists")
)
