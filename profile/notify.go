package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/eapache/go-resiliency/retrier"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

var (
	DefaultNotifyClientOptions = NotifyClientOptions{
		Retries: 1,
	}
)

type NotifyClientOptions struct {
	Retries    uint
	RetryDelay time.Duration
}

type NotifyClient struct {
	api     *model.ApiModel
	backoff *retrier.Retrier
}

func NewNotifyClient(api *model.ApiModel, options NotifyClientOptions) *NotifyClient {
	return &NotifyClient{
		api:     api,
		backoff: retrier.New(retrier.ConstantBackoff(int(options.Retries), options.RetryDelay), nil),
	}
}

func (c *NotifyClient) PublishExtendedProfiles(ctx context.Context, data []*model.ExtendedProfileTierStatusData) error {
	var batch model.PubsubBatch = make(model.PubsubBatch, len(data))

	for _, p := range data {
		val := struct {
			Data *model.ExtendedProfileTierStatusData `json:"data"`
		}{
			Data: p,
		}
		encoded, err := json.Marshal(val)
		if err != nil {
			return errors.Wrapf(err, "Marshal ProfileCache: profile-id=%d", p.ProfileID)
		}
		batch[fmt.Sprint("account@", p.ProfileID)] = encoded
	}

	return c.backoff.Run(func() error {
		err := c.api.PublishBatch(ctx, batch, 0, 0, 0)
		if err != nil {
			return TarantoolError{err}
		}
		return nil
	})
}
