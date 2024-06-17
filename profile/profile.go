package profile

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/eapache/go-resiliency/retrier"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile/types"
)

var (
	DefaultProfileClientOptions = ProfileClientOptions{
		Retries: 1,
	}
)

type ProfileId = types.ProfileId

type ProfileClientOptions struct {
	Retries    uint
	RetryDelay time.Duration
}

type ProfileClient struct {
	api     *model.ApiModel
	options ProfileClientOptions
	backoff *retrier.Retrier
}

func NewProfileClient(api *model.ApiModel, options ProfileClientOptions) *ProfileClient {
	return &ProfileClient{
		api:     api,
		backoff: retrier.New(retrier.ConstantBackoff(int(options.Retries), options.RetryDelay), nil),
	}
}

func (c *ProfileClient) GetExtendedProfiles(ctx context.Context, profilesIds ...ProfileId) ([]*model.ExtendedProfile, error) {
	var profiles []*model.ExtendedProfile

	err := c.backoff.Run(func() error {
		var err error
		profiles, err = c.api.GetExtendedProfiles(ctx, profilesIds...)
		if err != nil {
			return TarantoolError{err}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func (c *ProfileClient) UpdateProfilesCachesMetas(ctx context.Context, data []*model.ProfileCacheMetas) error {
	err := c.backoff.Run(func() error {
		err := c.api.UpdateProfilesCachesAndMetas(ctx, data)
		if err != nil {
			return TarantoolError{errors.Wrap(err, "UpdateProfilesCachesAndMetas")}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
